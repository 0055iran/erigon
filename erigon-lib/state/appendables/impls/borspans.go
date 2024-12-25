package impls

import (
	"github.com/erigontech/erigon-lib/common/hexutility"
	"github.com/erigontech/erigon-lib/kv"
	"github.com/erigontech/erigon-lib/kv/stream"
	ca "github.com/erigontech/erigon-lib/state/appendables"
	"github.com/erigontech/erigon/polygon/heimdall"
	"github.com/tidwall/btree"
)

// appendables which don't store non-canonical data
// 1. or which tsId = tsNum always
// 2. bor spans/milestones/checkpoints
// 3. two level lookup: stepKey -> tsId/tsNum -> value
/// forget about blobs right now

// 1. a valsTable stores the value simply

const (
	BorSpans ca.ApEnum = "borspans.appe"
)


type SpanAppendable struct {
	ca.BaseAppendable[[]byte, []byte]
	valsTable string
}

func NewSpanAppendable(valsTable string) *SpanAppendable {
	ap := &SpanAppendable{
		BaseAppendable: ca.BaseAppendable[[]byte, []byte]{},
		valsTable:      valsTable,
	}

	gen := &SpanSourceKeyGenerator{}
	ap.SetSourceKeyGenerator(gen)
	ap.SetValueFetcher(ca.NewPlainFetcher(valsTable))
	ap.SetValuePutter(ca.NewPlainPutter(valsTable))
	ap.SetFreezer(ca.NewPlainFreezer(valsTable, gen))

	salt := uint32(4343) // load from salt-blocks.txt etc.

	indexb := ca.NewSimpleAccessorBuilder(ca.NewAccessorArgs(true, false, false, salt), ca.BorSpans)

	ap.rosnapshot = &ca.RoSnapshots{
		enums: []ApEnum{BorSpans},
		dirty: map[ApEnum]*btree.BTreeG[*DirtySegment]{},
		visible: map[ApEnum]VisibleSegments{
			BorSpans: {},
		},
	}
	ap.enum = BorSpans
	return ap
}

type SpanSourceKeyGenerator struct{}

func (s *SpanSourceKeyGenerator) FromStepKey(stepKeyFrom, stepKeyTo uint64, tx kv.Tx) stream.Uno[[]byte] {
	spanFrom := heimdall.SpanIdAt(stepKeyFrom)
	spanTo := heimdall.SpanIdAt(stepKeyTo)
	return ca.NewSequentialStream(uint64(spanFrom), uint64(spanTo))
}

func (s *SpanSourceKeyGenerator) FromTsNum(tsNum uint64, tx kv.Tx) []byte {
	return hexutility.EncodeTs(tsNum)
}

func (s *SpanSourceKeyGenerator) FromTsId(tsId uint64, forkId []byte, tx kv.Tx) []byte {
	return hexutility.EncodeTs(tsId)
}

// two things to set
// 1. source key generator
// 2. Can freeze
