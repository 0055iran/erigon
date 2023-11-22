// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/holiman/uint256"

	libcommon "github.com/ledgerwatch/erigon-lib/common"

	"github.com/ledgerwatch/erigon/consensus/misc"
	"github.com/ledgerwatch/erigon/params"
)

var activators = map[int]func(*JumpTable){
	7516: enable7516,
	6780: enable6780,
	5656: enable5656,
	4844: enable4844,
	3860: enable3860,
	3855: enable3855,
	3529: enable3529,
	3198: enable3198,
	2929: enable2929,
	2200: enable2200,
	1884: enable1884,
	1344: enable1344,
	1153: enable1153,
}

// EnableEIP enables the given EIP on the config.
// This operation writes in-place, and callers need to ensure that the globally
// defined jump tables are not polluted.
func EnableEIP(eipNum int, jt *JumpTable) error {
	enablerFn, ok := activators[eipNum]
	if !ok {
		return fmt.Errorf("undefined eip %d", eipNum)
	}
	enablerFn(jt)
	validateAndFillMaxStack(jt)
	return nil
}

func ValidEip(eipNum int) bool {
	_, ok := activators[eipNum]
	return ok
}
func ActivateableEips() []string {
	var nums []string //nolint:prealloc
	for k := range activators {
		nums = append(nums, fmt.Sprintf("%d", k))
	}
	sort.Strings(nums)
	return nums
}

// enable1884 applies EIP-1884 to the given jump table:
// - Increase cost of BALANCE to 700
// - Increase cost of EXTCODEHASH to 700
// - Increase cost of SLOAD to 800
// - Define SELFBALANCE, with cost GasFastStep (5)
func enable1884(jt *JumpTable) {
	// Gas cost changes
	jt[SLOAD].constantGas = params.SloadGasEIP1884
	jt[BALANCE].constantGas = params.BalanceGasEIP1884
	jt[EXTCODEHASH].constantGas = params.ExtcodeHashGasEIP1884

	// New opcode
	jt[SELFBALANCE] = &operation{
		execute:     opSelfBalance,
		constantGas: GasFastStep,
		numPop:      0,
		numPush:     1,
	}
}

func opSelfBalance(pc *uint64, interpreter *EVMInterpreter, callContext *ScopeContext) ([]byte, error) {
	balance := interpreter.evm.IntraBlockState().GetBalance(callContext.Contract.Address())
	callContext.Stack.Push(balance)
	return nil, nil
}

// enable1344 applies EIP-1344 (ChainID Opcode)
// - Adds an opcode that returns the current chain’s EIP-155 unique identifier
func enable1344(jt *JumpTable) {
	// New opcode
	jt[CHAINID] = &operation{
		execute:     opChainID,
		constantGas: GasQuickStep,
		numPop:      0,
		numPush:     1,
	}
}

// opChainID implements CHAINID opcode
func opChainID(pc *uint64, interpreter *EVMInterpreter, callContext *ScopeContext) ([]byte, error) {
	chainId, _ := uint256.FromBig(interpreter.evm.ChainRules().ChainID)
	callContext.Stack.Push(chainId)
	return nil, nil
}

// enable2200 applies EIP-2200 (Rebalance net-metered SSTORE)
func enable2200(jt *JumpTable) {
	jt[SLOAD].constantGas = params.SloadGasEIP2200
	jt[SSTORE].dynamicGas = gasSStoreEIP2200
}

