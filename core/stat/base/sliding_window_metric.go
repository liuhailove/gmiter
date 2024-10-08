package base

import (
	"github.com/liuhailove/gmiter/core/base"
	"github.com/liuhailove/gmiter/logging"
	"github.com/liuhailove/gmiter/util"
	"github.com/pkg/errors"
	"reflect"
	"sync/atomic"
)

// SlidingWindowMetric represents the sliding window metric wrapper.
// It does not store any data and is the wrapper of BucketLeapArray to adapt to different internal bucket.
//
// SlidingWindowMetric is designed as a high-level, read-only statistic structure for functionalities of sea
type SlidingWindowMetric struct {
	bucketLengthInMs uint32
	sampleCount      uint32
	intervalInMs     uint32
	real             *BucketLeapArray
}

// NewSlidingWindowMetric creates a SlidingWindowMetric with given attributes.
// The pointer to the internal statistic BucketLeapArray should be valid.
func NewSlidingWindowMetric(sampleCount, intervalInMs uint32, real *BucketLeapArray) (*SlidingWindowMetric, error) {
	if real == nil {
		return nil, errors.New("nil BucketLeapArray")
	}
	if err := base.CheckValidityForReuseStatistic(sampleCount, intervalInMs, real.SampleCount(), real.IntervalInMs()); err != nil {
		return nil, err
	}
	bucketLengthInMs := intervalInMs / sampleCount

	return &SlidingWindowMetric{
		bucketLengthInMs: bucketLengthInMs,
		sampleCount:      sampleCount,
		intervalInMs:     intervalInMs,
		real:             real,
	}, nil
}

// getBucketStartRange 返回给定时间的存储桶的开始时间范围。
// 实际时间跨度为：[start, end + in.bucketTimeLength)
func (m *SlidingWindowMetric) getBucketStartRange(timeMs uint64) (start, end uint64) {
	curBucketStartTime := calculateStartTime(timeMs, m.real.BucketLengthInMs())
	end = curBucketStartTime
	start = end - uint64(m.intervalInMs) + uint64(m.real.BucketLengthInMs())
	return
}

func (m *SlidingWindowMetric) getIntervalInSecond() float64 {
	return float64(m.intervalInMs) / 1000.0
}

func (m *SlidingWindowMetric) count(event base.MetricEvent, values []*BucketWrap) int64 {
	ret := int64(0)
	for _, ww := range values {
		mb := ww.Value.Load()
		if mb == nil {
			logging.Error(errors.New("nil BucketWrap"), "Current bucket value is nil in SlidingWindowMetric.count()")
			continue
		}
		counter, ok := mb.(*MetricBucket)
		if !ok {
			logging.Error(errors.New("type assert failed"), "Fail to do type assert in SlidingWindowMetric.count()", "expectType", "*MetricBucket", "actualType", reflect.TypeOf(mb).Name())
			continue
		}
		ret += counter.Get(event)
	}
	return ret
}

func (m *SlidingWindowMetric) GetSum(event base.MetricEvent) int64 {
	return m.getSumWithTime(util.CurrentTimeMillis(), event)
}

func (m *SlidingWindowMetric) GetIntervalInSecond() float64 {
	return m.getIntervalInSecond()
}

func (m *SlidingWindowMetric) GetSumWithTime(now uint64, event base.MetricEvent) int64 {
	return m.getSumWithTime(now, event)
}

func (m *SlidingWindowMetric) getSumWithTime(now uint64, event base.MetricEvent) int64 {
	satisfiedBuckets := m.getSatisfiedBuckets(now)
	return m.count(event, satisfiedBuckets)
}

func (m *SlidingWindowMetric) GetQPS(event base.MetricEvent) float64 {
	return m.getQPSWithTime(util.CurrentTimeMillis(), event)
}

func (m *SlidingWindowMetric) GetPreviousQPS(event base.MetricEvent) float64 {
	return m.getQPSWithTime(util.CurrentTimeMillis()-uint64(m.bucketLengthInMs), event)
}

