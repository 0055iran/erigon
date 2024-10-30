// Copyright 2024 The Erigon Authors
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

package merkle_tree_test

import (
	_ "embed"
	"testing"

	"github.com/erigontech/erigon-lib/common"
	"github.com/erigontech/erigon/v3/cl/clparams"
	"github.com/erigontech/erigon/v3/cl/merkle_tree"
	"github.com/erigontech/erigon/v3/cl/phase1/core/state"
	"github.com/erigontech/erigon/v3/cl/utils"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/serialized.ssz_snappy
var beaconState []byte

func TestHashTreeRoot(t *testing.T) {
	bs := state.New(&clparams.MainnetBeaconConfig)
	require.NoError(t, utils.DecodeSSZSnappy(bs, beaconState, int(clparams.DenebVersion)))
	root, err := bs.HashSSZ()
	require.NoError(t, err)
	require.Equal(t, common.Hash(root), common.HexToHash("0x9f684cf34c4ac8eb9056051f93498c552b59de6b0977c453ee099be68e58d90c"))
}

func TestHashTreeRootTxs(t *testing.T) {
	txs := [][]byte{
		{1, 2, 3},
		{1, 2, 3},
		{1, 2, 3},
	}
	root, err := merkle_tree.TransactionsListRoot(txs)
	require.NoError(t, err)
	require.Equal(t, common.Hash(root), common.HexToHash("0x987269bc1075122edff32bfc38479757103cee5c1ed6e990de7ffee85b5dd18a"))
}