// enable2929 enables "EIP-2929: Gas cost increases for state access opcodes"
// https://eips.ethereum.org/EIPS/eip-2929
func enable2929(jt *JumpTable) {
	jt[SSTORE].dynamicGas = gasSStoreEIP2929

	jt[SLOAD].constantGas = 0
	jt[SLOAD].dynamicGas = gasSLoadEIP2929

	jt[EXTCODECOPY].constantGas = params.WarmStorageReadCostEIP2929
	jt[EXTCODECOPY].dynamicGas = gasExtCodeCopyEIP2929

	jt[EXTCODESIZE].constantGas = params.WarmStorageReadCostEIP2929
	jt[EXTCODESIZE].dynamicGas = gasEip2929AccountCheck

	jt[EXTCODEHASH].constantGas = params.WarmStorageReadCostEIP2929
	jt[EXTCODEHASH].dynamicGas = gasEip2929AccountCheck

	jt[BALANCE].constantGas = params.WarmStorageReadCostEIP2929
	jt[BALANCE].dynamicGas = gasEip2929AccountCheck

	jt[CALL].constantGas = params.WarmStorageReadCostEIP2929
	jt[CALL].dynamicGas = gasCallEIP2929

	jt[CALLCODE].constantGas = params.WarmStorageReadCostEIP2929
	jt[CALLCODE].dynamicGas = gasCallCodeEIP2929

	jt[STATICCALL].constantGas = params.WarmStorageReadCostEIP2929
	jt[STATICCALL].dynamicGas = gasStaticCallEIP2929

	jt[DELEGATECALL].constantGas = params.WarmStorageReadCostEIP2929
	jt[DELEGATECALL].dynamicGas = gasDelegateCallEIP2929

	// This was previously part of the dynamic cost, but we're using it as a constantGas
	// factor here
	jt[SELFDESTRUCT].constantGas = params.SelfdestructGasEIP150
	jt[SELFDESTRUCT].dynamicGas = gasSelfdestructEIP2929
}

func enable3529(jt *JumpTable) {
	jt[SSTORE].dynamicGas = gasSStoreEIP3529
	jt[SELFDESTRUCT].dynamicGas = gasSelfdestructEIP3529
}

// enable3198 applies EIP-3198 (BASEFEE Opcode)
// - Adds an opcode that returns the current block's base fee.
func enable3198(jt *JumpTable) {
	// New opcode
	jt[BASEFEE] = &operation{
		execute:     opBaseFee,
		constantGas: GasQuickStep,
		numPop:      0,
		numPush:     1,
	}
}

// enable1153 applies EIP-1153 "Transient Storage"
// - Adds TLOAD that reads from transient storage
// - Adds TSTORE that writes to transient storage
func enable1153(jt *JumpTable) {
	jt[TLOAD] = &operation{
		execute:     opTload,
		constantGas: params.WarmStorageReadCostEIP2929,
		numPop:      1,
		numPush:     1,
	}

	jt[TSTORE] = &operation{
		execute:     opTstore,
		constantGas: params.WarmStorageReadCostEIP2929,
		numPop:      2,
		numPush:     0,
	}
}

// opTload implements TLOAD opcode
func opTload(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	loc := scope.Stack.Peek()
	hash := libcommon.Hash(loc.Bytes32())
	val := interpreter.evm.IntraBlockState().GetTransientState(scope.Contract.Address(), hash)
	loc.SetBytes(val.Bytes())
	return nil, nil
}

// opTstore implements TSTORE opcode
func opTstore(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	if interpreter.readOnly {
		return nil, ErrWriteProtection
	}
	loc := scope.Stack.Pop()
	val := scope.Stack.Pop()
	interpreter.evm.IntraBlockState().SetTransientState(scope.Contract.Address(), loc.Bytes32(), val)
	return nil, nil
}

// opBaseFee implements BASEFEE opcode
func opBaseFee(pc *uint64, interpreter *EVMInterpreter, callContext *ScopeContext) ([]byte, error) {
	baseFee := interpreter.evm.Context().BaseFee
	callContext.Stack.Push(baseFee)
	return nil, nil
}

// enable3855 applies EIP-3855 (PUSH0 opcode)
func enable3855(jt *JumpTable) {
	// New opcode
	jt[PUSH0] = &operation{
		execute:     opPush0,
		constantGas: GasQuickStep,
		numPop:      0,
		numPush:     1,
	}
}

// opPush0 implements the PUSH0 opcode
func opPush0(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	scope.Stack.Push(new(uint256.Int))
	return nil, nil
}

// EIP-3860: Limit and meter initcode
// https://eips.ethereum.org/EIPS/eip-3860
func enable3860(jt *JumpTable) {
	jt[CREATE].dynamicGas = gasCreateEip3860
	jt[CREATE2].dynamicGas = gasCreate2Eip3860
}

// enable4844 applies mini-danksharding (BLOBHASH opcode)
// - Adds an opcode that returns the versioned blob hash of the tx at a index.
func enable4844(jt *JumpTable) {
	jt[BLOBHASH] = &operation{
		execute:     opBlobHash,
		constantGas: GasFastestStep,
		numPop:      1,
		numPush:     1,
	}
}