func (m *SlidingWindowMetric) getQPSWithTime(now uint64, event base.MetricEvent) float64 {
	return float64(m.getSumWithTime(now, event)) / m.getIntervalInSecond()
}

func (m *SlidingWindowMetric) getSatisfiedBuckets(now uint64) []*BucketWrap {
	start, end := m.getBucketStartRange(now)
	// 提取startTime在[start, end]之间的桶
	// 这意味着bucket的时间视图是[firstStart, endStart+bucketLength)
	satisfiedBuckets := m.real.ValuesConditional(now, func(ws uint64) bool {
		return ws >= start && ws <= end
	})
	return satisfiedBuckets
}

func (m *SlidingWindowMetric) GetMaxOfSingleBucket(event base.MetricEvent) int64 {
	now := util.CurrentTimeMillis()
	satisfiedBuckets := m.getSatisfiedBuckets(now)
	var curMax int64 = 0
	for _, w := range satisfiedBuckets {
		mb := w.Value.Load()
		if mb == nil {
			logging.Error(errors.New("nil BucketWrap"), "Current bucket value is nil in SlidingWindowMetric.GetMaxOfSingleBucket()")
			continue
		}
		counter, ok := mb.(*MetricBucket)
		if !ok {
			logging.Error(errors.New("type assert failed"), "Fail to do type assert in SlidingWindowMetric.GetMaxOfSingleBucket()", "expectType", "*MetricBucket", "actualType", reflect.TypeOf(mb).Name())
			continue
		}
		v := counter.Get(event)
		if v > curMax {
			curMax = v
		}
	}
	return curMax
}

func (m *SlidingWindowMetric) MinRT() float64 {
	now := util.CurrentTimeMillis()
	satisfiedBuckets := m.getSatisfiedBuckets(now)
	minRt := base.DefaultStatisticMaxRt
	for _, w := range satisfiedBuckets {
		mb := w.Value.Load()
		if mb == nil {
			logging.Error(errors.New("nil BucketWrap"), "Current bucket value is nil in SlidingWindowMetric.MinRT()")
			continue
		}
		counter, ok := mb.(*MetricBucket)
		if !ok {
			logging.Error(errors.New("type assert failed"), "Fail to do type assert in SlidingWindowMetric.MinRT()", "expectType", "*MetricBucket", "actualType", reflect.TypeOf(mb).Name())
			continue
		}
		v := counter.MinRt()
		if v < minRt {
			minRt = v
		}
	}
	if minRt < 1 {
		minRt = 1
	}
	return float64(minRt)
}

func (m *SlidingWindowMetric) MaxConcurrency() int32 {
	now := util.CurrentTimeMillis()
	satisfiedBuckets := m.getSatisfiedBuckets(now)
	maxConcurrency := int32(0)
	for _, w := range satisfiedBuckets {
		mb := w.Value.Load()
		if mb == nil {
			logging.Error(errors.New("nil BucketWrap"), "Current bucket value is nil in SlidingWindowMetric.MaxConcurrency()")
			continue
		}
		counter, ok := mb.(*MetricBucket)
		if !ok {
			logging.Error(errors.New("type assert failed"), "Fail to do type assert in SlidingWindowMetric.MaxConcurrency()", "expectType", "*MetricBucket", "actualType", reflect.TypeOf(mb).Name())
			continue
		}
		v := counter.MaxConcurrency()
		if v > maxConcurrency {
			maxConcurrency = v
		}
	}
	return maxConcurrency
}

func (m *SlidingWindowMetric) AvgRT() float64 {
	return float64(m.GetSum(base.MetricEventRt)) / float64(m.GetSum(base.MetricEventComplete))
}

