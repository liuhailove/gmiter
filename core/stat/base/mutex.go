package base

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const mutexLocked = 1 << iota

// mutex 支持try-locking
type mutex struct {
	sync.Mutex
}

// TryLock 仅当调用时锁空闲时才获取锁
func (tl *mutex) TryLock() bool {
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&tl.Mutex)), 0, mutexLocked)
}