// opBlobHash implements the BLOBHASH opcode
func opBlobHash(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	idx := scope.Stack.Peek()
	if idx.LtUint64(uint64(len(interpreter.evm.TxContext().BlobHashes))) {
		hash := interpreter.evm.TxContext().BlobHashes[idx.Uint64()]
		idx.SetBytes(hash.Bytes())
	} else {
		idx.Clear()
	}
	return nil, nil
}

// enable5656 enables EIP-5656 (MCOPY opcode)
// https://eips.ethereum.org/EIPS/eip-5656
func enable5656(jt *JumpTable) {
	jt[MCOPY] = &operation{
		execute:     opMcopy,
		constantGas: GasFastestStep,
		dynamicGas:  gasMcopy,
		numPop:      3,
		numPush:     0,
		memorySize:  memoryMcopy,
	}
}

// opMcopy implements the MCOPY opcode (https://eips.ethereum.org/EIPS/eip-5656)
func opMcopy(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	var (
		dst    = scope.Stack.Pop()
		src    = scope.Stack.Pop()
		length = scope.Stack.Pop()
	)
	// These values are checked for overflow during memory expansion calculation
	// (the memorySize function on the opcode).
	scope.Memory.Copy(dst.Uint64(), src.Uint64(), length.Uint64())
	return nil, nil
}

// enable6780 applies EIP-6780 (deactivate SELFDESTRUCT)
func enable6780(jt *JumpTable) {
	jt[SELFDESTRUCT] = &operation{
		execute:     opSelfdestruct6780,
		dynamicGas:  gasSelfdestructEIP3529,
		constantGas: params.SelfdestructGasEIP150,
		numPop:      1,
		numPush:     0,
	}
}

// opBlobBaseFee implements the BLOBBASEFEE opcode
func opBlobBaseFee(pc *uint64, interpreter *EVMInterpreter, callContext *ScopeContext) ([]byte, error) {
	excessBlobGas := interpreter.evm.Context().ExcessBlobGas
	blobBaseFee, err := misc.GetBlobGasPrice(interpreter.evm.ChainConfig(), *excessBlobGas)
	if err != nil {
		return nil, err
	}
	callContext.Stack.Push(blobBaseFee)
	return nil, nil
}

// enable7516 applies EIP-7516 (BLOBBASEFEE opcode)
// - Adds an opcode that returns the current block's blob base fee.
func enable7516(jt *JumpTable) {
	jt[BLOBBASEFEE] = &operation{
		execute:     opBlobBaseFee,
		constantGas: GasQuickStep,
		numPop:      0,
		numPush:     1,
	}
}

// enableEOF applies the EOF changes.
func enableEOF(jt *JumpTable) {
	// Deprecate opcodes
	undefined := &operation{
		execute:     opUndefined,
		constantGas: 0,
		numPop:      0,
		numPush:     0,
		undefined:   true,
	}
	jt[CALLCODE] = undefined
	jt[SELFDESTRUCT] = undefined
	jt[JUMP] = undefined
	jt[JUMPI] = undefined
	jt[PC] = undefined

	// TODO(racytech): Make sure everything is correct
	// New opcodes
	jt[RJUMP] = &operation{
		execute:     opRjump,
		constantGas: GasQuickStep,
		numPop:      0,
		numPush:     0,
		terminal:    true,
	}
	jt[RJUMPI] = &operation{
		execute:     opRjumpi,
		constantGas: GasSwiftStep,
		numPop:      1,
		numPush:     0,
	}
	jt[RJUMPV] = &operation{
		execute:     opRjumpv,
		constantGas: GasSwiftStep,
		numPop:      1,
		numPush:     0,
	}
	jt[CALLF] = &operation{
		execute:     opCallf,
		constantGas: GasFastStep,
		numPop:      0,
		numPush:     0,
	}
	jt[RETF] = &operation{
		execute:     opRetf,
		constantGas: GasSwiftStep,
		numPop:      0,
		numPush:     0,
		terminal:    true,
	}
	jt[JUMPF] = &operation{
		execute:     opJumpf,
		constantGas: GasFastStep,
		numPop:      0,
		numPush:     0,
	}
	jt[DUPN] = &operation{
		execute:     opDupN,
		constantGas: GasFastestStep,
		numPop:      0,
		numPush:     1,
	}
	jt[SWAPN] = &operation{
		execute:     opSwapN,
		constantGas: GasFastestStep,
		numPop:      0,
		numPush:     0,
	}
	jt[DATALOAD] = &operation{
		execute:     opDataLoad,
		constantGas: GasSwiftStep,
		numPop:      1,
		numPush:     1,
	}
	jt[DATALOADN] = &operation{
		execute:     opDataLoad,
		constantGas: GasFastestStep,
		numPop:      0,
		numPush:     1,
	}
	jt[DATASIZE] = &operation{
		execute:     opDataSize,
		constantGas: GasQuickStep,
		numPop:      0,
		numPush:     1,
	}
	jt[DATACOPY] = &operation{
		execute:     opDataCopy,
		constantGas: GasFastestStep,
		dynamicGas:  gasDataCopyEIP7480,
		numPop:      3,
		numPush:     0,
		memorySize:  memoryDataCopy,
	}
	jt[CREATE3] = &operation{
		execute:     opCreate3,
		constantGas: params.Create3Gas,
		// dynamicGas:  gasCreate2,
		numPop:     4,
		numPush:    1,
		memorySize: memoryCreate2,
	}
	jt[CREATE4] = &operation{
		execute:     opCreate4,
		constantGas: params.Create4Gas,
		// dynamicGas:  gasCreate2,
		numPop:     4,
		numPush:    1,
		memorySize: memoryCreate2,
	}
	jt[RETURNCONTRACT] = &operation{}
}

