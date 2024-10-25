// Copyright 2022 The Erigon Authors
// This file is part of Erigon.
//
// Erigon is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Erigon is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Erigon. If not, see <http://www.gnu.org/licenses/>.

package commitment

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"math/bits"
	"path/filepath"
	"runtime"
	"time"

	"github.com/erigontech/erigon-lib/etl"
	"github.com/erigontech/erigon-lib/log/v3"
	"github.com/erigontech/erigon/core/types/accounts"
	witnesstypes "github.com/erigontech/erigon/core/types/witness"
	"github.com/erigontech/erigon/turbo/trie"

	"github.com/erigontech/erigon-lib/common/dbg"

	"github.com/erigontech/erigon-lib/common"
	libcommon "github.com/erigontech/erigon-lib/common"
	"github.com/erigontech/erigon-lib/common/hexutility"
	"github.com/erigontech/erigon-lib/common/length"
	"github.com/erigontech/erigon-lib/rlp"
	"golang.org/x/crypto/sha3"
)

// keccakState wraps sha3.state. In addition to the usual hash methods, it also supports
// Read to get a variable amount of data from the hash state. Read is faster than Sum
// because it doesn't copy the internal state, but also modifies the internal state.
type keccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}

// HexPatriciaHashed implements commitment based on patricia merkle tree with radix 16,
// with keys pre-hashed by keccak256
type HexPatriciaHashed struct {
	root cell // Root cell of the tree
	// How many rows (starting from row 0) are currently active and have corresponding selected columns
	// Last active row does not have selected column
	activeRows int
	// Length of the key that reflects current positioning of the grid. It maybe larger than number of active rows,
	// if an account leaf cell represents multiple nibbles in the key
	currentKeyLen int
	accountKeyLen int
	// Rows of the grid correspond to the level of depth in the patricia tree
	// Columns of the grid correspond to pointers to the nodes further from the root
	grid          [128][16]cell // First 64 rows of this grid are for account trie, and next 64 rows are for storage trie
	currentKey    [128]byte     // For each row indicates which column is currently selected
	depths        [128]int      // For each row, the depth of cells in that row
	branchBefore  [128]bool     // For each row, whether there was a branch node in the database loaded in unfold
	touchMap      [128]uint16   // For each row, bitmap of cells that were either present before modification, or modified or deleted
	afterMap      [128]uint16   // For each row, bitmap of cells that were present after modification
	keccak        keccakState
	keccak2       keccakState
	rootChecked   bool // Set to false if it is not known whether the root is empty, set to true if it is checked
	rootTouched   bool
	rootPresent   bool
	trace         bool
	Ctx           PatriciaContext
	hashAuxBuffer [128]byte     // buffer to compute cell hash or write hash-related things
	auxBuffer     *bytes.Buffer // auxiliary buffer used during branch updates encoding
	branchEncoder *BranchEncoder
}

func NewHexPatriciaHashed(accountKeyLen int, ctx PatriciaContext, tmpdir string) *HexPatriciaHashed {
	hph := &HexPatriciaHashed{
		Ctx:           ctx,
		keccak:        sha3.NewLegacyKeccak256().(keccakState),
		keccak2:       sha3.NewLegacyKeccak256().(keccakState),
		accountKeyLen: accountKeyLen,
		auxBuffer:     bytes.NewBuffer(make([]byte, 8192)),
	}
	hph.branchEncoder = NewBranchEncoder(1024, filepath.Join(tmpdir, "branch-encoder"))
	return hph
}

type cell struct {
	hashedExtension [128]byte
	extension       [64]byte
	accountAddr     [length.Addr]byte               // account plain key
	storageAddr     [length.Addr + length.Hash]byte // storage plain key
	hash            [length.Hash]byte               // cell hash
	hashLen         int                             // Length of the hash (or embedded)
	accountAddrLen  int                             // length of account plain key
	storageAddrLen  int                             // length of the storage plain key
	hashedExtLen    int                             // length of the hashed extension, if any
	extLen          int                             // length of the extension, if any
	Update                                          // state update
}

