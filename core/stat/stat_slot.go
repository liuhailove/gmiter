package stat

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	metric_exporter "git.garena.com/honggang.liu/seamiter-go/exporter/metric"
	"git.garena.com/honggang.liu/seamiter-go/util"
)

const (
	StatSlotOrder = 1000
	ResultPass    = "pass"
	ResultBlock   = "block"
)

var (
	DefaultSlot = &Slot{}

	handledCounter = metric_exporter.NewCounter(
		"handled_total",
		"Total handled count",
		[]string{"resource", "result", "block_type"})
)

func init() {
	metric_exporter.Register(handledCounter)
}

type Slot struct {
	isInitialized util.AtomicBool
}

func (s Slot) Order() uint32 {
	return StatSlotOrder
}

// Initial
//
// 初始化，如果有初始化工作放入其中
func (s *Slot) Initial() {}

func (s Slot) OnEntryPassed(ctx *base.EntryContext) {
	s.recordPassFor(ctx.StatNode, ctx.Input.BatchCount)
	if ctx.Resource.FlowType() == base.Inbound {
		s.recordPassFor(InboundNode(), ctx.Input.BatchCount)
	}
	if handledCounter != nil {
		handledCounter.Add(float64(ctx.Input.BatchCount), ctx.Resource.Name(), ResultPass, "")
	}
}

func (s Slot) OnEntryBlocked(ctx *base.EntryContext, blockError *base.BlockError) {
	s.recordBlockFor(ctx.StatNode, ctx.Input.BatchCount)
	if blockError != nil {
		s.recordBlockForType(ctx.StatNode, blockError.BlockType(), ctx.Input.BatchCount)
	}
	if ctx.Resource.FlowType() == base.Inbound {
		s.recordBlockFor(InboundNode(), ctx.Input.BatchCount)
	}
	if handledCounter != nil {
		handledCounter.Add(float64(ctx.Input.BatchCount), ctx.Resource.Name(), ResultBlock, blockError.BlockType().String())
	}
}

func (s Slot) OnCompleted(ctx *base.EntryContext) {
	rt := util.CurrentTimeMillis() - ctx.StartTime()
	ctx.PutRt(rt)
	s.recordCompleteFor(ctx.StatNode, ctx.Input.BatchCount, rt, ctx.Err())
	if ctx.Resource.FlowType() == base.Inbound {
		s.recordCompleteFor(InboundNode(), ctx.Input.BatchCount, rt, ctx.Err())
	}
}

func (s *Slot) recordPassFor(sn base.StatNode, count uint32) {
	if sn == nil {
		return
	}
	sn.IncreaseConcurrency()
	sn.AddCount(base.MetricEventPass, int64(count))
	//sn.AddCountWithTime(startTime, base.MetricEventPass, int64(count))
}

func (s *Slot) recordBlockFor(sn base.StatNode, count uint32) {
	if sn == nil {
		return
	}
	sn.AddCount(base.MetricEventBlock, int64(count))
	//sn.AddCountWithTime(startTime, base.MetricEventBlock, int64(count))
}

// 记录Block类型
func (s *Slot) recordBlockForType(sn base.StatNode, blockType base.BlockType, count uint32) {
	if sn == nil {
		return
	}
	if base.BlockTypeFlow == blockType {
		sn.AddCount(base.MetricEventBlockFlow, int64(count))
		//sn.AddCountWithTime(startTime, base.MetricEventBlockFlow, int64(count))
	} else if base.BlockTypeIsolation == blockType {
		sn.AddCount(base.MetricEventBlockIsolation, int64(count))
		//sn.AddCountWithTime(startTime, base.MetricEventBlockIsolation, int64(count))
	} else if base.BlockTypeCircuitBreaking == blockType {
		sn.AddCount(base.MetricEventBlockCircuitBreaking, int64(count))
		//sn.AddCountWithTime(startTime, base.MetricEventBlockCircuitBreaking, int64(count))
	} else if base.BlockTypeSystemFlow == blockType {
		sn.AddCount(base.MetricEventBlockSystem, int64(count))
		//sn.AddCountWithTime(startTime, base.MetricEventBlockSystem, int64(count))
	} else if base.BlockTypeHotSpotParamFlow == blockType {
		sn.AddCount(base.MetricEventBlockHotSpotParamFlow, int64(count))
		//sn.AddCountWithTime(startTime, base.MetricEventBlockHotSpotParamFlow, int64(count))
	} else if base.BlockTypeMock == blockType || base.BlockTypeMockError == blockType || base.BlockTypeMockRequest == blockType || base.BlockTypeMockCtxTimeout == blockType {
		sn.AddCount(base.MetricEventBlockMock, int64(count))
		//sn.AddCountWithTime(startTime, base.MetricEventBlockMock, int64(count))
	}
	return
}

func (s *Slot) recordCompleteFor(sn base.StatNode, count uint32, rt uint64, err error) {
	if sn == nil {
		return
	}
	if err != nil {
		sn.AddCount(base.MetricEventError, int64(count))
	}
	sn.AddCount(base.MetricEventRt, int64(rt))
	sn.AddCount(base.MetricEventComplete, int64(count))
	sn.DecreaseConcurrency()
}
