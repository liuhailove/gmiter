package flow

import (
	"github.com/liuhailove/gmiter/constants"
	"github.com/liuhailove/gmiter/core/config"
	"github.com/liuhailove/gmiter/spi"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"github.com/liuhailove/gmiter/core/base"

	"github.com/liuhailove/gmiter/core/stat"
	metric_exporter "github.com/liuhailove/gmiter/exporter/metric"
	"github.com/liuhailove/gmiter/logging"
	"github.com/liuhailove/gmiter/util"
)

const (
	RuleCheckSlotOrder = 2000
)

var (
	DefaultSlot   = &Slot{}
	flowWaitCount = metric_exporter.NewCounter(
		"flow_wait_total",
		"Flow wait count",
		[]string{"resource"})
	// DefaultDowngradeIntervalInNs 默认降级2s
	DefaultDowngradeIntervalInNs = int64(2 * time.Second)
)

func init() {
	metric_exporter.Register(flowWaitCount)
}

type Slot struct {
	isInitialized util.AtomicBool
}

func (s Slot) Order() uint32 {
	return RuleCheckSlotOrder
}

// Initial
//
// 初始化，如果有初始化工作放入其中
func (s Slot) Initial() {
}

func (s Slot) Check(ctx *base.EntryContext) *base.TokenResult {
	res := ctx.Resource.Name()
	logging.Debug("flow_slot", "res", res)
	tcs := getTrafficControllerListFor(res)
	result := ctx.RuleCheckResult

	// 按序检查规则
	for _, tc := range tcs {
		if tc == nil {
			logging.Warn("[FlowSlot Check]Nil traffic controller found", "resourceName", res)
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
		logging.Debug("flow_slot canPassCheck", "res", res)
		r := canPassCheck(tc, ctx.StatNode, ctx.Input.BatchCount)
		if r == nil {
			// nil means pass
			continue
		}
		if r.Status() == base.ResultStatusBlocked {
			logging.Info("flow_slot blocked", "res", res)
			return r
		}
		if r.Status() == base.ResultStatusShouldWait {
			logging.Info("flow_slot should wait", "res", res)
			if nanosToWait := r.NanosToWait(); nanosToWait > 0 {
				if flowWaitCount != nil {
					flowWaitCount.Add(float64(ctx.Input.BatchCount), ctx.Resource.Name())
				}
				// Handle waiting action.
				util.Sleep(nanosToWait)
			}
			continue
		}
	}
	return result
}

func canPassCheck(tc *TrafficShapingController, node base.StatNode, batchCount uint32) *base.TokenResult {
	return canPassCheckWithFlag(tc, node, batchCount, 0)
}

func canPassCheckWithFlag(tc *TrafficShapingController, node base.StatNode, batchCount uint32, flag int32) *base.TokenResult {
	if tc.rule.ClusterMode && (tc.rule.ClusterConfig.ClusterStrategy == int32(ThresholdGlobal) || tc.rule.ClusterConfig.ClusterStrategy == int32(ThresholdGlobalRedis)) {
		return checkInCluster(tc, node, batchCount, flag)
	}
	return checkInLocal(tc, node, batchCount, flag)
}

func selectNodeByRelStrategy(rule *Rule, node base.StatNode) base.StatNode {
	if rule.RelationStrategy == AssociatedResource {
		return stat.GetResourceNode(rule.RefResource)
	}
	return node
}

func checkInLocal(tc *TrafficShapingController, resStat base.StatNode, batchCount uint32, flag int32) *base.TokenResult {
	actual := selectNodeByRelStrategy(tc.rule, resStat)
	if actual == nil {
		logging.FrequentErrorOnce.Do(func() {
			logging.Error(errors.Errorf("nil resource node"), "No resource node for flow rule in FlowSlot.checkInLocal()", "rule", tc.rule)
		})
		return base.NewTokenResultPass()
	}
	return tc.PerformChecking(actual, batchCount, flag)
}

// checkInCluster 集群限流Check
func checkInCluster(tc *TrafficShapingController, resStat base.StatNode, batchCount uint32, flag int32) *base.TokenResult {
	// 线性限流时，使用提前占用方式
	if tc.rule.ControlBehavior == Throttling {
		flag = 1
	}
	var tokenResult *base.TokResult
	// 如果TokenServer由于某种原因产生降级，则必须降级时间到后才走TokenServer逻辑，否则走本地限流
	downgradeTimeInNsTo := atomic.LoadInt64(&tc.DowngradeTimeInNsTo)
	// 当前时间纳秒
	curNano := int64(util.CurrentTimeNano())
	if curNano < downgradeTimeInNsTo {
		logging.Warn("fallback to local", "curNano", curNano, "downgradeTimeInNsTo", downgradeTimeInNsTo)
		return fallbackToLocalOrPass(tc, resStat, batchCount, flag)
	}
	// 控制请求TokenServer的频率，可以理解为当单机频率超过阈值时，则不在请求TokenServer，避免TokenServer压力过大
	if !allowRemoteProceed() {
		logging.Warn("too many requests, fallback to local")
		return fallbackToLocalOrPass(tc, resStat, batchCount, flag)
	}
	switch tc.rule.ClusterConfig.ClusterStrategy {
	case int32(ThresholdGlobalRedis):
		var data, _ = jsonTraffic.Marshal(tc.rule)
		var inst = spi.GetRegisterTokenServiceInst(constants.RedisTokenServiceType)
		tokenResult = inst.GetTokenService().RequestToken(string(data), batchCount, flag)
	case int32(ThresholdGlobal):
		// 降级处理
		if tc.rule.ClusterConfig.TokenServerMasterHost == "" || tc.rule.ClusterConfig.TokenServerMasterPort <= 0 {
			return fallbackToLocalOrPass(tc, resStat, batchCount, flag)
		}
		tokenResult = TokenClient.RequestToken(tc.rule.ClusterConfig.TokenServerMasterHost, tc.rule.ClusterConfig.TokenServerMasterPort, tc.rule.ID, batchCount, flag)
	}

	switch base.TokResultStatus(tokenResult.Status) {
	case base.TokResultStatusOk:
		return nil
	case base.TokResultStatusShouldWait:
		// 等待下一个Tick
		if nanosToWait := tokenResult.WaitInMs; nanosToWait > 0 {
			// Handle waiting action.
			util.Sleep(time.Duration(nanosToWait) * time.Millisecond)
		}
		return nil
	case base.TokResultStatusFail:
		if swapped := atomic.CompareAndSwapInt64(&tc.DowngradeTimeInNsTo, downgradeTimeInNsTo, curNano+DefaultDowngradeIntervalInNs); swapped {
			// 当出现失败时此请求回退到本地限流
			logging.Warn("fallback to local because remote error", "curNano", curNano)
			return fallbackToLocalOrPass(tc, resStat, batchCount, flag)
		}
		fallthrough
	case base.TokResultStatusNoRuleExists:
		fallthrough
	case base.TokResultStatusBadRequest:
		fallthrough
	case base.TokResultStatusTooManyRequest:
		return fallbackToLocalOrPass(tc, resStat, batchCount, flag)
	case base.TokResultStatusBlocked:
		return base.NewTokenResultBlockedWithCause(base.BlockTypeFlow, "flow reject check blocked", tc.rule, nil)
	default:
	}
	return nil
}

func fallbackToLocalOrPass(tc *TrafficShapingController, resStat base.StatNode, batchCount uint32, flag int32) *base.TokenResult {
	if tc.rule.ClusterConfig.FallbackToLocalWhenFail {
		return checkInLocal(tc, resStat, batchCount, flag)
	}
	return nil
}

// allowRemoteProceed 是否允许继续请求远程Token
func allowRemoteProceed() bool {
	return ClientTryPass(config.ClientNamespace())
}
