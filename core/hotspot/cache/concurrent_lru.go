package cache

import "sync"

// LruCacheMap 采用LRU策略缓存最常访问的热点参数
type LruCacheMap struct {
	// Not thread safe
	lru  *LRU
	lock *sync.RWMutex
}

func (c *LruCacheMap) Add(key interface{}, value *int64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.lru.Add(key, value)
	return
}

func (c *LruCacheMap) AddIfAbsent(key interface{}, value *int64) (priorValue *int64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	val := c.lru.AddIfAbsent(key, value)
	if val == nil {
		return nil
	}
	priorValue = val.(*int64)
	return
}

func (c *LruCacheMap) Get(key interface{}) (value *int64, isFound bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	val, found := c.lru.Get(key)
	if found {
		return val.(*int64), true
	}
	return nil, false
}

func (c *LruCacheMap) Remove(key interface{}) (isFound bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.lru.Remove(key)
}

func (c *LruCacheMap) Contains(key interface{}) (ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.lru.Contains(key)
}

func (c *LruCacheMap) Keys() []interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.lru.Keys()
}

func (c *LruCacheMap) Len() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.lru.Len()
}

func (c *LruCacheMap) Purge() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.lru.Purge()
}

func NewLRUCacheMap(size int) ConcurrentCounterCache {
	lru, err := NewLRU(size, nil)
	if err != nil {
		return nil
	}
	return &LruCacheMap{
		lru:  lru,
		lock: new(sync.RWMutex),
	}
}
