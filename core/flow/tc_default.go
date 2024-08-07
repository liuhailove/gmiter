package flow

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/util"
)

type DirectTrafficShapingCalculator struct {
	owner     *TrafficShapingController
	threshold float64
}

func NewDirectTrafficShapingCalculator(owner *TrafficShapingController, threshold float64) *DirectTrafficShapingCalculator {
	return &DirectTrafficShapingCalculator{
		owner:     owner,
		threshold: threshold,
	}
}

func (d *DirectTrafficShapingCalculator) CalculateAllowedTokens(uint32, int32) float64 {
	return d.threshold
}

func (d *DirectTrafficShapingCalculator) BoundOwner() *TrafficShapingController {
	return d.owner
}

type RejectTrafficShapingChecker struct {
	owner *TrafficShapingController
	rule  *Rule
}

func NewRejectTrafficShapingChecker(owner *TrafficShapingController, rule *Rule) *RejectTrafficShapingChecker {
	return &RejectTrafficShapingChecker{
		owner: owner,
		rule:  rule,
	}
}

func (d *RejectTrafficShapingChecker) BoundOwner() *TrafficShapingController {
	return d.owner
}

func (d *RejectTrafficShapingChecker) DoCheck(resStat base.StatNode, batchCount uint32, threshold float64) *base.TokenResult {
	metricReadonlyStat := d.BoundOwner().boundStat.readOnlyMetric
	if metricReadonlyStat == nil {
		return nil
	}
	metricWriteOnlyStat := d.BoundOwner().boundStat.writeOnlyMetric
	if metricWriteOnlyStat == nil {
		return nil
	}
	var now = util.CurrentTimeMillis()
	// 求和
	curCount := float64(metricReadonlyStat.GetSumWithTime(now, base.MetricEventRejectPass))
	// 有可能存在并发问题，此处不加锁
	if curCount > threshold {
		// 超过阈值了，直接拒绝
		msg := "flow reject check blocked"
		return base.NewTokenResultBlockedWithCause(base.BlockTypeFlow, msg, d.rule, curCount)
	}
	// 预先加1
	metricWriteOnlyStat.AddCountWithTime(now, base.MetricEventRejectPass, int64(batchCount))
	// 由于并发问题，+1时有可能超过阈值，需要再次判断
	curCount = float64(metricReadonlyStat.GetSumWithTime(now, base.MetricEventRejectPass))
	// 再次判断，如果超过了阈值，需要减回去
	if curCount > threshold {
		// 超过阈值了，需要减回来
		metricWriteOnlyStat.AddCountWithTime(now, base.MetricEventRejectPass, int64(batchCount)*(-1))
		msg := "flow reject check blocked"
		return base.NewTokenResultBlockedWithCause(base.BlockTypeFlow, msg, d.rule, curCount)
	}
	return nil
}
