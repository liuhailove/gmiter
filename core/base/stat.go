package base

import "errors"

type TimePredicate func(uint64) bool

type MetricEvent int8

// There are five events to record
// pass + block == Total
const (
	// MetricEventPass sea rules check pass
	MetricEventPass MetricEvent = iota

	// MetricEventBlock sea rules check block
	MetricEventBlock

	MetricEventComplete

	// MetricEventError Biz error, used for circuit breaker
	MetricEventError
	// MetricEventRt request execute rt, unit is millisecond
	MetricEventRt

	// MetricEventWaiting 由于流控或者为请求下一个bucket tick而等待
	MetricEventWaiting

	// MetricEventPassRequest token请求通过(从client)
	MetricEventPassRequest
	// MetricEventBlockRequest token请求被阻塞（从client）
	MetricEventBlockRequest

	// MetricEventBlockFlow 请求被流控拒绝
	MetricEventBlockFlow

	// MetricEventBlockIsolation 由于资源隔离被阻塞实现
	MetricEventBlockIsolation

	// MetricEventBlockCircuitBreaking 由于熔断被阻塞事件
	MetricEventBlockCircuitBreaking

	// MetricEventBlockSystem 由于系统限流被阻塞事件
	MetricEventBlockSystem

	// MetricEventBlockHotSpotParamFlow 由于热点限流被阻塞事件
	MetricEventBlockHotSpotParamFlow

	// MetricEventBlockMock 由于Mock被阻塞事件
	MetricEventBlockMock

	// MetricEventRejectPass 拒绝策略下的通过
	MetricEventRejectPass

	// MetricEventTotal hack for the number of event
	MetricEventTotal
)

var (
	globalNopReadStat  = &nopReadStat{}
	globalNopWriteStat = &nopWriteStat{}
)

type ReadStat interface {
	GetQPS(event MetricEvent) float64
	GetPreviousQPS(event MetricEvent) float64
	GetSum(event MetricEvent) int64
	GetSumWithTime(now uint64, event MetricEvent) int64

	MinRT() float64
	AvgRT() float64
}

func NopReadStat() *nopReadStat {
	return globalNopReadStat
}

type nopReadStat struct {
}

func (n nopReadStat) GetQPS(_ MetricEvent) float64 {
	return 0.0
}

func (n nopReadStat) GetPreviousQPS(_ MetricEvent) float64 {
	return 0.0
}

func (n nopReadStat) GetSum(_ MetricEvent) int64 {
	return 0
}

func (n nopReadStat) MinRT() float64 {
	return 0.0
}

func (n nopReadStat) AvgRT() float64 {
	return 0.0
}

func (n nopReadStat) GetSumWithTime(_ uint64, _ MetricEvent) int64 {
	return 0
}

type WriteStat interface {
	// AddCount adds given count to the metric of provided MetricEvent.
	AddCount(event MetricEvent, count int64)

	AddCountWithTime(now uint64, event MetricEvent, count int64)

	//// AddAndGetSum 增加count，并GetSum汇总
	//AddAndGetSum(event MetricEvent, count int64) int64
}

func NopWriteStat() *nopWriteStat {
	return globalNopWriteStat
}

type nopWriteStat struct {
}

func (ws *nopWriteStat) AddCount(_ MetricEvent, _ int64) {}

func (ws *nopWriteStat) AddCountWithTime(_ uint64, _ MetricEvent, _ int64) {}

// ConcurrencyStat provides read/update operation for concurrency statistics.
type ConcurrencyStat interface {
	CurrentConcurrency() int32
	IncreaseConcurrency()
	DecreaseConcurrency()
}

// StatNode holds real-time statistics for resources.
type StatNode interface {
	MetricItemRetriever

	ReadStat
	WriteStat
	ConcurrencyStat
	// GenerateReadStat generates the readonly metric statistic based on resource level global statistic
	// If parameters, sampleCount and intervalInMs, are not suitable for resource level global statistic, return (nil, error)
	GenerateReadStat(sampleCount uint32, intervalInMs uint32) (ReadStat, error)
}

var (
	IllegalGlobalStatisticParamsError = errors.New("invalid parameters, sampleCount or interval, for resource's global statistic")
	IllegalStatisticParamsError       = errors.New("invalid parameters, sampleCount or interval, for metric statistic")
	GlobalStatisticNonReusableError   = errors.New("the parameters, sampleCount and interval, mismatch for reusing between resource's global statistic and readonly metric statistic")
)

func CheckValidityForStatistic(sampleCount, intervalInMs uint32) error {
	if intervalInMs == 0 || sampleCount == 0 || intervalInMs%sampleCount != 0 {
		return IllegalStatisticParamsError
	}
	return nil
}

// CheckValidityForReuseStatistic checks whether the read-only stat-metric with given attributes
// (i.e. sampleCount and intervalInMs) can be built based on underlying global statistics data-structure
// with given attributes (parentSampleCount and parentIntervalInMs). Returns nil if the attributes
// satisfy the validation, or return specific error if not.
//
// The parameters, sampleCount and intervalInMs, are the attributes of the stat-metric view you want to build.
// The parameters, parentSampleCount and parentIntervalInMs, are the attributes of the underlying statistics data-structure.
func CheckValidityForReuseStatistic(sampleCount, intervalInMs uint32, parentSampleCount, parentIntervalInMs uint32) error {
	if intervalInMs == 0 || sampleCount == 0 || intervalInMs%sampleCount != 0 {
		return IllegalStatisticParamsError
	}
	bucketLengthInMs := intervalInMs / sampleCount
	if parentIntervalInMs == 0 || parentSampleCount == 0 || parentIntervalInMs%parentSampleCount != 0 {
		return IllegalGlobalStatisticParamsError
	}

	parentBucketLengthInMs := parentIntervalInMs / parentSampleCount
	// intervalInMs of the SlidingWindowMetric is not divisible by BucketLeapArray's intervalInMs
	if parentIntervalInMs%intervalInMs != 0 {
		return GlobalStatisticNonReusableError
	}
	// BucketLeapArray's BucketLengthInMs is not divisible by BucketLengthInMs of SlidingWindowMetric
	if bucketLengthInMs%parentBucketLengthInMs != 0 {
		return GlobalStatisticNonReusableError
	}
	return nil
}
