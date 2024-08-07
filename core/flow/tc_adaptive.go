package flow

import (
	"git.garena.com/honggang.liu/seamiter-go/core/system_metric"
	"git.garena.com/honggang.liu/seamiter-go/logging"
)

// MemoryAdaptiveTrafficShapingCalculator is a memory adaptive traffic shaping calculator
//
// adaptive flow control algorithm
// If the watermark is less than Rule.MemLowWaterMarkBytes, the threshold is Rule.LowMemUsageThreshold.
// If the watermark is greater than Rule.MemHighWaterMarkBytes, the threshold is Rule.HighMemUsageThreshold.
// Otherwise, the threshold is ((watermark - MemLowWaterMarkBytes)/(MemHighWaterMarkBytes - MemLowWaterMarkBytes)) *
//	(HighMemUsageThreshold - LowMemUsageThreshold) + LowMemUsageThreshold.
type MemoryAdaptiveTrafficShapingCalculator struct {
	owner                 *TrafficShapingController
	lowMemUsageThreshold  float64
	highMemUsageThreshold float64
	memLowWaterMark       int64
	memHighWaterMark      int64
}

func NewMemoryAdaptiveTrafficShapingCalculator(owner *TrafficShapingController, r *Rule) *MemoryAdaptiveTrafficShapingCalculator {
	return &MemoryAdaptiveTrafficShapingCalculator{
		owner:                 owner,
		lowMemUsageThreshold:  r.LowMemUsageThreshold,
		highMemUsageThreshold: r.HighMemUsageThreshold,
		memLowWaterMark:       r.MemLowWaterMarkBytes,
		memHighWaterMark:      r.MemHighWaterMarkBytes,
	}
}

func (m *MemoryAdaptiveTrafficShapingCalculator) BoundOwner() *TrafficShapingController {
	return m.owner
}

func (m *MemoryAdaptiveTrafficShapingCalculator) CalculateAllowedTokens(_ uint32, _ int32) float64 {
	var threshold float64
	mem := system_metric.CurrentMemoryUsage()
	logging.Debug("[MemoryAdaptiveTrafficShapingCalculator CalculateAllowedTokens] Load memory usage", "mem", mem)
	if mem == system_metric.NotRetrievedMemoryValue {
		logging.Warn("[MemoryAdaptiveTrafficShapingCalculator CalculateAllowedTokens]Fail to load memory usage")
		return m.lowMemUsageThreshold
	}
	if mem <= m.memLowWaterMark {
		threshold = m.lowMemUsageThreshold
	} else if mem >= m.memHighWaterMark {
		threshold = m.highMemUsageThreshold
	} else {
		threshold = ((m.highMemUsageThreshold-m.lowMemUsageThreshold)/float64(m.memHighWaterMark-m.memLowWaterMark))*float64(mem-m.memLowWaterMark) + m.lowMemUsageThreshold
	}
	return threshold
}