// SecondMetricsOnCondition aggregates metric items by second on condition that
// the startTime of the statistic buckets satisfies the time predicate.
func (m *SlidingWindowMetric) SecondMetricsOnCondition(predicate base.TimePredicate) []*base.MetricItem {
	ws := m.real.ValuesConditional(util.CurrentTimeMillis(), predicate)

	// Aggregate second-level MetricItem (only for stable metrics)
	wm := make(map[uint64][]*BucketWrap, 8)
	for _, w := range ws {
		bucketStart := atomic.LoadUint64(&w.BucketStart)
		secStart := bucketStart - bucketStart%1000
		if arr, hasData := wm[secStart]; hasData {
			wm[secStart] = append(arr, w)
		} else {
			wm[secStart] = []*BucketWrap{w}
		}
	}
	items := make([]*base.MetricItem, 0, 8)
	for ts, values := range wm {
		if len(values) == 0 {
			continue
		}
		if item := m.metricItemFromBuckets(ts, values); item != nil {
			items = append(items, item)
		}
	}
	return items
}

// metricItemFromBuckets aggregates multiple bucket wrappers (based on the same startTime in second)
// to the single MetricItem.
func (m *SlidingWindowMetric) metricItemFromBuckets(ts uint64, ws []*BucketWrap) *base.MetricItem {
	item := &base.MetricItem{Timestamp: ts}
	var allRt int64 = 0
	for _, w := range ws {
		mi := w.Value.Load()
		if mi == nil {
			logging.Error(errors.New("nil BucketWrap"), "Current bucket value is nil in SlidingWindowMetric.metricItemFromBuckets()")
			return nil
		}
		mb, ok := mi.(*MetricBucket)
		if !ok {
			logging.Error(errors.New("type assert failed"), "Fail to do type assert in SlidingWindowMetric.metricItemFromBuckets()", "bucketStartTime", w.BucketStart, "expectType", "*MetricBucket", "actualType", reflect.TypeOf(mb).Name())
			return nil
		}
		item.PassQps += uint64(mb.Get(base.MetricEventPass))
		item.BlockQps += uint64(mb.Get(base.MetricEventBlock))
		item.ErrorQps += uint64(mb.Get(base.MetricEventError))
		item.CompleteQps += uint64(mb.Get(base.MetricEventComplete))
		mc := uint32(mb.MaxConcurrency())
		if mc > item.Concurrency {
			item.Concurrency = mc
		}
		allRt += mb.Get(base.MetricEventRt)
		item.BlockedNumByFlow += uint64(mb.Get(base.MetricEventBlockFlow))
		item.BlockedNumByIsolation += uint64(mb.Get(base.MetricEventBlockIsolation))
		item.BlockedNumByCircuitBreaking += uint64(mb.Get(base.MetricEventBlockCircuitBreaking))
		item.BlockedNumBySystem += uint64(mb.Get(base.MetricEventBlockSystem))
		item.BlockedNumByHotspotParam += uint64(mb.Get(base.MetricEventBlockHotSpotParamFlow))
		item.BlockedNumByMock += uint64(mb.Get(base.MetricEventBlockMock))
	}
	if item.CompleteQps > 0 {
		item.AvgRt = uint64(allRt) / item.CompleteQps
	} else {
		item.AvgRt = uint64(allRt)
	}
	return item
}

func (m *SlidingWindowMetric) metricItemFromBucket(w *BucketWrap) *base.MetricItem {
	mi := w.Value.Load()
	if mi == nil {
		logging.Error(errors.New("nil BucketWrap"), "Current bucket value is nil in SlidingWindowMetric.metricItemFromBucket()")
		return nil
	}
	mb, ok := mi.(*MetricBucket)
	if !ok {
		logging.Error(errors.New("type assert failed"), "Fail to do type assert in SlidingWindowMetric.metricItemFromBucket()", "expectType", "*MetricBucket", "actualType", reflect.TypeOf(mb).Name())
		return nil
	}
	completeQps := mb.Get(base.MetricEventComplete)
	item := &base.MetricItem{
		PassQps:     uint64(mb.Get(base.MetricEventPass)),
		BlockQps:    uint64(mb.Get(base.MetricEventBlock)),
		ErrorQps:    uint64(mb.Get(base.MetricEventError)),
		CompleteQps: uint64(completeQps),
		Timestamp:   w.BucketStart,
	}
	if completeQps > 0 {
		item.AvgRt = uint64(mb.Get(base.MetricEventRt) / completeQps)
	} else {
		item.AvgRt = uint64(mb.Get(base.MetricEventRt))
	}
	return item
}