var (
	EmptyRootHash      = hexutility.MustDecodeHex("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
	EmptyCodeHash      = hexutility.MustDecodeHex("c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")
	EmptyCodeHashArray = *(*[length.Hash]byte)(EmptyCodeHash)
)

func (cell *cell) reset() {
	cell.accountAddrLen = 0
	cell.storageAddrLen = 0
	cell.hashedExtLen = 0
	cell.extLen = 0
	cell.hashLen = 0
	clear(cell.hashedExtension[:])
	clear(cell.extension[:])
	clear(cell.accountAddr[:])
	clear(cell.storageAddr[:])
	clear(cell.hash[:])
	cell.Update.Reset()
}

func (cell *cell) setFromUpdate(update *Update) { cell.Update.Merge(update) }

func (cell *cell) fillFromUpperCell(upCell *cell, depth, depthIncrement int) {
	if upCell.hashedExtLen >= depthIncrement {
		cell.hashedExtLen = upCell.hashedExtLen - depthIncrement
	} else {
		cell.hashedExtLen = 0
	}
	if upCell.hashedExtLen > depthIncrement {
		copy(cell.hashedExtension[:], upCell.hashedExtension[depthIncrement:upCell.hashedExtLen])
	}
	if upCell.extLen >= depthIncrement {
		cell.extLen = upCell.extLen - depthIncrement
	} else {
		cell.extLen = 0
	}
	if upCell.extLen > depthIncrement {
		copy(cell.extension[:], upCell.extension[depthIncrement:upCell.extLen])
	}
	if depth <= 64 {
		cell.accountAddrLen = upCell.accountAddrLen
		if upCell.accountAddrLen > 0 {
			copy(cell.accountAddr[:], upCell.accountAddr[:cell.accountAddrLen])
			cell.Balance.Set(&upCell.Balance)
			cell.Nonce = upCell.Nonce
			copy(cell.CodeHash[:], upCell.CodeHash[:])
			cell.extLen = upCell.extLen
			if upCell.extLen > 0 {
				copy(cell.extension[:], upCell.extension[:upCell.extLen])
			}
		}
	} else {
		cell.accountAddrLen = 0
	}
	cell.storageAddrLen = upCell.storageAddrLen
	if upCell.storageAddrLen > 0 {
		copy(cell.storageAddr[:], upCell.storageAddr[:upCell.storageAddrLen])
		cell.StorageLen = upCell.StorageLen
		if upCell.StorageLen > 0 {
			copy(cell.Storage[:], upCell.Storage[:upCell.StorageLen])
		}
	}
	cell.hashLen = upCell.hashLen
	if upCell.hashLen > 0 {
		copy(cell.hash[:], upCell.hash[:upCell.hashLen])
	}
}

func (cell *cell) fillFromLowerCell(lowCell *cell, lowDepth int, preExtension []byte, nibble int) {
	if lowCell.accountAddrLen > 0 || lowDepth < 64 {
		cell.accountAddrLen = lowCell.accountAddrLen
	}
	if lowCell.accountAddrLen > 0 {
		copy(cell.accountAddr[:], lowCell.accountAddr[:cell.accountAddrLen])
		cell.Balance.Set(&lowCell.Balance)
		cell.Nonce = lowCell.Nonce
		copy(cell.CodeHash[:], lowCell.CodeHash[:])
	}
	cell.storageAddrLen = lowCell.storageAddrLen
	if lowCell.storageAddrLen > 0 {
		copy(cell.storageAddr[:], lowCell.storageAddr[:cell.storageAddrLen])
		cell.StorageLen = lowCell.StorageLen
		if lowCell.StorageLen > 0 {
			copy(cell.Storage[:], lowCell.Storage[:lowCell.StorageLen])
		}
	}
	if lowCell.hashLen > 0 {
		if (lowCell.accountAddrLen == 0 && lowDepth < 64) || (lowCell.storageAddrLen == 0 && lowDepth > 64) {
			// Extension is related to either accounts branch node, or storage branch node, we prepend it by preExtension | nibble
			if len(preExtension) > 0 {
				copy(cell.extension[:], preExtension)
			}
			cell.extension[len(preExtension)] = byte(nibble)
			if lowCell.extLen > 0 {
				copy(cell.extension[1+len(preExtension):], lowCell.extension[:lowCell.extLen])
			}
			cell.extLen = lowCell.extLen + 1 + len(preExtension)
		} else {
			// Extension is related to a storage branch node, so we copy it upwards as is
			cell.extLen = lowCell.extLen
			if lowCell.extLen > 0 {
				copy(cell.extension[:], lowCell.extension[:lowCell.extLen])
			}
		}
	}
	cell.hashLen = lowCell.hashLen
	if lowCell.hashLen > 0 {
		copy(cell.hash[:], lowCell.hash[:lowCell.hashLen])
	}
}

func hashKey(keccak keccakState, plainKey []byte, dest []byte, hashedKeyOffset int) error {
	keccak.Reset()
	var hashBufBack [length.Hash]byte
	hashBuf := hashBufBack[:]
	if _, err := keccak.Write(plainKey); err != nil {
		return err
	}
	if _, err := keccak.Read(hashBuf); err != nil {
		return err
	}
	hashBuf = hashBuf[hashedKeyOffset/2:]
	var k int
	if hashedKeyOffset%2 == 1 {
		dest[0] = hashBuf[0] & 0xf
		k++
		hashBuf = hashBuf[1:]
	}
	for _, c := range hashBuf {
		dest[k] = (c >> 4) & 0xf
		k++
		dest[k] = c & 0xf
		k++
	}
	return nil
}

func (cell *cell) deriveHashedKeys(depth int, keccak keccakState, accountKeyLen int) error {
	extraLen := 0
	if cell.accountAddrLen > 0 {
		if depth > 64 {
			return errors.New("deriveHashedKeys accountAddr present at depth > 64")
		}
		extraLen = 64 - depth
	}
	if cell.storageAddrLen > 0 {
		if depth >= 64 {
			extraLen = 128 - depth
		} else {
			extraLen += 64
		}
	}
	if extraLen > 0 {
		if cell.hashedExtLen > 0 {
			copy(cell.hashedExtension[extraLen:], cell.hashedExtension[:cell.hashedExtLen])
		}
		cell.hashedExtLen = min(extraLen+cell.hashedExtLen, len(cell.hashedExtension))
		var hashedKeyOffset, downOffset int
		if cell.accountAddrLen > 0 {
			if err := hashKey(keccak, cell.accountAddr[:cell.accountAddrLen], cell.hashedExtension[:], depth); err != nil {
				return err
			}
			downOffset = 64 - depth
		}
		if cell.storageAddrLen > 0 {
			if depth >= 64 {
				hashedKeyOffset = depth - 64
			}
			if depth == 0 {
				accountKeyLen = 0
			}
			if err := hashKey(keccak, cell.storageAddr[accountKeyLen:cell.storageAddrLen], cell.hashedExtension[downOffset:], hashedKeyOffset); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cell *cell) fillFromFields(data []byte, pos int, fieldBits cellFields) (int, error) {
	if fieldBits&fieldExtension != 0 {
		l, n := binary.Uvarint(data[pos:])
		if n == 0 {
			return 0, errors.New("fillFromFields buffer too small for hashedKey len")
		} else if n < 0 {
			return 0, errors.New("fillFromFields value overflow for hashedKey len")
		}
		pos += n
		if len(data) < pos+int(l) {
			return 0, fmt.Errorf("fillFromFields buffer too small for hashedKey exp %d got %d", pos+int(l), len(data))
		}
		cell.hashedExtLen = int(l)
		cell.extLen = int(l)
		if l > 0 {
			copy(cell.hashedExtension[:], data[pos:pos+int(l)])
			copy(cell.extension[:], data[pos:pos+int(l)])
			pos += int(l)
		}
	} else {
		cell.hashedExtLen = 0
		cell.extLen = 0
	}
	if fieldBits&fieldAccountAddr != 0 {
		l, n := binary.Uvarint(data[pos:])
		if n == 0 {
			return 0, errors.New("fillFromFields buffer too small for accountAddr len")
		} else if n < 0 {
			return 0, errors.New("fillFromFields value overflow for accountAddr len")
		}
		pos += n
		if len(data) < pos+int(l) {
			return 0, errors.New("fillFromFields buffer too small for accountAddr")
		}
		cell.accountAddrLen = int(l)
		if l > 0 {
			copy(cell.accountAddr[:], data[pos:pos+int(l)])
			pos += int(l)
		}
	} else {
		cell.accountAddrLen = 0
	}
	if fieldBits&fieldStorageAddr != 0 {
		l, n := binary.Uvarint(data[pos:])
		if n == 0 {
			return 0, errors.New("fillFromFields buffer too small for storageAddr len")
		} else if n < 0 {
			return 0, errors.New("fillFromFields value overflow for storageAddr len")
		}
		pos += n
		if len(data) < pos+int(l) {
			return 0, errors.New("fillFromFields buffer too small for storageAddr")
		}
		cell.storageAddrLen = int(l)
		if l > 0 {
			copy(cell.storageAddr[:], data[pos:pos+int(l)])
			pos += int(l)
		}
	} else {
		cell.storageAddrLen = 0
	}
	if fieldBits&fieldHash != 0 {
		l, n := binary.Uvarint(data[pos:])
		if n == 0 {
			return 0, errors.New("fillFromFields buffer too small for hash len")
		} else if n < 0 {
			return 0, errors.New("fillFromFields value overflow for hash len")
		}
		pos += n
		if len(data) < pos+int(l) {
			return 0, errors.New("fillFromFields buffer too small for hash")
		}
		cell.hashLen = int(l)
		if l > 0 {
			copy(cell.hash[:], data[pos:pos+int(l)])
			pos += int(l)
		}
	} else {
		cell.hashLen = 0
	}
	return pos, nil
}

func (cell *cell) accountForHashing(buffer []byte, storageRootHash [length.Hash]byte) int {
	balanceBytes := 0
	if !cell.Balance.LtUint64(128) {
		balanceBytes = cell.Balance.ByteLen()
	}

	var nonceBytes int
	if cell.Nonce < 128 && cell.Nonce != 0 {
		nonceBytes = 0
	} else {
		nonceBytes = common.BitLenToByteLen(bits.Len64(cell.Nonce))
	}

	var structLength = uint(balanceBytes + nonceBytes + 2)
	structLength += 66 // Two 32-byte arrays + 2 prefixes

	var pos int
	if structLength < 56 {
		buffer[0] = byte(192 + structLength)
		pos = 1
	} else {
		lengthBytes := common.BitLenToByteLen(bits.Len(structLength))
		buffer[0] = byte(247 + lengthBytes)

		for i := lengthBytes; i > 0; i-- {
			buffer[i] = byte(structLength)
			structLength >>= 8
		}

		pos = lengthBytes + 1
	}

	// Encoding nonce
	if cell.Nonce < 128 && cell.Nonce != 0 {
		buffer[pos] = byte(cell.Nonce)
	} else {
		buffer[pos] = byte(128 + nonceBytes)
		var nonce = cell.Nonce
		for i := nonceBytes; i > 0; i-- {
			buffer[pos+i] = byte(nonce)
			nonce >>= 8
		}
	}
	pos += 1 + nonceBytes

	// Encoding balance
	if cell.Balance.LtUint64(128) && !cell.Balance.IsZero() {
		buffer[pos] = byte(cell.Balance.Uint64())
		pos++
	} else {
		buffer[pos] = byte(128 + balanceBytes)
		pos++
		cell.Balance.WriteToSlice(buffer[pos : pos+balanceBytes])
		pos += balanceBytes
	}

	// Encoding Root and CodeHash
	buffer[pos] = 128 + 32
	pos++
	copy(buffer[pos:], storageRootHash[:])
	pos += 32
	buffer[pos] = 128 + 32
	pos++
	copy(buffer[pos:], cell.CodeHash[:])
	pos += 32
	return pos
}

func (hph *HexPatriciaHashed) completeLeafHash(buf, keyPrefix []byte, kp, kl, compactLen int, key []byte, compact0 byte, ni int, val rlp.RlpSerializable, singleton bool) ([]byte, error) {
	totalLen := kp + kl + val.DoubleRLPLen()
	var lenPrefix [4]byte
	pt := rlp.GenerateStructLen(lenPrefix[:], totalLen)
	embedded := !singleton && totalLen+pt < length.Hash
	var writer io.Writer
	if embedded {
		//hph.byteArrayWriter.Setup(buf)
		hph.auxBuffer.Reset()
		writer = hph.auxBuffer
	} else {
		hph.keccak.Reset()
		writer = hph.keccak
	}
	if _, err := writer.Write(lenPrefix[:pt]); err != nil {
		return nil, err
	}
	if _, err := writer.Write(keyPrefix[:kp]); err != nil {
		return nil, err
	}
	var b [1]byte
	b[0] = compact0
	if _, err := writer.Write(b[:]); err != nil {
		return nil, err
	}
	for i := 1; i < compactLen; i++ {
		b[0] = key[ni]*16 + key[ni+1]
		if _, err := writer.Write(b[:]); err != nil {
			return nil, err
		}
		ni += 2
	}
	var prefixBuf [8]byte
	if err := val.ToDoubleRLP(writer, prefixBuf[:]); err != nil {
		return nil, err
	}
	if embedded {
		buf = hph.auxBuffer.Bytes()
	} else {
		var hashBuf [33]byte
		hashBuf[0] = 0x80 + length.Hash
		if _, err := hph.keccak.Read(hashBuf[1:]); err != nil {
			return nil, err
		}
		buf = append(buf, hashBuf[:]...)
	}
	return buf, nil
}

func (hph *HexPatriciaHashed) leafHashWithKeyVal(buf, key []byte, val rlp.RlpSerializableBytes, singleton bool) ([]byte, error) {
	// Compute the total length of binary representation
	var kp, kl int
	// Write key
	var compactLen int
	var ni int
	var compact0 byte
	compactLen = (len(key)-1)/2 + 1
	if len(key)&1 == 0 {
		compact0 = 0x30 + key[0] // Odd: (3<<4) + first nibble
		ni = 1
	} else {
		compact0 = 0x20
	}
	var keyPrefix [1]byte
	if compactLen > 1 {
		keyPrefix[0] = 0x80 + byte(compactLen)
		kp = 1
		kl = compactLen
	} else {
		kl = 1
	}
	return hph.completeLeafHash(buf, keyPrefix[:], kp, kl, compactLen, key, compact0, ni, val, singleton)
}

func (hph *HexPatriciaHashed) accountLeafHashWithKey(buf, key []byte, val rlp.RlpSerializable) ([]byte, error) {
	// Compute the total length of binary representation
	var kp, kl int
	// Write key
	var compactLen int
	var ni int
	var compact0 byte
	if hasTerm(key) {
		compactLen = (len(key)-1)/2 + 1
		if len(key)&1 == 0 {
			compact0 = 48 + key[0] // Odd (1<<4) + first nibble
			ni = 1
		} else {
			compact0 = 32
		}
	} else {
		compactLen = len(key)/2 + 1
		if len(key)&1 == 1 {
			compact0 = 16 + key[0] // Odd (1<<4) + first nibble
			ni = 1
		}
	}
	var keyPrefix [1]byte
	if compactLen > 1 {
		keyPrefix[0] = byte(128 + compactLen)
		kp = 1
		kl = compactLen
	} else {
		kl = 1
	}
	return hph.completeLeafHash(buf, keyPrefix[:], kp, kl, compactLen, key, compact0, ni, val, true)
}

func (hph *HexPatriciaHashed) extensionHash(key []byte, hash []byte) ([length.Hash]byte, error) {
	var hashBuf [length.Hash]byte

	// Compute the total length of binary representation
	var kp, kl int
	// Write key
	var compactLen int
	var ni int
	var compact0 byte
	if hasTerm(key) {
		compactLen = (len(key)-1)/2 + 1
		if len(key)&1 == 0 {
			compact0 = 0x30 + key[0] // Odd: (3<<4) + first nibble
			ni = 1
		} else {
			compact0 = 0x20
		}
	} else {
		compactLen = len(key)/2 + 1
		if len(key)&1 == 1 {
			compact0 = 0x10 + key[0] // Odd: (1<<4) + first nibble
			ni = 1
		}
	}
	var keyPrefix [1]byte
	if compactLen > 1 {
		keyPrefix[0] = 0x80 + byte(compactLen)
		kp = 1
		kl = compactLen
	} else {
		kl = 1
	}
	totalLen := kp + kl + 33
	var lenPrefix [4]byte
	pt := rlp.GenerateStructLen(lenPrefix[:], totalLen)
	hph.keccak.Reset()
	if _, err := hph.keccak.Write(lenPrefix[:pt]); err != nil {
		return hashBuf, err
	}
	if _, err := hph.keccak.Write(keyPrefix[:kp]); err != nil {
		return hashBuf, err
	}
	var b [1]byte
	b[0] = compact0
	if _, err := hph.keccak.Write(b[:]); err != nil {
		return hashBuf, err
	}
	for i := 1; i < compactLen; i++ {
		b[0] = key[ni]*16 + key[ni+1]
		if _, err := hph.keccak.Write(b[:]); err != nil {
			return hashBuf, err
		}
		ni += 2
	}
	b[0] = 0x80 + length.Hash
	if _, err := hph.keccak.Write(b[:]); err != nil {
		return hashBuf, err
	}
	if _, err := hph.keccak.Write(hash); err != nil {
		return hashBuf, err
	}
	// Replace previous hash with the new one
	if _, err := hph.keccak.Read(hashBuf[:]); err != nil {
		return hashBuf, err
	}
	return hashBuf, nil
}

func (hph *HexPatriciaHashed) computeCellHashLen(cell *cell, depth int) int {
	if cell.storageAddrLen > 0 && depth >= 64 {
		keyLen := 128 - depth + 1 // Length of hex key with terminator character
		var kp, kl int
		compactLen := (keyLen-1)/2 + 1
		if compactLen > 1 {
			kp = 1
			kl = compactLen
		} else {
			kl = 1
		}
		val := rlp.RlpSerializableBytes(cell.Storage[:cell.StorageLen])
		totalLen := kp + kl + val.DoubleRLPLen()
		var lenPrefix [4]byte
		pt := rlp.GenerateStructLen(lenPrefix[:], totalLen)
		if totalLen+pt < length.Hash {
			return totalLen + pt
		}
	}
	return length.Hash + 1
}

func (hph *HexPatriciaHashed) computeCellHash(cell *cell, depth int, buf []byte) ([]byte, bool, []byte, error) {
	var err error
	var storageRootHash [length.Hash]byte
	storageRootHashIsSet := false
	if cell.storageAddrLen > 0 {
		var hashedKeyOffset int
		if depth >= 64 {
			hashedKeyOffset = depth - 64
		}
		singleton := depth <= 64
		koffset := hph.accountKeyLen
		if depth == 0 && cell.accountAddrLen == 0 {
			// if account key is empty, then we need to hash storage key from the key beginning
			koffset = 0
		}
		if err := hashKey(hph.keccak, cell.storageAddr[koffset:cell.storageAddrLen], cell.hashedExtension[:], hashedKeyOffset); err != nil {
			return nil, storageRootHashIsSet, nil, err
		}
		cell.hashedExtension[64-hashedKeyOffset] = 16 // Add terminator
		if singleton {
			if hph.trace {
				fmt.Printf("leafHashWithKeyVal(singleton) for [%x]=>[%x]\n", cell.hashedExtension[:64-hashedKeyOffset+1], cell.Storage[:cell.StorageLen])
			}
			aux := make([]byte, 0, 33)
			if aux, err = hph.leafHashWithKeyVal(aux, cell.hashedExtension[:64-hashedKeyOffset+1], cell.Storage[:cell.StorageLen], true); err != nil {
				return nil, storageRootHashIsSet, nil, err
			}
			if hph.trace {
				fmt.Printf("leafHashWithKeyVal(singleton) storage hash [%x]\n", aux)
			}
			storageRootHash = *(*[length.Hash]byte)(aux[1:])
			storageRootHashIsSet = true
		} else {
			if hph.trace {
				fmt.Printf("leafHashWithKeyVal for [%x]=>[%x]\n", cell.hashedExtension[:64-hashedKeyOffset+1], cell.Storage[:cell.StorageLen])
			}
			result, err := hph.leafHashWithKeyVal(buf, cell.hashedExtension[:64-hashedKeyOffset+1], cell.Storage[:cell.StorageLen], false)
			return result, storageRootHashIsSet, storageRootHash[:], err
		}
	}
	if cell.accountAddrLen > 0 {
		if err := hashKey(hph.keccak, cell.accountAddr[:cell.accountAddrLen], cell.hashedExtension[:], depth); err != nil {
			return nil, storageRootHashIsSet, storageRootHash[:], err
		}
		cell.hashedExtension[64-depth] = 16 // Add terminator
		if !storageRootHashIsSet {
			if cell.extLen > 0 {
				// Extension
				if cell.hashLen > 0 {
					if hph.trace {
						fmt.Printf("extensionHash for [%x]=>[%x]\n", cell.extension[:cell.extLen], cell.hash[:cell.hashLen])
					}
					storageRootHashIsSet = true
					if storageRootHash, err = hph.extensionHash(cell.extension[:cell.extLen], cell.hash[:cell.hashLen]); err != nil {
						return nil, storageRootHashIsSet, storageRootHash[:], err
					}
				} else {
					return nil, storageRootHashIsSet, storageRootHash[:], errors.New("computeCellHash extension without hash")
				}
			} else if cell.hashLen > 0 {
				storageRootHash = cell.hash
				storageRootHashIsSet = true
			} else {
				storageRootHash = *(*[length.Hash]byte)(EmptyRootHash)
			}
		}
		var valBuf [128]byte
		valLen := cell.accountForHashing(valBuf[:], storageRootHash)
		if hph.trace {
			fmt.Printf("accountLeafHashWithKey for [%x]=>[%x]\n", cell.hashedExtension[:65-depth], rlp.RlpEncodedBytes(valBuf[:valLen]))
		}
		result, err := hph.accountLeafHashWithKey(buf, cell.hashedExtension[:65-depth], rlp.RlpEncodedBytes(valBuf[:valLen]))
		return result, storageRootHashIsSet, storageRootHash[:], err
	}
	buf = append(buf, 0x80+32)
	if cell.extLen > 0 {
		// Extension
		if cell.hashLen > 0 {
			if hph.trace {
				fmt.Printf("extensionHash for [%x]=>[%x]\n", cell.extension[:cell.extLen], cell.hash[:cell.hashLen])
			}
			var hash [length.Hash]byte
			if hash, err = hph.extensionHash(cell.extension[:cell.extLen], cell.hash[:cell.hashLen]); err != nil {
				return nil, storageRootHashIsSet, storageRootHash[:], err
			}
			buf = append(buf, hash[:]...)
		} else {
			return nil, storageRootHashIsSet, storageRootHash[:], errors.New("computeCellHash extension without hash")
		}
	} else if cell.hashLen > 0 {
		buf = append(buf, cell.hash[:cell.hashLen]...)
		//} else if storageRootHashIsSet {
		//	buf = append(buf, storageRootHash[:]...)
		//	copy(cell.h[:], storageRootHash[:])
	} else {
		buf = append(buf, EmptyRootHash...)
	}
	return buf, storageRootHashIsSet, storageRootHash[:], nil
}

func (hph *HexPatriciaHashed) needUnfolding(hashedKey []byte) int {
	var cell *cell
	var depth int
	if hph.activeRows == 0 {
		if hph.trace {
			fmt.Printf("needUnfolding root, rootChecked = %t\n", hph.rootChecked)
		}
		if hph.root.hashedExtLen == 0 && hph.root.hashLen == 0 {
			if hph.rootChecked {
				// Previously checked, empty root, no unfolding needed
				return 0
			}
			// Need to attempt to unfold the root
			return 1
		}
		cell = &hph.root
	} else {
		col := int(hashedKey[hph.currentKeyLen])
		cell = &hph.grid[hph.activeRows-1][col]
		depth = hph.depths[hph.activeRows-1]
		if hph.trace {
			fmt.Printf("needUnfolding cell (%d, %x), currentKey=[%x], depth=%d, cell.hash=[%x]\n", hph.activeRows-1, col, hph.currentKey[:hph.currentKeyLen], depth, cell.hash[:cell.hashLen])
		}
	}
	if len(hashedKey) <= depth {
		return 0
	}
	if cell.hashedExtLen == 0 {
		if cell.hashLen == 0 {
			// cell is empty, no need to unfold further
			return 0
		}
		// unfold branch node
		return 1
	}
	cpl := commonPrefixLen(hashedKey[depth:], cell.hashedExtension[:cell.hashedExtLen-1])
	if hph.trace {
		fmt.Printf("cpl=%d, cell.hashedExtension=[%x], depth=%d, hashedKey[depth:]=[%x]\n", cpl, cell.hashedExtension[:cell.hashedExtLen], depth, hashedKey[depth:])
	}
	unfolding := cpl + 1
	if depth < 64 && depth+unfolding > 64 {
		// This is to make sure that unfolding always breaks at the level where storage subtrees start
		unfolding = 64 - depth
		if hph.trace {
			fmt.Printf("adjusted unfolding=%d\n", unfolding)
		}
	}
	return unfolding
}

var temporalReplacementForEmpty = []byte("root")

func (c *cell) IsEmpty() bool {
	return c.hashLen == 0 && c.hashedExtLen == 0 && c.extLen == 0 && c.accountAddrLen == 0 && c.storageAddrLen == 0
}

func (c *cell) String() string {
	s := "("
	if c.hashLen > 0 {
		s += fmt.Sprintf("hash(len=%d)=%x, ", c.hashLen, c.hash)
	}
	if c.hashedExtLen > 0 {
		s += fmt.Sprintf("hashedExtension(len=%d)=%x, ", c.hashedExtLen, c.hashedExtension[:c.hashedExtLen])
	}
	if c.extLen > 0 {
		s += fmt.Sprintf("extension(len=%d)=%x, ", c.extLen, c.extension[:c.extLen])
	}
	if c.accountAddrLen > 0 {
		s += fmt.Sprintf("accountAddr=%x, ", c.accountAddr)
	}
	if c.storageAddrLen > 0 {
		s += fmt.Sprintf("storageAddr=%x, ", c.storageAddr)
	}

	s += ")"
	return s
}

func (hph *HexPatriciaHashed) PrintGrid() {
	fmt.Printf("GRID:\n")
	for row := 0; row < hph.activeRows; row++ {
		fmt.Printf("row %d depth %d:\n", row, hph.depths[row])
		for col := 0; col < 16; col++ {
			cell := &hph.grid[row][col]
			if cell.hashedExtLen > 0 || cell.accountAddrLen > 0 {
				var cellHash []byte
				hph.trace = false
				cellHash, _, _, err := hph.computeCellHash(cell, hph.depths[row], nil)
				hph.trace = true
				if err != nil {
					panic("failed to compute cell hash")
				}
				fmt.Printf("\t %x: %v cellHash=%x, \n", col, cell, cellHash)
			} else {
				fmt.Printf("\t %x: %v , \n", col, cell)
			}
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
}

func (hph *HexPatriciaHashed) ToTrie(hashedKey []byte, codeReads map[libcommon.Hash]witnesstypes.CodeWithHash) (*trie.Trie, error) {
	rootNode := &trie.FullNode{}
	var currentNode trie.Node = rootNode
	keyPos := 0 // current position in hashedKey (usually same as row, but could be different due to extension nodes)
	for row := 0; row < hph.activeRows && keyPos < len(hashedKey); row++ {
		currentNibble := hashedKey[keyPos]
		// determine the type of the next node to expand (in the next iteration)
		var nextNode trie.Node
		// need to check node type along the key path
		cellToExpand := &hph.grid[row][currentNibble]
		// determine the next node
		if cellToExpand.hashedExtLen > 0 { // extension cell
			hashedExtKey := cellToExpand.hashedExtension[:cellToExpand.hashedExtLen]
			extensionKey := make([]byte, len(hashedExtKey)+1) // +1 for the terminator 0x10 ([16]) byte
			copy(extensionKey, hashedExtKey)
			extensionKey[len(extensionKey)-1] = 16        // append terminator byte
			nextNode = &trie.ShortNode{Key: extensionKey} // Value will be in the next iteration
			keyPos += cellToExpand.hashedExtLen           // jump ahead
			if keyPos+1 == len(hashedKey) && cellToExpand.storageAddrLen > 0 {
				storageUpdate, err := hph.Ctx.Storage(cellToExpand.storageAddr[:cellToExpand.storageAddrLen])
				if err != nil {
					return nil, err
				}
				storageValueNode := trie.ValueNode(storageUpdate.Storage[:storageUpdate.StorageLen])
				nextNode = &trie.ShortNode{Key: extensionKey, Val: storageValueNode}
			}
			// // for debugging only
			// cellHash, err := hph.computeCellHash(cellToExpand, hph.depths[row], nil)
			// if err != nil {
			// 	return nil, err
			// }
			// nextNode = trie.NewHashNode(cellHash[1:])
		} else if cellToExpand.storageAddrLen > 0 { // storage cell
			storageUpdate, err := hph.Ctx.Storage(cellToExpand.storageAddr[:cellToExpand.storageAddrLen])
			if err != nil {
				return nil, err
			}
			storageValueNode := trie.ValueNode(storageUpdate.Storage[:storageUpdate.StorageLen])
			nextNode = &storageValueNode
			break
		} else if cellToExpand.accountAddrLen > 0 { // account cell
			_, storageIsSet, storageRootHash, err := hph.computeCellHash(cellToExpand, hph.depths[row], nil)
			if err != nil {
				return nil, err
			}
			accountUpdate, err := hph.Ctx.Account(cellToExpand.accountAddr[:cellToExpand.accountAddrLen])
			if err != nil {
				return nil, err
			}
			var account accounts.Account
			account.Nonce = accountUpdate.Nonce
			account.Balance = accountUpdate.Balance
			account.Initialised = true
			account.Root = accountUpdate.Storage
			account.CodeHash = accountUpdate.CodeHash

			addrHash, err := compactKey(hashedKey[:64])
			if err != nil {
				return nil, err
			}

			// get code
			var code []byte
			codeWithHash, hasCode := codeReads[[32]byte(addrHash)]
			if !hasCode {
				code = nil
			} else {
				code = codeWithHash.Code
				// sanity check
				if account.CodeHash != codeWithHash.CodeHash {
					return nil, fmt.Errorf("account.CodeHash(%x)!=codeReads[%x].CodeHash(%x)", account.CodeHash, addrHash, codeWithHash.CodeHash)
				}
			}

			if !storageIsSet {
				account.Root = trie.EmptyRoot
				nextNode = &trie.AccountNode{Account: account, Storage: nil, RootCorrect: true, Code: code, CodeSize: -1}
			} else {
				nextNode = &trie.AccountNode{Account: account, Storage: trie.NewHashNode(storageRootHash), RootCorrect: true, Code: code, CodeSize: -1}
			}
			keyPos++ // only move one nibble
		} else if cellToExpand.hashLen > 0 { // hash cell means we will expand using a full node
			nextNode = &trie.FullNode{}
			keyPos++
		} else if cellToExpand.IsEmpty() {
			nextNode = nil // no more expanding can happen (this could be due )
		} else { // default for now before we handle extLen
			nextNode = &trie.FullNode{}
			keyPos++
		}

		// process the current node
		if fullNode, ok := currentNode.(*trie.FullNode); ok { // handle full node case
			for col := 0; col < 16; col++ {
				currentCell := &hph.grid[row][col]
				if currentCell.IsEmpty() {
					fullNode.Children[col] = nil
					continue
				}
				cellHash, _, _, err := hph.computeCellHash(currentCell, hph.depths[row], nil)
				if err != nil {
					return nil, err
				}
				fullNode.Children[col] = trie.NewHashNode(cellHash[1:]) // because cellHash has 33 bytes and we want 32
			}
			fullNode.Children[currentNibble] = nextNode // ready to expand next nibble in the path                                  // move to the next nibble of the key
		} else if accNode, ok := currentNode.(*trie.AccountNode); ok {
			if len(hashedKey) <= 64 { // no storage, stop here
				nextNode = nil
				break
			}
			// there is storage so we need to expand further
			accNode.Storage = nextNode
		} else if extNode, ok := currentNode.(*trie.ShortNode); ok { // handle extension node case
			// expect only one item in this row, so take the first one
			// technically it should be at the last nibble of the key but we will adjust this later
			if extNode.Val != nil { // early termination
				break
			}
			extNode.Val = nextNode
		} else {
			break // break if currentNode is nil
		}
		// we need to check if we are dealing with the next node being an account node and we have a storage key,
		// in that case start a new tree for the storage
		if nextAccNode, ok := nextNode.(*trie.AccountNode); ok && len(hashedKey) > 64 {
			nextNode = &trie.FullNode{}
			nextAccNode.Storage = nextNode
		}
		currentNode = nextNode
	}
	tr := trie.NewInMemoryTrie(rootNode)
	return tr, nil
}

// unfoldBranchNode returns true if unfolding has been done
func (hph *HexPatriciaHashed) unfoldBranchNode(row int, deleted bool, depth int) (bool, error) {
	key := hexToCompact(hph.currentKey[:hph.currentKeyLen])
	if len(key) == 0 {
		key = temporalReplacementForEmpty
	}
	branchData, _, err := hph.Ctx.Branch(key)
	if err != nil {
		return false, err
	}
	if len(branchData) >= 2 {
		branchData = branchData[2:] // skip touch map and hold aftermap and rest
	}
	if hph.trace {
		fmt.Printf("unfoldBranchNode prefix '%x', compacted [%x] depth %d row %d '%x'\n", key, hph.currentKey[:hph.currentKeyLen], depth, row, branchData)
	}
	if !hph.rootChecked && hph.currentKeyLen == 0 && len(branchData) == 0 {
		// Special case - empty or deleted root
		hph.rootChecked = true
		return false, nil
	}
	if len(branchData) == 0 {
		log.Warn("got empty branch data during unfold", "key", hex.EncodeToString(key), "row", row, "depth", depth, "deleted", deleted)
		return false, fmt.Errorf("empty branch data read during unfold, prefix %x", hexToCompact(hph.currentKey[:hph.currentKeyLen]))
	}
	hph.branchBefore[row] = true
	bitmap := binary.BigEndian.Uint16(branchData[0:])
	pos := 2
	if deleted {
		// All cells come as deleted (touched but not present after)
		hph.afterMap[row] = 0
		hph.touchMap[row] = bitmap
	} else {
		hph.afterMap[row] = bitmap
		hph.touchMap[row] = 0
	}
	//fmt.Printf("unfoldBranchNode prefix '%x' [%x], afterMap = [%016b], touchMap = [%016b]\n", key, branchData, hph.afterMap[row], hph.touchMap[row])
	// Loop iterating over the set bits of modMask
	for bitset, j := bitmap, 0; bitset != 0; j++ {
		bit := bitset & -bitset
		nibble := bits.TrailingZeros16(bit)
		cell := &hph.grid[row][nibble]
		fieldBits := branchData[pos]
		pos++
		var err error
		if pos, err = cell.fillFromFields(branchData, pos, cellFields(fieldBits)); err != nil {
			return false, fmt.Errorf("prefix [%x], branchData[%x]: %w", hph.currentKey[:hph.currentKeyLen], branchData, err)
		}
		if hph.trace {
			fmt.Printf("cell (%d, %x) depth=%d, hash=[%x], accountAddr=[%x], storageAddr=[%x], extension=[%x]\n", row, nibble, depth, cell.hash[:cell.hashLen], cell.accountAddr[:cell.accountAddrLen], cell.storageAddr[:cell.storageAddrLen], cell.extension[:cell.extLen])
		}
		if cell.accountAddrLen > 0 {
			update, err := hph.Ctx.Account(cell.accountAddr[:cell.accountAddrLen])
			if err != nil {
				return false, fmt.Errorf("unfoldBranchNode GetAccount: %w", err)
			}
			cell.setFromUpdate(update)
			if hph.trace {
				fmt.Printf("GetAccount[%x] return balance=%d, nonce=%d code=%x\n", cell.accountAddr[:cell.accountAddrLen], &cell.Balance, cell.Nonce, cell.CodeHash[:])
			}
		}
		if cell.storageAddrLen > 0 {
			update, err := hph.Ctx.Storage(cell.storageAddr[:cell.storageAddrLen])
			if err != nil {
				return false, fmt.Errorf("unfoldBranchNode GetAccount: %w", err)
			}
			cell.setFromUpdate(update)
		}
		if err = cell.deriveHashedKeys(depth, hph.keccak, hph.accountKeyLen); err != nil {
			return false, err
		}
		bitset ^= bit
	}
	return true, nil
}

func (hph *HexPatriciaHashed) unfold(hashedKey []byte, unfolding int) error {
	if hph.trace {
		fmt.Printf("unfold %d: activeRows: %d\n", unfolding, hph.activeRows)
	}
	var upCell *cell
	var touched, present bool
	var col byte
	var upDepth, depth int
	if hph.activeRows == 0 {
		if hph.rootChecked && hph.root.hashLen == 0 && hph.root.hashedExtLen == 0 {
			// No unfolding for empty root
			return nil
		}
		upCell = &hph.root
		touched = hph.rootTouched
		present = hph.rootPresent
		if hph.trace {
			fmt.Printf("unfold root, touched %t, present %t, column %d hashedExtension %x\n", touched, present, col, upCell.hashedExtension[:upCell.hashedExtLen])
		}
	} else {
		upDepth = hph.depths[hph.activeRows-1]
		col = hashedKey[upDepth-1]
		upCell = &hph.grid[hph.activeRows-1][col]
		touched = hph.touchMap[hph.activeRows-1]&(uint16(1)<<col) != 0
		present = hph.afterMap[hph.activeRows-1]&(uint16(1)<<col) != 0
		if hph.trace {
			fmt.Printf("upCell (%d, %x), touched %t, present %t\n", hph.activeRows-1, col, touched, present)
		}
		hph.currentKey[hph.currentKeyLen] = col
		hph.currentKeyLen++
	}
	row := hph.activeRows
	for i := 0; i < 16; i++ {
		hph.grid[row][i].reset()
	}
	hph.touchMap[row] = 0
	hph.afterMap[row] = 0
	hph.branchBefore[row] = false

	if upCell.hashedExtLen == 0 { // normal branch node, but could also be just having hash set
		// root unfolded
		depth = upDepth + 1
		if unfolded, err := hph.unfoldBranchNode(row, touched && !present /* deleted */, depth); err != nil {
			return err
		} else if !unfolded {
			// Return here to prevent activeRow from being incremented
			return nil
		}
	} else if upCell.hashedExtLen >= unfolding { // extension node
		depth = upDepth + unfolding
		nibble := upCell.hashedExtension[unfolding-1]
		if touched {
			hph.touchMap[row] = uint16(1) << nibble
		}
		if present {
			hph.afterMap[row] = uint16(1) << nibble
		}
		cell := &hph.grid[row][nibble]
		cell.fillFromUpperCell(upCell, depth, unfolding) // upcell is the account cell before the extension
		if hph.trace {
			fmt.Printf("cell (%d, %x) depth=%d\n", row, nibble, depth)
		}
		if row >= 64 {
			cell.accountAddrLen = 0
		}
		if unfolding > 1 {
			copy(hph.currentKey[hph.currentKeyLen:], upCell.hashedExtension[:unfolding-1])
		}
		hph.currentKeyLen += unfolding - 1
	} else {
		// upCell.hashedExtLen < unfolding
		depth = upDepth + upCell.hashedExtLen
		nibble := upCell.hashedExtension[upCell.hashedExtLen-1]
		if touched {
			hph.touchMap[row] = uint16(1) << nibble
		}
		if present {
			hph.afterMap[row] = uint16(1) << nibble
		}
		cell := &hph.grid[row][nibble]
		cell.fillFromUpperCell(upCell, depth, upCell.hashedExtLen)
		if hph.trace {
			fmt.Printf("cell (%d, %x) depth=%d\n", row, nibble, depth)
		}
		if row >= 64 {
			cell.accountAddrLen = 0
		}
		if upCell.hashedExtLen > 1 {
			copy(hph.currentKey[hph.currentKeyLen:], upCell.hashedExtension[:upCell.hashedExtLen-1])
		}
		hph.currentKeyLen += upCell.hashedExtLen - 1
	}
	hph.depths[hph.activeRows] = depth
	hph.activeRows++
	return nil
}

func (hph *HexPatriciaHashed) needFolding(hashedKey []byte) bool {
	return !bytes.HasPrefix(hashedKey, hph.currentKey[:hph.currentKeyLen])
}

// The purpose of fold is to reduce hph.currentKey[:hph.currentKeyLen]. It should be invoked
// until that current key becomes a prefix of hashedKey that we will process next
// (in other words until the needFolding function returns 0)
func (hph *HexPatriciaHashed) fold() (err error) {
	updateKeyLen := hph.currentKeyLen
	if hph.activeRows == 0 {
		return errors.New("cannot fold - no active rows")
	}
	if hph.trace {
		fmt.Printf("fold [%x] activeRows: %d touchMap: %016b afterMap: %016b\n", hph.currentKey[:hph.currentKeyLen], hph.activeRows, hph.touchMap[hph.activeRows-1], hph.afterMap[hph.activeRows-1])
	}
	// Move information to the row above
	var upCell *cell
	var nibble, upDepth int
	row := hph.activeRows - 1
	if row == 0 {
		if hph.trace {
			fmt.Printf("upcell is root\n")
		}
		upCell = &hph.root
	} else {
		upDepth = hph.depths[hph.activeRows-2]
		nibble = int(hph.currentKey[upDepth-1])
		if hph.trace {
			fmt.Printf("upCell is (%d, %x, upDepth=%d)\n", row-1, nibble, upDepth)
		}
		upCell = &hph.grid[row-1][nibble]
	}

	depth := hph.depths[row]
	updateKey := hexToCompact(hph.currentKey[:updateKeyLen])
	if len(updateKey) == 0 {
		updateKey = temporalReplacementForEmpty
	}
	partsCount := bits.OnesCount16(hph.afterMap[row])

	if hph.trace {
		fmt.Printf("current key %x touchMap[%d]=%016b afterMap[%d]=%016b\n", hph.currentKey[:hph.currentKeyLen], row, hph.touchMap[row], row, hph.afterMap[row])
	}
	switch partsCount {
	case 0:
		// Everything deleted
		if hph.touchMap[row] != 0 {
			if row == 0 {
				// Root is deleted because the tree is empty
				hph.rootTouched = true
				hph.rootPresent = false
			} else if upDepth == 64 {
				// Special case - all storage items of an account have been deleted, but it does not automatically delete the account, just makes it empty storage
				// Therefore we are not propagating deletion upwards, but turn it into a modification
				hph.touchMap[row-1] |= uint16(1) << nibble
			} else {
				// Deletion is propagated upwards
				hph.touchMap[row-1] |= uint16(1) << nibble
				hph.afterMap[row-1] &^= uint16(1) << nibble
			}
		}
		upCell.hashLen = 0
		upCell.accountAddrLen = 0
		upCell.storageAddrLen = 0
		upCell.extLen = 0
		upCell.hashedExtLen = 0
		if hph.branchBefore[row] {
			_, err := hph.branchEncoder.CollectUpdate(hph.Ctx, updateKey, 0, hph.touchMap[row], 0, RetrieveCellNoop)
			if err != nil {
				return fmt.Errorf("failed to encode leaf node update: %w", err)
			}
		}
		hph.activeRows--
		if upDepth > 0 {
			hph.currentKeyLen = upDepth - 1
		} else {
			hph.currentKeyLen = 0
		}
	case 1:
		// Leaf or extension node
		if hph.touchMap[row] != 0 {
			// any modifications
			if row == 0 {
				hph.rootTouched = true
			} else {
				// Modifiction is propagated upwards
				hph.touchMap[row-1] |= (uint16(1) << nibble)
			}
		}
		nibble := bits.TrailingZeros16(hph.afterMap[row])
		cell := &hph.grid[row][nibble]
		upCell.extLen = 0
		upCell.fillFromLowerCell(cell, depth, hph.currentKey[upDepth:hph.currentKeyLen], nibble)
		// Delete if it existed
		if hph.branchBefore[row] {
			_, err := hph.branchEncoder.CollectUpdate(hph.Ctx, updateKey, 0, hph.touchMap[row], 0, RetrieveCellNoop)
			if err != nil {
				return fmt.Errorf("failed to encode leaf node update: %w", err)
			}
		}
		hph.activeRows--
		if upDepth > 0 {
			hph.currentKeyLen = upDepth - 1
		} else {
			hph.currentKeyLen = 0
		}
	default:
		// Branch node
		if hph.touchMap[row] != 0 {
			// any modifications
			if row == 0 {
				hph.rootTouched = true
			} else {
				// Modifiction is propagated upwards
				hph.touchMap[row-1] |= (uint16(1) << nibble)
			}
		}
		bitmap := hph.touchMap[row] & hph.afterMap[row]
		if !hph.branchBefore[row] {
			// There was no branch node before, so we need to touch even the singular child that existed
			hph.touchMap[row] |= hph.afterMap[row]
			bitmap |= hph.afterMap[row]
		}
		// Calculate total length of all hashes
		totalBranchLen := 17 - partsCount // For every empty cell, one byte
		for bitset, j := hph.afterMap[row], 0; bitset != 0; j++ {
			bit := bitset & -bitset
			nibble := bits.TrailingZeros16(bit)
			cell := &hph.grid[row][nibble]
			totalBranchLen += hph.computeCellHashLen(cell, depth)
			bitset ^= bit
		}

		hph.keccak2.Reset()
		pt := rlp.GenerateStructLen(hph.hashAuxBuffer[:], totalBranchLen)
		if _, err := hph.keccak2.Write(hph.hashAuxBuffer[:pt]); err != nil {
			return err
		}

		b := [...]byte{0x80}
		cellGetter := func(nibble int, skip bool) (*cell, error) {
			if skip {
				if _, err := hph.keccak2.Write(b[:]); err != nil {
					return nil, fmt.Errorf("failed to write empty nibble to hash: %w", err)
				}
				if hph.trace {
					fmt.Printf("%x: empty(%d,%x)\n", nibble, row, nibble)
				}
				return nil, nil
			}
			cell := &hph.grid[row][nibble]
			cellHash, _, _, err := hph.computeCellHash(cell, depth, hph.hashAuxBuffer[:0])
			if err != nil {
				return nil, err
			}
			if hph.trace {
				fmt.Printf("%x: computeCellHash(%d,%x,depth=%d)=[%x]\n", nibble, row, nibble, depth, cellHash)
			}
			if _, err := hph.keccak2.Write(cellHash); err != nil {
				return nil, err
			}

			return cell, nil
		}

		var lastNibble int
		var err error

		lastNibble, err = hph.branchEncoder.CollectUpdate(hph.Ctx, updateKey, bitmap, hph.touchMap[row], hph.afterMap[row], cellGetter)
		if err != nil {
			return fmt.Errorf("failed to encode branch update: %w", err)
		}
		for i := lastNibble; i < 17; i++ {
			if _, err := hph.keccak2.Write(b[:]); err != nil {
				return err
			}
			if hph.trace {
				fmt.Printf("%x: empty(%d,%x)\n", i, row, i)
			}
		}
		upCell.extLen = depth - upDepth - 1
		upCell.hashedExtLen = upCell.extLen
		if upCell.extLen > 0 {
			copy(upCell.extension[:], hph.currentKey[upDepth:hph.currentKeyLen])
			copy(upCell.hashedExtension[:], hph.currentKey[upDepth:hph.currentKeyLen])
		}
		if depth < 64 {
			upCell.accountAddrLen = 0
		}
		upCell.storageAddrLen = 0
		upCell.hashLen = 32
		if _, err := hph.keccak2.Read(upCell.hash[:]); err != nil {
			return err
		}
		if hph.trace {
			fmt.Printf("} [%x]\n", upCell.hash[:])
		}
		hph.activeRows--
		if upDepth > 0 {
			hph.currentKeyLen = upDepth - 1
		} else {
			hph.currentKeyLen = 0
		}
	}
	return nil
}

func (hph *HexPatriciaHashed) deleteCell(hashedKey []byte) {
	if hph.trace {
		fmt.Printf("deleteCell, activeRows = %d\n", hph.activeRows)
	}
	var cell *cell
	if hph.activeRows == 0 { // Remove the root
		cell = &hph.root
		hph.rootTouched, hph.rootPresent = true, false
	} else {
		row := hph.activeRows - 1
		if hph.depths[row] < len(hashedKey) {
			if hph.trace {
				fmt.Printf("deleteCell skipping spurious delete depth=%d, len(hashedKey)=%d\n", hph.depths[row], len(hashedKey))
			}
			return
		}
		nibble := int(hashedKey[hph.currentKeyLen])
		cell = &hph.grid[row][nibble]
		col := uint16(1) << nibble
		if hph.afterMap[row]&col != 0 {
			// Prevent "spurios deletions", i.e. deletion of absent items
			hph.touchMap[row] |= col
			hph.afterMap[row] &^= col
			if hph.trace {
				fmt.Printf("deleteCell setting (%d, %x)\n", row, nibble)
			}
		} else {
			if hph.trace {
				fmt.Printf("deleteCell ignoring (%d, %x)\n", row, nibble)
			}
		}
	}
	cell.reset()
}

// fetches cell by key and set touch/after maps
func (hph *HexPatriciaHashed) updateCell(plainKey, hashedKey []byte, u *Update) (cell *cell) {
	if u.Deleted() {
		hph.deleteCell(hashedKey)
		return nil
	}

	var depth int
	if hph.activeRows == 0 {
		cell = &hph.root
		hph.rootTouched, hph.rootPresent = true, true
	} else {
		row := hph.activeRows - 1
		depth = hph.depths[row]
		nibble := int(hashedKey[hph.currentKeyLen])
		cell = &hph.grid[row][nibble]
		col := uint16(1) << nibble

		hph.touchMap[row] |= col
		hph.afterMap[row] |= col
		if hph.trace {
			fmt.Printf("updateCell setting (%d, %x), depth=%d\n", row, nibble, depth)
		}
	}
	if cell.hashedExtLen == 0 {
		copy(cell.hashedExtension[:], hashedKey[depth:])
		cell.hashedExtLen = len(hashedKey) - depth
		if hph.trace {
			fmt.Printf("set downHasheKey=[%x]\n", cell.hashedExtension[:cell.hashedExtLen])
		}
	} else {
		if hph.trace {
			fmt.Printf("keep downHasheKey=[%x]\n", cell.hashedExtension[:cell.hashedExtLen])
		}
	}
	if len(plainKey) == hph.accountKeyLen {
		cell.accountAddrLen = len(plainKey)
		copy(cell.accountAddr[:], plainKey)
		copy(cell.CodeHash[:], EmptyCodeHash)
	} else { // set storage key
		cell.storageAddrLen = len(plainKey)
		copy(cell.storageAddr[:], plainKey)
	}

	cell.setFromUpdate(u)
	if hph.trace {
		fmt.Printf("updateCell %x => %s\n", plainKey, u.String())
	}
	return cell
}

func (hph *HexPatriciaHashed) RootHash() ([]byte, error) {
	rootHash, _, _, err := hph.computeCellHash(&hph.root, 0, nil)
	if err != nil {
		return nil, err
	}
	return rootHash[1:], nil // first byte is 128+hash_len=160
}

func (hph *HexPatriciaHashed) GenerateWitness(ctx context.Context, updates *Updates, codeReads map[libcommon.Hash]witnesstypes.CodeWithHash, expectedRootHash []byte, logPrefix string) (witnessTrie *trie.Trie, rootHash []byte, err error) {
	var (
		m      runtime.MemStats
		ki     uint64
		update *Update

		updatesCount = updates.Size()
		logEvery     = time.NewTicker(20 * time.Second)
	)
	defer logEvery.Stop()
	var tries []*trie.Trie = make([]*trie.Trie, 0, len(updates.keys)) // slice of tries, i.e the witness for each key, these will be all merged into single trie
	err = updates.HashSort(ctx, func(hashedKey, plainKey []byte, stateUpdate *Update) error {
		select {
		case <-logEvery.C:
			dbg.ReadMemStats(&m)
			log.Info(fmt.Sprintf("[%s][agg] computing trie", logPrefix),
				"progress", fmt.Sprintf("%s/%s", common.PrettyCounter(ki), common.PrettyCounter(updatesCount)),
				"alloc", common.ByteCount(m.Alloc), "sys", common.ByteCount(m.Sys))

		default:
		}

		var tr *trie.Trie
		var computedRootHash []byte

		fmt.Printf("\n%d/%d) plainKey [%x] hashedKey [%x] currentKey [%x]\n", ki+1, updatesCount, plainKey, hashedKey, hph.currentKey[:hph.currentKeyLen])

		// Keep folding until the currentKey is the prefix of the key we modify
		for hph.needFolding(hashedKey) {
			if err := hph.fold(); err != nil {
				return fmt.Errorf("fold: %w", err)
			}
		}
		// Now unfold until we step on an empty cell
		for unfolding := hph.needUnfolding(hashedKey); unfolding > 0; unfolding = hph.needUnfolding(hashedKey) {
			if err := hph.unfold(hashedKey, unfolding); err != nil {
				return fmt.Errorf("unfold: %w", err)
			}
		}

		fmt.Printf("\n%d/%d) PRINT plainKey [%x] hashedKey [%x] currentKey [%x]\n", ki+1, updatesCount, plainKey, hashedKey, hph.currentKey[:hph.currentKeyLen])
		hph.PrintGrid()

		if len(plainKey) == 20 { // account
			account, err := hph.Ctx.Account(plainKey)
			if err != nil {
				return fmt.Errorf("account with plainkey=%x not found: %w", plainKey, err)
			} else {
				fmt.Printf("account FOUND = %v\n", account)
			}
		} else {
			storage, err := hph.Ctx.Storage(plainKey)
			if err != nil {
				return fmt.Errorf("storage with plainkey=%x not found: %w", plainKey, err)
			}
			fmt.Printf("storage FOUND = %v\n", storage.Storage)
			// hashedKeyForStorage := crypto.Keccak256Hash(plainKey)
			// address := plainKey[:20]
			// addrHash := crypto.Keccak256(address)
			// // Concatenate address and hashedKeyForStorage into a new slice
			// fullHashedKeyForStorage := make([]byte, len(addrHash)+len(hashedKeyForStorage.Bytes()))
			// copy(fullHashedKeyForStorage, addrHash)
			// copy(fullHashedKeyForStorage[32:], hashedKeyForStorage.Bytes())

			// fullNibblizedKey := nibblize(fullHashedKeyForStorage)

			// fmt.Printf("fullNibblizedKey = %x\n", fullNibblizedKey)
			// fmt.Printf("hashedKey = %x\n", hashedKey)
			// if !bytes.Equal(fullNibblizedKey, hashedKey) {
			// 	panic("WRONG hashedkey for storage")
			// }
		}

		tr, err = hph.ToTrie(hashedKey, codeReads) // build witness trie for this key, based on the current state of the grid
		if err != nil {
			return err
		}
		computedRootHash = tr.Root()
		fmt.Printf("computedRootHash = %x\n", computedRootHash)

		if !bytes.Equal(computedRootHash, expectedRootHash) {
			err = fmt.Errorf("ROOT HASH MISMATCH computedRootHash(%x)!=expectedRootHash(%x)", computedRootHash, expectedRootHash)
			log.Error(err.Error())
		}

		tries = append(tries, tr)

		// // actual root hash
		// actualRootHash, err := hph.RootHash()
		// if err != nil {
		// 	return err
		// }
		// fmt.Printf("actualRootHash= %x\n", actualRootHash)
		// if !bytes.Equal(computedRootHash, actualRootHash) {
		// 	return fmt.Errorf("root hash mismatch computed(using trie.Trie)=%x , actual(using HexPatriciaHashed)=%x", computedRootHash, actualRootHash)
		// }

		if stateUpdate == nil {
			// Update the cell
			if len(plainKey) == hph.accountKeyLen {
				update, err = hph.Ctx.Account(plainKey)
				if err != nil {
					return fmt.Errorf("GetAccount for key %x failed: %w", plainKey, err)
				}
			} else {
				update, err = hph.Ctx.Storage(plainKey)
				if err != nil {
					return fmt.Errorf("GetStorage for key %x failed: %w", plainKey, err)
				}
			}
		} else {
			if update == nil {
				update = stateUpdate
			} else {
				update.Reset()
				update.Merge(stateUpdate)
			}
		}

		hph.updateCell(plainKey, hashedKey, update)

		mxKeys.Inc()
		ki++
		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("hash sort failed: %w", err)
	}

	// Folding everything up to the root
	for hph.activeRows > 0 {
		if err := hph.fold(); err != nil {
			return nil, nil, fmt.Errorf("final fold: %w", err)
		}
	}

	rootHash, err = hph.RootHash()
	if err != nil {
		return nil, nil, fmt.Errorf("root hash evaluation failed: %w", err)
	}
	if hph.trace {
		fmt.Printf("root hash %x updates %d\n", rootHash, updatesCount)
	}
	err = hph.branchEncoder.Load(hph.Ctx, etl.TransformArgs{Quit: ctx.Done()})
	if err != nil {
		return nil, nil, fmt.Errorf("branch update failed: %w", err)
	}

	witnessTrie, err = trie.MergeTries(tries)
	if err != nil {
		return nil, nil, err
	}

	witnessTrieRootHash := witnessTrie.Root()

	fmt.Printf("mergedTrieRootHash = %x\n", witnessTrieRootHash)

	if !bytes.Equal(witnessTrieRootHash, expectedRootHash) {
		return nil, nil, fmt.Errorf("ROOT HASH MISMATCH witnessTrieRootHash(%x)!=expectedRootHash(%x)", witnessTrieRootHash, expectedRootHash)
	}

	return witnessTrie, rootHash, nil
}

func (hph *HexPatriciaHashed) Process(ctx context.Context, updates *Updates, logPrefix string) (rootHash []byte, err error) {
	var (
		m      runtime.MemStats
		ki     uint64
		update *Update

		updatesCount = updates.Size()
		logEvery     = time.NewTicker(20 * time.Second)
	)
	defer logEvery.Stop()
	hph.trace = true
	err = updates.HashSort(ctx, func(hashedKey, plainKey []byte, stateUpdate *Update) error {
		select {
		case <-logEvery.C:
			dbg.ReadMemStats(&m)
			log.Info(fmt.Sprintf("[%s][agg] computing trie", logPrefix),
				"progress", fmt.Sprintf("%s/%s", common.PrettyCounter(ki), common.PrettyCounter(updatesCount)),
				"alloc", common.ByteCount(m.Alloc), "sys", common.ByteCount(m.Sys))

		default:
		}

		if hph.trace {
			fmt.Printf("\n%d/%d) plainKey [%x] hashedKey [%x] currentKey [%x]\n", ki+1, updatesCount, plainKey, hashedKey, hph.currentKey[:hph.currentKeyLen])
		}
		// Keep folding until the currentKey is the prefix of the key we modify
		for hph.needFolding(hashedKey) {
			if err := hph.fold(); err != nil {
				return fmt.Errorf("fold: %w", err)
			}
		}
		// Now unfold until we step on an empty cell
		for unfolding := hph.needUnfolding(hashedKey); unfolding > 0; unfolding = hph.needUnfolding(hashedKey) {
			if err := hph.unfold(hashedKey, unfolding); err != nil {
				return fmt.Errorf("unfold: %w", err)
			}
		}

		if stateUpdate == nil {
			// Update the cell
			if len(plainKey) == hph.accountKeyLen {
				update, err = hph.Ctx.Account(plainKey)
				if err != nil {
					return fmt.Errorf("GetAccount for key %x failed: %w", plainKey, err)
				}
			} else {
				update, err = hph.Ctx.Storage(plainKey)
				if err != nil {
					return fmt.Errorf("GetStorage for key %x failed: %w", plainKey, err)
				}
			}
		} else {
			if update == nil {
				update = stateUpdate
			} else {
				update.Reset()
				update.Merge(stateUpdate)
			}
		}
		hph.updateCell(plainKey, hashedKey, update)

		mxKeys.Inc()
		ki++
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("hash sort failed: %w", err)
	}

	// Folding everything up to the root
	for hph.activeRows > 0 {
		if err := hph.fold(); err != nil {
			return nil, fmt.Errorf("final fold: %w", err)
		}
	}

	rootHash, err = hph.RootHash()
	if err != nil {
		return nil, fmt.Errorf("root hash evaluation failed: %w", err)
	}
	if hph.trace {
		fmt.Printf("root hash %x updates %d\n", rootHash, updatesCount)
	}
	err = hph.branchEncoder.Load(hph.Ctx, etl.TransformArgs{Quit: ctx.Done()})
	if err != nil {
		return nil, fmt.Errorf("branch update failed: %w", err)
	}
	return rootHash, nil
}

func (hph *HexPatriciaHashed) SetTrace(trace bool) { hph.trace = trace }

func (hph *HexPatriciaHashed) Variant() TrieVariant { return VariantHexPatriciaTrie }

// Reset allows HexPatriciaHashed instance to be reused for the new commitment calculation
func (hph *HexPatriciaHashed) Reset() {
	hph.root.reset()
	hph.rootTouched = false
	hph.rootChecked = false
	hph.rootPresent = true
}

func (hph *HexPatriciaHashed) ResetContext(ctx PatriciaContext) {
	hph.Ctx = ctx
}

type stateRootFlag int8

var (
	stateRootPresent stateRootFlag = 1
	stateRootChecked stateRootFlag = 2
	stateRootTouched stateRootFlag = 4
)

// represents state of the tree
type state struct {
	Root         []byte      // encoded root cell
	Depths       [128]int    // For each row, the depth of cells in that row
	TouchMap     [128]uint16 // For each row, bitmap of cells that were either present before modification, or modified or deleted
	AfterMap     [128]uint16 // For each row, bitmap of cells that were present after modification
	BranchBefore [128]bool   // For each row, whether there was a branch node in the database loaded in unfold
	RootChecked  bool        // Set to false if it is not known whether the root is empty, set to true if it is checked
	RootTouched  bool
	RootPresent  bool
}

func (s *state) Encode(buf []byte) ([]byte, error) {
	var rootFlags stateRootFlag
	if s.RootPresent {
		rootFlags |= stateRootPresent
	}
	if s.RootChecked {
		rootFlags |= stateRootChecked
	}
	if s.RootTouched {
		rootFlags |= stateRootTouched
	}

	ee := bytes.NewBuffer(buf)
	if err := binary.Write(ee, binary.BigEndian, int8(rootFlags)); err != nil {
		return nil, fmt.Errorf("encode rootFlags: %w", err)
	}
	if err := binary.Write(ee, binary.BigEndian, uint16(len(s.Root))); err != nil {
		return nil, fmt.Errorf("encode root len: %w", err)
	}
	if n, err := ee.Write(s.Root); err != nil || n != len(s.Root) {
		return nil, fmt.Errorf("encode root: %w", err)
	}
	d := make([]byte, len(s.Depths))
	for i := 0; i < len(s.Depths); i++ {
		d[i] = byte(s.Depths[i])
	}
	if n, err := ee.Write(d); err != nil || n != len(s.Depths) {
		return nil, fmt.Errorf("encode depths: %w", err)
	}
	if err := binary.Write(ee, binary.BigEndian, s.TouchMap); err != nil {
		return nil, fmt.Errorf("encode touchMap: %w", err)
	}
	if err := binary.Write(ee, binary.BigEndian, s.AfterMap); err != nil {
		return nil, fmt.Errorf("encode afterMap: %w", err)
	}

	var before1, before2 uint64
	for i := 0; i < 64; i++ {
		if s.BranchBefore[i] {
			before1 |= 1 << i
		}
	}
	for i, j := 64, 0; i < 128; i, j = i+1, j+1 {
		if s.BranchBefore[i] {
			before2 |= 1 << j
		}
	}
	if err := binary.Write(ee, binary.BigEndian, before1); err != nil {
		return nil, fmt.Errorf("encode branchBefore_1: %w", err)
	}
	if err := binary.Write(ee, binary.BigEndian, before2); err != nil {
		return nil, fmt.Errorf("encode branchBefore_2: %w", err)
	}
	return ee.Bytes(), nil
}

func (s *state) Decode(buf []byte) error {
	aux := bytes.NewBuffer(buf)
	var rootFlags stateRootFlag
	if err := binary.Read(aux, binary.BigEndian, &rootFlags); err != nil {
		return fmt.Errorf("rootFlags: %w", err)
	}

	if rootFlags&stateRootPresent != 0 {
		s.RootPresent = true
	}
	if rootFlags&stateRootTouched != 0 {
		s.RootTouched = true
	}
	if rootFlags&stateRootChecked != 0 {
		s.RootChecked = true
	}

	var rootSize uint16
	if err := binary.Read(aux, binary.BigEndian, &rootSize); err != nil {
		return fmt.Errorf("root size: %w", err)
	}
	s.Root = make([]byte, rootSize)
	if _, err := aux.Read(s.Root); err != nil {
		return fmt.Errorf("root: %w", err)
	}
	d := make([]byte, len(s.Depths))
	if err := binary.Read(aux, binary.BigEndian, &d); err != nil {
		return fmt.Errorf("depths: %w", err)
	}
	for i := 0; i < len(s.Depths); i++ {
		s.Depths[i] = int(d[i])
	}
	if err := binary.Read(aux, binary.BigEndian, &s.TouchMap); err != nil {
		return fmt.Errorf("touchMap: %w", err)
	}
	if err := binary.Read(aux, binary.BigEndian, &s.AfterMap); err != nil {
		return fmt.Errorf("afterMap: %w", err)
	}
	var branch1, branch2 uint64
	if err := binary.Read(aux, binary.BigEndian, &branch1); err != nil {
		return fmt.Errorf("branchBefore1: %w", err)
	}
	if err := binary.Read(aux, binary.BigEndian, &branch2); err != nil {
		return fmt.Errorf("branchBefore2: %w", err)
	}

	for i := 0; i < 64; i++ {
		if branch1&(1<<i) != 0 {
			s.BranchBefore[i] = true
		}
	}
	for i, j := 64, 0; i < 128; i, j = i+1, j+1 {
		if branch2&(1<<j) != 0 {
			s.BranchBefore[i] = true
		}
	}
	return nil
}

func (cell *cell) Encode() []byte {
	var pos = 1
	size := pos + 5 + cell.hashLen + cell.accountAddrLen + cell.storageAddrLen + cell.hashedExtLen + cell.extLen // max size
	buf := make([]byte, size)

	var flags uint8
	if cell.hashLen != 0 {
		flags |= cellFlagHash
		buf[pos] = byte(cell.hashLen)
		pos++
		copy(buf[pos:pos+cell.hashLen], cell.hash[:])
		pos += cell.hashLen
	}
	if cell.accountAddrLen != 0 {
		flags |= cellFlagAccount
		buf[pos] = byte(cell.accountAddrLen)
		pos++
		copy(buf[pos:pos+cell.accountAddrLen], cell.accountAddr[:])
		pos += cell.accountAddrLen
	}
	if cell.storageAddrLen != 0 {
		flags |= cellFlagStorage
		buf[pos] = byte(cell.storageAddrLen)
		pos++
		copy(buf[pos:pos+cell.storageAddrLen], cell.storageAddr[:])
		pos += cell.storageAddrLen
	}
	if cell.hashedExtLen != 0 {
		flags |= cellFlagDownHash
		buf[pos] = byte(cell.hashedExtLen)
		pos++
		copy(buf[pos:pos+cell.hashedExtLen], cell.hashedExtension[:cell.hashedExtLen])
		pos += cell.hashedExtLen
	}
	if cell.extLen != 0 {
		flags |= cellFlagExtension
		buf[pos] = byte(cell.extLen)
		pos++
		copy(buf[pos:pos+cell.extLen], cell.extension[:])
		pos += cell.extLen //nolint
	}
	if cell.Deleted() {
		flags |= cellFlagDelete
	}
	buf[0] = flags
	return buf
}

const (
	cellFlagHash = uint8(1 << iota)
	cellFlagAccount
	cellFlagStorage
	cellFlagDownHash
	cellFlagExtension
	cellFlagDelete
)

func (cell *cell) Decode(buf []byte) error {
	if len(buf) < 1 {
		return errors.New("invalid buffer size to contain cell (at least 1 byte expected)")
	}
	cell.reset()

	var pos int
	flags := buf[pos]
	pos++

	if flags&cellFlagHash != 0 {
		cell.hashLen = int(buf[pos])
		pos++
		copy(cell.hash[:], buf[pos:pos+cell.hashLen])
		pos += cell.hashLen
	}
	if flags&cellFlagAccount != 0 {
		cell.accountAddrLen = int(buf[pos])
		pos++
		copy(cell.accountAddr[:], buf[pos:pos+cell.accountAddrLen])
		pos += cell.accountAddrLen
	}
	if flags&cellFlagStorage != 0 {
		cell.storageAddrLen = int(buf[pos])
		pos++
		copy(cell.storageAddr[:], buf[pos:pos+cell.storageAddrLen])
		pos += cell.storageAddrLen
	}
	if flags&cellFlagDownHash != 0 {
		cell.hashedExtLen = int(buf[pos])
		pos++
		copy(cell.hashedExtension[:], buf[pos:pos+cell.hashedExtLen])
		pos += cell.hashedExtLen
	}
	if flags&cellFlagExtension != 0 {
		cell.extLen = int(buf[pos])
		pos++
		copy(cell.extension[:], buf[pos:pos+cell.extLen])
		pos += cell.extLen //nolint
	}
	if flags&cellFlagDelete != 0 {
		log.Warn("deleted cell should not be encoded", "cell", cell.String())
		cell.Update.Flags = DeleteUpdate
	}
	return nil
}

// Encode current state of hph into bytes
func (hph *HexPatriciaHashed) EncodeCurrentState(buf []byte) ([]byte, error) {
	s := state{
		RootChecked: hph.rootChecked,
		RootTouched: hph.rootTouched,
		RootPresent: hph.rootPresent,
	}
	if hph.currentKeyLen > 0 {
		panic("currentKeyLen > 0")
	}

	s.Root = hph.root.Encode()
	copy(s.Depths[:], hph.depths[:])
	copy(s.BranchBefore[:], hph.branchBefore[:])
	copy(s.TouchMap[:], hph.touchMap[:])
	copy(s.AfterMap[:], hph.afterMap[:])

	return s.Encode(buf)
}

// buf expected to be encoded hph state. Decode state and set up hph to that state.
func (hph *HexPatriciaHashed) SetState(buf []byte) error {
	hph.Reset()

	if buf == nil {
		// reset state to 'empty'
		hph.currentKeyLen = 0
		hph.rootChecked = false
		hph.rootTouched = false
		hph.rootPresent = false
		hph.activeRows = 0

		for i := 0; i < len(hph.depths); i++ {
			hph.depths[i] = 0
			hph.branchBefore[i] = false
			hph.touchMap[i] = 0
			hph.afterMap[i] = 0
		}
		return nil
	}
	if hph.activeRows != 0 {
		return errors.New("target trie has active rows, could not reset state before fold")
	}

	var s state
	if err := s.Decode(buf); err != nil {
		return err
	}

	if err := hph.root.Decode(s.Root); err != nil {
		return err
	}
	hph.rootChecked = s.RootChecked
	hph.rootTouched = s.RootTouched
	hph.rootPresent = s.RootPresent

	copy(hph.depths[:], s.Depths[:])
	copy(hph.branchBefore[:], s.BranchBefore[:])
	copy(hph.touchMap[:], s.TouchMap[:])
	copy(hph.afterMap[:], s.AfterMap[:])

	if hph.root.accountAddrLen > 0 {
		if hph.Ctx == nil {
			panic("nil ctx")
		}

		update, err := hph.Ctx.Account(hph.root.accountAddr[:hph.root.accountAddrLen])
		if err != nil {
			return err
		}
		hph.root.setFromUpdate(update)
	}
	if hph.root.storageAddrLen > 0 {
		if hph.Ctx == nil {
			panic("nil ctx")
		}
		update, err := hph.Ctx.Storage(hph.root.storageAddr[:hph.root.storageAddrLen])
		if err != nil {
			return err
		}
		hph.root.setFromUpdate(update)
		//hph.root.deriveHashedKeys(0, hph.keccak, hph.accountKeyLen)
	}

	return nil
}

//func bytesToUint64(buf []byte) (x uint64) {
//	for i, b := range buf {
//		x = x<<8 + uint64(b)
//		if i == 7 {
//			return
//		}
//	}
//	return
//}

func hexToCompact(key []byte) []byte {
	zeroByte, keyPos, keyLen := makeCompactZeroByte(key)
	bufLen := keyLen/2 + 1 // always > 0
	buf := make([]byte, bufLen)
	buf[0] = zeroByte
	return decodeKey(key[keyPos:], buf)
}

func makeCompactZeroByte(key []byte) (compactZeroByte byte, keyPos, keyLen int) {
	keyLen = len(key)
	if hasTerm(key) {
		keyLen--
		compactZeroByte = 0x20
	}
	var firstNibble byte
	if len(key) > 0 {
		firstNibble = key[0]
	}
	if keyLen&1 == 1 {
		compactZeroByte |= 0x10 | firstNibble // Odd: (1<<4) + first nibble
		keyPos++
	}

	return
}

func decodeKey(key, buf []byte) []byte {
	keyLen := len(key)
	if hasTerm(key) {
		keyLen--
	}
	for keyIndex, bufIndex := 0, 1; keyIndex < keyLen; keyIndex, bufIndex = keyIndex+2, bufIndex+1 {
		if keyIndex == keyLen-1 {
			buf[bufIndex] = buf[bufIndex] & 0x0f
		} else {
			buf[bufIndex] = key[keyIndex+1]
		}
		buf[bufIndex] |= key[keyIndex] << 4
	}
	return buf
}

func CompactedKeyToHex(compact []byte) []byte {
	if len(compact) == 0 {
		return compact
	}
	base := keybytesToHexNibbles(compact)
	// delete terminator flag
	if base[0] < 2 {
		base = base[:len(base)-1]
	}
	// apply odd flag
	chop := 2 - base[0]&1
	return base[chop:]
}

func keybytesToHexNibbles(str []byte) []byte {
	l := len(str)*2 + 1
	var nibbles = make([]byte, l)
	for i, b := range str {
		nibbles[i*2] = b / 16
		nibbles[i*2+1] = b % 16
	}
	nibbles[l-1] = 16
	return nibbles
}

// hasTerm returns whether a hex key has the terminator flag.
func hasTerm(s []byte) bool {
	return len(s) > 0 && s[len(s)-1] == 16
}

func commonPrefixLen(b1, b2 []byte) int {
	var i int
	for i = 0; i < len(b1) && i < len(b2); i++ {
		if b1[i] != b2[i] {
			break
		}
	}
	return i
}

// nolint
// Hashes provided key and expands resulting hash into nibbles (each byte split into two nibbles by 4 bits)
func (hph *HexPatriciaHashed) HashAndNibblizeKey(key []byte) []byte {
	hashedKey := make([]byte, length.Hash)

	hph.keccak.Reset()
	fp := length.Addr
	if len(key) < length.Addr {
		fp = len(key)
	}
	hph.keccak.Write(key[:fp])
	hph.keccak.Read(hashedKey[:length.Hash])

	if len(key[fp:]) > 0 {
		hashedKey = append(hashedKey, make([]byte, length.Hash)...)
		hph.keccak.Reset()
		hph.keccak.Write(key[fp:])
		hph.keccak.Read(hashedKey[length.Hash:])
	}

	nibblized := make([]byte, len(hashedKey)*2)
	for i, b := range hashedKey {
		nibblized[i*2] = (b >> 4) & 0xf
		nibblized[i*2+1] = b & 0xf
	}
	return nibblized
}

func nibblize(key []byte) []byte {
	nibblized := make([]byte, len(key)*2)
	for i, b := range key {
		nibblized[i*2] = (b >> 4) & 0xf
		nibblized[i*2+1] = b & 0xf
	}
	return nibblized
}

// compactKey takes a slice of nibbles and compacts them into the original byte slice.
// It returns an error if the input contains invalid nibbles (values > 0xF).
func compactKey(nibbles []byte) ([]byte, error) {
	// If the number of nibbles is odd, you might decide to handle it differently.
	// For this example, we'll return an error.
	if len(nibbles)%2 != 0 {
		return nil, errors.New("nibbles slice has an odd length")
	}

	key := make([]byte, len(nibbles)/2)
	for i := 0; i < len(key); i++ {
		highNibble := nibbles[i*2]
		lowNibble := nibbles[i*2+1]

		// Validate that each nibble is indeed a nibble
		if highNibble > 0xF || lowNibble > 0xF {
			return nil, fmt.Errorf("invalid nibble at position %d or %d: 0x%X, 0x%X", i*2, i*2+1, highNibble, lowNibble)
		}

		key[i] = (highNibble << 4) | (lowNibble & 0x0F)
	}
	return key, nil
}

func (hph *HexPatriciaHashed) Grid() [128][16]cell {
	return hph.grid
}

func (hph *HexPatriciaHashed) PrintAccountsInGrid() {
	fmt.Printf("SEARCHING FOR ACCOUNTS IN GRID\n")
	for row := 0; row < 128; row++ {
		for col := 0; col < 16; col++ {
			c := hph.grid[row][col]
			if c.accountAddr[19] != 0 && c.accountAddr[0] != 0 {
				fmt.Printf("FOUND account %x in position (%d,%d)\n", c.accountAddr, row, col)
			}
		}
	}
}
