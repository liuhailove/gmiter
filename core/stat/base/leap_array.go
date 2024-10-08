package base

import (
	"fmt"
	"github.com/liuhailove/gmiter/core/base"
	"github.com/liuhailove/gmiter/logging"
	"github.com/liuhailove/gmiter/util"
	"github.com/pkg/errors"
	"runtime"
	"sync/atomic"
	"unsafe"
)

// BucketWrap represents a slot to record metrics.
//
// In order to reduce memory footprint, BucketWrap does not hold the length of the bucket.
// The length of BucketWrap could be seen in LeapArray.
// The scope of time is [startTime, startTime+bucketLength).
// The size of BucketWrap is 24(8+16) bytes.
type BucketWrap struct {
	// BucketStart represents start timestamp of this statistic bucket wrapper.
	BucketStart uint64
	// Value represents the actual data structure of the metrics (e.g. MetricBucket).
	Value atomic.Value
}

func (ww *BucketWrap) resetTo(startTime uint64) {
	ww.BucketStart = startTime
}

func (ww *BucketWrap) isTimeInBucket(now uint64, bucketLengthInMs uint32) bool {
	return ww.BucketStart <= now && now < ww.BucketStart+uint64(bucketLengthInMs)
}

func calculateStartTime(now uint64, bucketLengthInMs uint32) uint64 {
	return now - (now % uint64(bucketLengthInMs))
}

// AtomicBucketWrapArray represents a thread-safe circular array.
//
// The length of the array should be provided on-create and cannot be modified.
type AtomicBucketWrapArray struct {
	// The base address for real data array
	base unsafe.Pointer
	// The length of slice(array), it can not be modified.
	length int
	data   []*BucketWrap
}

func NewAtomicBucketWrapArrayWithTime(len int, bucketLengthInMs uint32, now uint64, generator BucketGenerator) *AtomicBucketWrapArray {
	ret := &AtomicBucketWrapArray{
		length: len,
		data:   make([]*BucketWrap, len),
	}

	idx := int((now / uint64(bucketLengthInMs)) % uint64(len))
	startTime := calculateStartTime(now, bucketLengthInMs)

	for i := idx; i <= len-1; i++ {
		ww := &BucketWrap{
			BucketStart: startTime,
			Value:       atomic.Value{},
		}
		ww.Value.Store(generator.NewEmptyBucket())
		ret.data[i] = ww
		startTime += uint64(bucketLengthInMs)
	}
	for i := 0; i < idx; i++ {
		ww := &BucketWrap{
			BucketStart: startTime,
			Value:       atomic.Value{},
		}
		ww.Value.Store(generator.NewEmptyBucket())
		ret.data[i] = ww
		startTime += uint64(bucketLengthInMs)
	}

	// calculate base address for real data array
	sliHeader := (*util.SliceHeader)(unsafe.Pointer(&ret.data))
	ret.base = unsafe.Pointer((**BucketWrap)(unsafe.Pointer(sliHeader.Data)))
	return ret
}

// NewAtomicBucketWrapArray creates an AtomicBucketWrapArray and initializes data of each BucketWrap.
//
// The len represents the length of the circular array.
// The bucketLengthInMs represents bucket length of each bucket (in milliseconds).
// The generator accepts a BucketGenerator to generate and refresh buckets.
func NewAtomicBucketWrapArray(len int, bucketLengthInMs uint32, generator BucketGenerator) *AtomicBucketWrapArray {
	return NewAtomicBucketWrapArrayWithTime(len, bucketLengthInMs, util.CurrentTimeMillis(), generator)
}

func (aa *AtomicBucketWrapArray) elementOffset(idx int) (unsafe.Pointer, bool) {
	if idx >= aa.length || idx < 0 {
		logging.Error(errors.New("array index out of bounds"),
			"array index out of bounds in AtomicBucketWrapArray.elementOffset()",
			"idx", idx, "arrayLength", aa.length)
		return nil, false
	}
	basePtr := aa.base
	return unsafe.Pointer(uintptr(basePtr) + uintptr(idx)*unsafe.Sizeof(basePtr)), true
}

