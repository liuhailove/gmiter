package flow

import (
	"math"
	"sync/atomic"
	"time"

	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/util"
)

const (
	BlockMsgQueueing = "flow throttling check blocked, estimated queueing time exceeds max queueing time"

	MillisToNanosOffset = int64(time.Millisecond / time.Nanosecond)
)

// ThrottlingChecker limits the time interval between two requests.
type ThrottlingChecker struct {
	owner             *TrafficShapingController
	maxQueueingTimeNs int64
	statIntervalNs    int64
	lastPassedTime    int64
	lastSystemTime    int64
}

func NewThrottlingChecker(owner *TrafficShapingController, timeoutMs uint32, statIntervalMs uint32) *ThrottlingChecker {
	var statIntervalNs int64
	if statIntervalMs == 0 {
		statIntervalNs = 1000 * MillisToNanosOffset
	} else {
		statIntervalNs = int64(statIntervalMs) * MillisToNanosOffset
	}
	return &ThrottlingChecker{
		owner:             owner,
		maxQueueingTimeNs: int64(timeoutMs) * MillisToNanosOffset,
		statIntervalNs:    statIntervalNs,
		lastPassedTime:    0,
		lastSystemTime:    0,
	}
}

func (c *ThrottlingChecker) BoundOwner() *TrafficShapingController {
	return c.owner
}

func (c *ThrottlingChecker) DoCheck(_ base.StatNode, batchCount uint32, threshold float64) *base.TokenResult {
	defer func() {
		// 每次操作都要更新上次的系统时间
		if swapped2 := atomic.CompareAndSwapInt64(&c.lastSystemTime, atomic.LoadInt64(&c.lastSystemTime), int64(util.CurrentTimeNano())); swapped2 {
			// do nothing
		}
	}()
	// Pass when batch count is less or equal than 0.
	if batchCount <= 0 {
		return nil
	}

	var rule *Rule
	if c.BoundOwner() != nil {
		rule = c.BoundOwner().BoundRule()
	}

	if threshold <= 0.0 {
		msg := "flow throttling check blocked, threshold is <= 0.0"
		return base.NewTokenResultBlockedWithCause(base.BlockTypeFlow, msg, rule, nil)
	}

	// 这里我们使用【纳秒】以便可以更加精确的控制队列时间
	curNano := int64(util.CurrentTimeNano())

	// 计算两个请求间应该的时间差【纳秒】
	intervalNs := int64(math.Ceil(float64(batchCount) / threshold * float64(c.statIntervalNs)))

	// 找到上次请求的通过时间
	loadedLastPassedTime := atomic.LoadInt64(&c.lastPassedTime)
	// 找到上次的系统时间
	loadedLastSystemTime := atomic.LoadInt64(&c.lastSystemTime)
	if loadedLastSystemTime > curNano+int64(time.Second) {
		// 放行，系统时钟向前发生了偏移，此处容忍1S误差
		if swapped := atomic.CompareAndSwapInt64(&c.lastPassedTime, loadedLastPassedTime, curNano); swapped {
			// nil means pass
			return nil
		}
	}
	// 计算请求的预期到达时间
	expectedTime := loadedLastPassedTime + intervalNs
	// 如果预期到达时间小于当前时间，这说明已经可以放行了，
	// 如果大于当前时间，说明放行时间还未到，需要等待
	if expectedTime <= curNano {
		if swapped := atomic.CompareAndSwapInt64(&c.lastPassedTime, loadedLastPassedTime, curNano); swapped {
			// nil means pass
			return nil
		}
	}
	// 预期在队列的等待时间
	estimatedQueueingDuration := atomic.LoadInt64(&c.lastPassedTime) + intervalNs - curNano
	// 如果等待时间超过队列的最大等待时间，则直接拒绝
	if estimatedQueueingDuration > c.maxQueueingTimeNs {
		return base.NewTokenResultBlockedWithCause(base.BlockTypeFlow, BlockMsgQueueing, rule, nil)
	}
	// 上次通过时间+间隔时间为下次要放行时间
	oldTime := atomic.AddInt64(&c.lastPassedTime, intervalNs)
	// 重新计算预期等待时间
	estimatedQueueingDuration = oldTime - curNano
	// 重新计算和队列排队时间的关系，这主要是为了避免在上面计算后多个请求同时进入的场景
	if estimatedQueueingDuration > c.maxQueueingTimeNs {
		// 如果计算后发现排队过旧，把加到last通过的时间减回来
		atomic.AddInt64(&c.lastPassedTime, -intervalNs)
		return base.NewTokenResultBlockedWithCause(base.BlockTypeFlow, BlockMsgQueueing, rule, nil)
	}
	if estimatedQueueingDuration > 0 {
		return base.NewTokenResultShouldWait(time.Duration(estimatedQueueingDuration))
	}
	return base.NewTokenResultShouldWait(0)
}
