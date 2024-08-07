package hotspot

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"strings"
)

const (
	RuleCheckSlotOrder = 4000
)

var (
	DefaultSlot = &Slot{}
)

type Slot struct {
}

func (s *Slot) Order() uint32 {
	return RuleCheckSlotOrder
}

// Initial
//
// 初始化，如果有初始化工作放入其中
func (s *Slot) Initial() {}

func (s *Slot) Check(ctx *base.EntryContext) *base.TokenResult {
	res := ctx.Resource.Name()
	batch := int64(ctx.Input.BatchCount)

	result := ctx.RuleCheckResult
	tcs := getTrafficControllersFor(res)
	for _, tc := range tcs {
		arg := tc.ExtractArgs(ctx)
		if arg == nil {
			continue
		}
		// 来源检查
		var needContinueCheck = true
		if tc.BoundRule().LimitApp != "" && !strings.EqualFold(tc.BoundRule().LimitApp, "default") {
			if !util.Contains(ctx.FromService, strings.Split(tc.BoundRule().LimitApp, ",")) {
				needContinueCheck = false
			}
		}
		if !needContinueCheck {
			continue
		}
		r := canPassCheck(tc, arg, batch)
		if r == nil {
			continue
		}
		if r.Status() == base.ResultStatusBlocked {
			return r
		}
		if r.Status() == base.ResultStatusShouldWait {
			if nanosToWait := r.NanosToWait(); nanosToWait > 0 {
				// Handle waiting action.
				util.Sleep(nanosToWait)
			}
			continue
		}
	}
	return result
}

func canPassCheck(tc TrafficShapingController, arg interface{}, batch int64) *base.TokenResult {
	var rule = tc.BoundRule()
	if rule.MetricType == QPS &&
		rule.ClusterMode &&
		rule.ClusterConfig != nil &&
		rule.ClusterConfig.ClusterStrategy == int32(ThresholdGlobal) {
		// 集群策略
		return canPassClusterCheck(tc, arg, batch)
	}
	return canPassLocalCheck(tc, arg, batch)
}

func canPassLocalCheck(tc TrafficShapingController, arg interface{}, batch int64) *base.TokenResult {
	return tc.PerformChecking(arg, batch)
}

func canPassClusterCheck(tc TrafficShapingController, arg interface{}, batch int64) *base.TokenResult {
	// 降级处理
	var rule = tc.BoundRule()
	// 适配节点均摊模式
	if rule.ClusterConfig.TokenServerMasterHost == "" || rule.ClusterConfig.TokenServerMasterPort <= 0 {
		return fallbackToLocalOrPass(tc, arg, batch)
	}
	// TODO 对于选主模式，目前暂不支持，之后通过私有协议优化选主的通信
	return nil
}

func fallbackToLocalOrPass(tc TrafficShapingController, arg interface{}, batch int64) *base.TokenResult {
	if tc.BoundRule().ClusterConfig.FallbackToLocalWhenFail {
		return canPassLocalCheck(tc, arg, batch)
	}
	return nil
}