func (aa *AtomicBucketWrapArray) get(idx int) *BucketWrap {
	// aa.elementOffset(idx) return the secondary pointer of BucketWrap, which is the pointer to the aa.data[idx]
	// then convert to (*unsafe.Pointer)
	if offset, ok := aa.elementOffset(idx); ok {
		return (*BucketWrap)(atomic.LoadPointer((*unsafe.Pointer)(offset)))
	}
	return nil
}

func (aa *AtomicBucketWrapArray) compareAndSet(idx int, except, update *BucketWrap) bool {
	// aa.elementOffset(idx) return the secondary pointer of BucketWrap, which is the pointer to the aa.data[idx]
	// then convert to (*unsafe.Pointer)
	// update secondary pointer
	if offset, ok := aa.elementOffset(idx); ok {
		return atomic.CompareAndSwapPointer((*unsafe.Pointer)(offset), unsafe.Pointer(except), unsafe.Pointer(update))
	}
	return false
}

// LeapArray represents the fundamental implementation of a sliding window data-structure.
//
// Some important attributes: the sampleCount represents the number of buckets,
// while intervalInMs represents the total time span of the sliding window.
//
// For example, assuming sampleCount=5, intervalInMs is 1000ms, so the bucketLength is 200ms.
// Let's give a diagram to illustrate.
// Suppose current timestamp is 1188, bucketLength is 200ms, intervalInMs is 1000ms, then
// time span of current bucket is [1000, 1200). The representation of the underlying structure:
//
//	 B0       B1      B2     B3      B4
//	 |_______|_______|_______|_______|_______|
//	1000    1200    1400     1600     1800    (1000) ms
//	       ^
//	    time=1188
type LeapArray struct {
	bucketLengthInMs uint32
	// sampleCount represents the number of BucketWrap.
	sampleCount uint32
	// intervalInMs represents the total time span of the sliding window (in milliseconds).
	intervalInMs uint32
	// array represents the internal circular array.
	array *AtomicBucketWrapArray
	// updateLock is the internal lock for update operations.
	updateLock mutex
}

func NewLeapArray(sampleCount uint32, intervalInMs uint32, generator BucketGenerator) (*LeapArray, error) {
	if sampleCount == 0 || intervalInMs%sampleCount != 0 {
		return nil, errors.Errorf("Invalid parameters, intervalInMs is %d, sampleCount is %d", intervalInMs, sampleCount)
	}
	if generator == nil {
		return nil, errors.Errorf("Invalid parameters, BucketGenerator is nil")
	}
	bucketLengthInMs := intervalInMs / sampleCount
	return &LeapArray{
		bucketLengthInMs: bucketLengthInMs,
		sampleCount:      sampleCount,
		intervalInMs:     intervalInMs,
		array:            NewAtomicBucketWrapArray(int(sampleCount), bucketLengthInMs, generator),
	}, nil
}

func (la *LeapArray) CurrentBucket(bg BucketGenerator) (*BucketWrap, error) {
	return la.currentBucketOfTime(util.CurrentTimeMillis(), bg)
}