// opRjump implements the rjump opcode.
func opRjump(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	var (
		code   = scope.Contract.CodeAt(scope.CodeSection)
		offset = parseInt16(code[*pc+1:])
	)
	// move pc past op and operand (+3), add relative offset, subtract 1 to
	// account for interpreter loop.
	*pc = uint64(int64(*pc+3) + int64(offset) - 1)
	return nil, nil
}

// opRjumpi implements the RJUMPI opcode
func opRjumpi(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	condition := scope.Stack.Pop()
	if condition.BitLen() == 0 {
		// Not branching, just skip over immediate argument.
		*pc += 2
		return nil, nil
	}
	return opRjump(pc, interpreter, scope)
}

// opRjumpv implements the RJUMPV opcode
func opRjumpv(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	var (
		code  = scope.Contract.CodeAt(scope.CodeSection)
		count = uint64(code[*pc+1])
		idx   = scope.Stack.Pop()
	)
	if idx, overflow := idx.Uint64WithOverflow(); overflow || idx >= count {
		// Index out-of-bounds, don't branch, just skip over immediate
		// argument.
		*pc += 1 + count*2
		return nil, nil
	}
	offset := parseInt16(code[*pc+2+2*idx.Uint64():])
	*pc = uint64(int64(*pc+2+count*2) + int64(offset) - 1)
	return nil, nil
}

// opCallf implements the CALLF opcode
func opCallf(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	var (
		code = scope.Contract.CodeAt(scope.CodeSection)
		idx  = binary.BigEndian.Uint16(code[*pc+1:])
		typ  = scope.Contract.Container.Types[scope.CodeSection]
	)
	if scope.Stack.Len()+int(typ.MaxStackHeight) >= 1024 {
		return nil, fmt.Errorf("stack overflow")
	}
	retCtx := &ReturnContext{
		Section:     scope.CodeSection,
		Pc:          *pc + 3,
		StackHeight: scope.Stack.Len() - int(typ.Input),
	}
	scope.ReturnStack = append(scope.ReturnStack, retCtx)
	scope.CodeSection = uint64(idx)
	*pc = 0
	// *pc -= 1 // hacks xD
	return nil, nil
}

// opRetf implements the RETF opcode
func opRetf(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	var (
		last   = len(scope.ReturnStack) - 1
		retCtx = scope.ReturnStack[last]
	)
	scope.ReturnStack = scope.ReturnStack[:last]
	scope.CodeSection = retCtx.Section
	*pc = retCtx.Pc - 1

	// If returning from top frame, exit cleanly.
	if len(scope.ReturnStack) == 0 {
		return nil, errStopToken
	}
	return nil, nil
}

func opJumpf(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	var (
		code    = scope.Contract.CodeAt(scope.CodeSection)
		section = binary.BigEndian.Uint16(code[*pc+1:])
		typ     = scope.Contract.Container.Types[scope.CodeSection]
	)
	if scope.Stack.Len()+int(typ.MaxStackHeight) >= 1024 {
		return nil, fmt.Errorf("stack overflow")
	}
	scope.CodeSection = uint64(section)
	*pc = 0
	return nil, nil
}

func opDupN(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	// TODO(racytech): not yet merged
	return nil, nil
}

