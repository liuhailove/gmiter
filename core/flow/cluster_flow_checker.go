package flow

import (
	base2 "github.com/liuhailove/gmiter/core/base"
	"math"
	"sync/atomic"
	"time"

	"github.com/liuhailove/gmiter/core/config"
	"github.com/liuhailove/gmiter/logging"
	"github.com/liuhailove/gmiter/util"
)

func calcGlobalThreshold(rule *Rule) float64 {
	switch rule.ClusterConfig.ClusterStrategy {
	case int32(ThresholdGlobal):
		return rule.ClusterConfig.GlobalThreshold
	default:
		return rule.Threshold
	}
}

// allowProceed 是否允许继续处理
func allowProceed() bool {
	return GlobalTryPass(config.Namespace())
}

func acquireClusterToken(rule *Rule, acquireCount uint32, prioritized int32) *base2.TokResult {
	if !allowProceed() {
		return base2.TooManyRequestResult
	}
	var clusterMetric *ClusterMetric
	if rule.RelationStrategy == AssociatedResource {
		clusterMetric = getClusterMetric(rule.RefResource)
	} else {
		clusterMetric = getClusterMetric(rule.Resource)
	}
	if clusterMetric == nil {
		return base2.TokenResultStatusFailResult
	}
	var latestQps = clusterMetric.GetAvg(base2.MetricEventPass)
	var globalThreshold = calcGlobalThreshold(rule)
	var nextRemaining = globalThreshold - latestQps - float64(acquireCount)
	logging.Debug("in acquireClusterToken", "latestQps", latestQps, "nextRemaining", nextRemaining, "globalThreshold", globalThreshold)
	if prioritized > 0 {
		return priorityHandle(acquireCount, globalThreshold, clusterMetric, nextRemaining)
	}
	if nextRemaining >= 0 {
		clusterMetric.AddCount(base2.MetricEventPass, int64(acquireCount))
		clusterMetric.AddCount(base2.MetricEventPassRequest, 1)
		return base2.StatusOkResult
	} else {
		// 被阻塞
		clusterMetric.AddCount(base2.MetricEventBlock, int64(acquireCount))
		clusterMetric.AddCount(base2.MetricEventBlockRequest, 1)
		logging.Debug("flow block", "ruleId", rule.ID, "acquireCount", acquireCount)
		logging.Debug("flow block_request", "ruleId", rule.ID, "count", 1)
		return base2.BlockedResult
	}
}

// priorityHandle 线性限流处理
func priorityHandle(acquireCount uint32, globalThreshold float64, clusterMetric *ClusterMetric, nextRemaining float64) *base2.TokResult {
	// 这里我们使用【纳秒】以便可以更加精确的控制队列时间
	curNano := int64(util.CurrentTimeNano())

	// 计算两个请求间应该的时间差【纳秒】
	intervalNs := int64(math.Ceil(float64(acquireCount) / globalThreshold * float64(clusterMetric.statIntervalNs)))

	// 找到上次请求的通过时间
	loadedLastPassedTime := atomic.LoadInt64(&clusterMetric.lastPassedTime)
	// 找到上次的系统时间
	loadedLastSystemTime := atomic.LoadInt64(&clusterMetric.lastSystemTime)
	if loadedLastSystemTime > curNano+int64(time.Second) {
		// 放行，系统时钟向前发生了偏移，此处容忍1S误差
		if swapped := atomic.CompareAndSwapInt64(&clusterMetric.lastPassedTime, loadedLastPassedTime, curNano); swapped {
			// nil means pass
			return nil
		}
	}
	// 计算请求的预期到达时间
	expectedTime := loadedLastPassedTime + intervalNs
	// 如果预期到达时间小于当前时间，这说明已经可以放行了，
	// 如果大于当前时间，说明放行时间还未到，需要等待
	if expectedTime <= curNano {
		if swapped := atomic.CompareAndSwapInt64(&clusterMetric.lastPassedTime, loadedLastPassedTime, curNano); swapped {
			clusterMetric.AddCount(base2.MetricEventPass, int64(acquireCount))
			clusterMetric.AddCount(base2.MetricEventPassRequest, 1)
			return base2.StatusOkResult
		}
	}
	// 预期在队列的等待时间
	estimatedQueueingDuration := atomic.LoadInt64(&clusterMetric.lastPassedTime) + intervalNs - curNano
	// 如果等待时间超过队列的最大等待时间，则直接拒绝
	if estimatedQueueingDuration > clusterMetric.maxQueueingTimeNs {
		// 被阻塞
		clusterMetric.AddCount(base2.MetricEventBlock, int64(acquireCount))
		clusterMetric.AddCount(base2.MetricEventBlockRequest, 1)
		return base2.BlockedResult
	}
	// 上次通过时间+间隔时间为下次要放行时间
	oldTime := atomic.AddInt64(&clusterMetric.lastPassedTime, intervalNs)
	// 重新计算预期等待时间
	estimatedQueueingDuration = oldTime - curNano
	// 重新计算和队列排队时间的关系，这主要是为了避免在上面计算后多个请求同时进入的场景
	if estimatedQueueingDuration > clusterMetric.maxQueueingTimeNs {
		// 如果计算后发现排队过旧，把加到last通过的时间减回来
		atomic.AddInt64(&clusterMetric.lastPassedTime, -intervalNs)
		// 被阻塞
		clusterMetric.AddCount(base2.MetricEventBlock, int64(acquireCount))
		clusterMetric.AddCount(base2.MetricEventBlockRequest, 1)
		return base2.BlockedResult
	}
	clusterMetric.AddCount(base2.MetricEventPass, int64(acquireCount))
	clusterMetric.AddCount(base2.MetricEventPassRequest, 1)
	if estimatedQueueingDuration > 0 {
		return &base2.TokResult{Status: int(base2.TokResultStatusShouldWait), WaitInMs: int(estimatedQueueingDuration / MillisToNanosOffset)}
	}
	return &base2.TokResult{Status: int(base2.TokResultStatusOk), WaitInMs: 0}
}