func (la *LeapArray) currentBucketOfTime(now uint64, bg BucketGenerator) (*BucketWrap, error) {
	if now <= 0 {
		return nil, errors.New("Current time is less than 0.")
	}
	idx := la.calculateTimeIdx(now)
	bucketStart := calculateStartTime(now, la.bucketLengthInMs)

	for { // 自旋以获取当前的 BucketWrap
		old := la.array.get(idx)
		if old == nil {
			// 因为 la.array.data 在 new la.array 时就已经初始化了
			// 理论上，这里是无法到达的
			newWrap := &BucketWrap{BucketStart: bucketStart, Value: atomic.Value{}}
			newWrap.Value.Store(bg.NewEmptyBucket())
			if la.array.compareAndSet(idx, nil, newWrap) {
				return newWrap, nil
			} else {
				runtime.Gosched()
			}
		} else if bucketStart == atomic.LoadUint64(&old.BucketStart) {
			return old, nil
		} else if bucketStart > atomic.LoadUint64(&old.BucketStart) {
			// 当前时间已经是 LeapArray 的下一个周期，并且 LeapArray 不计入上一个周期。
			// 重置 BucketWrap
			if la.updateLock.TryLock() {
				old = bg.ResetBucketTo(old, bucketStart)
				la.updateLock.Unlock()
				return old, nil
			} else {
				runtime.Gosched()
			}
		} else if bucketStart < atomic.LoadUint64(&old.BucketStart) {
			// 当前时系统时间由于某种原因调到了过去时间，此时重置当前窗口.
			// 以便统计从过去重新计算
			if la.updateLock.TryLock() {
				old = bg.ResetBucketTo(old, bucketStart)
				la.updateLock.Unlock()
				return old, nil
			}
			if la.sampleCount == 1 {
				// if sampleCount==1 in leap array, in concurrency scenario, this case is possible
				return old, nil
			}
			// TODO: reserve for some special case (e.g. when occupying "future" buckets).
			return nil, errors.New(fmt.Sprintf("Provided time timeMillis=%d is already behind old.BucketStart=%d.", bucketStart, old.BucketStart))
		}
	}
}

func (la *LeapArray) calculateTimeIdx(now uint64) int {
	timeId := now / uint64(la.bucketLengthInMs)
	return int(timeId) % la.array.length
}

// Values returns all valid (non-expired) buckets between [curBucketEnd-windowInterval, curBucketEnd],
// where curBucketEnd=curBucketStart+bucketLength.
func (la *LeapArray) Values() []*BucketWrap {
	return la.valuesWithTime(util.CurrentTimeMillis())
}
func (la *LeapArray) valuesWithTime(now uint64) []*BucketWrap {
	if now <= 0 {
		return make([]*BucketWrap, 0)
	}
	ret := make([]*BucketWrap, 0, la.array.length)
	for i := 0; i < la.array.length; i++ {
		ww := la.array.get(i)
		if ww == nil || la.isBucketDeprecated(now, ww) {
			continue
		}
		ret = append(ret, ww)
	}
	return ret
}

// ValuesConditional 返回 startTimestamp 满足给定时间戳条件（谓词）的所有桶。
// 该函数使用参数“now”作为目标时间戳。
func (la *LeapArray) ValuesConditional(now uint64, predicate base.TimePredicate) []*BucketWrap {
	if now <= 0 {
		return make([]*BucketWrap, 0)
	}
	ret := make([]*BucketWrap, 0, la.array.length)
	for i := 0; i < la.array.length; i++ {
		ww := la.array.get(i)
		if ww == nil || la.isBucketDeprecated(now, ww) || !predicate(atomic.LoadUint64(&ww.BucketStart)) {
			continue
		}
		ret = append(ret, ww)
	}
	return ret
}

// isBucketDeprecated 根据给定的时间戳检查 BucketWrap 是否过期
func (la *LeapArray) isBucketDeprecated(now uint64, ww *BucketWrap) bool {
	ws := atomic.LoadUint64(&ww.BucketStart)
	return (now - ws) > uint64(la.intervalInMs)
}

// BucketGenerator represents the "generic" interface for generating and refreshing buckets.
type BucketGenerator interface {
	// NewEmptyBucket creates new raw data inside the bucket.
	NewEmptyBucket() interface{}

	// ResetBucketTo refreshes the BucketWrap to provided startTime and resets all data inside the given bucket.
	ResetBucketTo(bucket *BucketWrap, startTime uint64) *BucketWrap
}