func opSwapN(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	// TODO(racytech): not yet merged
	return nil, nil
}

func opDataLoad(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	var (
		index  = scope.Stack.Pop()
		data   = scope.Contract.Data()
		offset = int(index.Uint64()) // with overflow maybe?
	)
	if len(data) < 32 || len(data)-32 < offset {
		return nil, ErrInvalidMemoryAccess
	}
	val := new(uint256.Int).SetBytes(data[offset : offset+32])
	scope.Stack.Push(val)
	return nil, nil
}

func opDataLoadN(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	var (
		code   = scope.Contract.CodeAt(scope.CodeSection)
		data   = scope.Contract.Data()
		offset = int(binary.BigEndian.Uint16(code[*pc+1:]))
	)
	if len(data) < 32 || len(data)-32 < offset {
		return nil, ErrInvalidMemoryAccess
	}
	val := new(uint256.Int).SetBytes(data[offset : offset+32])
	scope.Stack.Push(val)
	return nil, nil
}

func opDataSize(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	dataSize := len(scope.Contract.Data())
	val := new(uint256.Int).SetUint64(uint64(dataSize))
	scope.Stack.Push(val)
	return nil, nil
}

func opDataCopy(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	var (
		memOffset256 = scope.Stack.Pop()
		dataIndex256 = scope.Stack.Pop()
		size256      = scope.Stack.Pop()

		data    = scope.Contract.Data()
		dataLen = uint64(len(data))
		src     = dataIndex256.Uint64()
		dst     = memOffset256.Uint64()
		size    = size256.Uint64()
	)

	if dataLen < size || dataLen-size < src {
		return nil, ErrInvalidMemoryAccess
	}

	if size > 0 {
		scope.Memory.Copy(dst, src, size)
	}

	return nil, nil
}

func opCreate3(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	if interpreter.readOnly {
		return nil, ErrWriteProtection
	}
	var (
		code             = scope.Contract.CodeAt(scope.CodeSection)
		initContainerIdx = int(code[*pc+1])
		endowment        = scope.Stack.Pop()
		salt             = scope.Stack.Pop()
		offset, size     = scope.Stack.Pop(), scope.Stack.Pop()
		inputOffset      = offset.Uint64()
		inputSize        = size.Uint64()
		gas              = scope.Contract.Gas
		input            = []byte{}
		initContainer    = scope.Contract.Container.SubContainer[initContainerIdx]
	)
	*pc += 2

	if inputSize > 0 {
		input = scope.Memory.GetCopy(int64(inputOffset), int64(inputSize))
	}
	// Apply EIP150
	gas -= gas / 64
	scope.Contract.UseGas(gas)

	stackValue := size

	res, addr, returnGas, suberr := interpreter.evm.Create3(scope.Contract, input, initContainer, gas, &endowment, &salt)

	// Push item on the stack based on the returned error.
	if suberr != nil {
		stackValue.Clear()
	} else {
		stackValue.SetBytes(addr.Bytes())
	}

	scope.Stack.Push(&stackValue)
	scope.Contract.Gas += returnGas

	if suberr == ErrExecutionReverted {
		interpreter.returnData = res // set REVERT data to return data buffer
		return res, nil
	}
	interpreter.returnData = nil // clear dirty return data buffer
	return nil, nil
}

func opCreate4(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	// TODO(racytech): Add new TxType
	// CREATE4 expects new transaction type = 4 which will carry initcodes
	return nil, nil
}

func opReturnContract(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	var (
		code               = scope.Contract.CodeAt(scope.CodeSection)
		deployContainerIdx = int(code[*pc+1])
		auxDataOffset      = scope.Stack.Pop()
		auxDataSize        = scope.Stack.Pop()
		deployContainer    = scope.Contract.Container.SubContainer[deployContainerIdx]
		auxData            = scope.Memory.GetPtr(int64(auxDataOffset.Uint64()), int64(auxDataSize.Uint64()))
	)
	var c Container
	// NOTE(racytech): UnmarshalBinary checks for correct EOF format
	// but it decodes entire container, which is a bit expensive. Do we need that?
	// Can we do better?
	if err := c.UnmarshalBinary(deployContainer); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidEOFInitcode, err)
	}
	// TODO(racytech): make sure this one refers to the same underlying slice as Container.SubContainer[deployContainerIdx]
	deployContainer = append(deployContainer, auxData...)
	return nil, nil
}
