package cache

// ConcurrentCounterCache 缓存热点参数
type ConcurrentCounterCache interface {
	// Add 向缓存添加一个值，
	// 更新密钥的“最近使用”状态。
	Add(key interface{}, value *int64)

	// AddIfAbsent 如果缓存中不存在该键，则将值添加到缓存中，然后返回 nil。并更新密钥的“最近使用”状态
	// 如果该键已经存在于缓存中，则不执行任何操作并返回先前的值
	AddIfAbsent(key interface{}, value *int64) (priorValue *int64)

	// Get 从缓存中返回键的值并更新键的“最近使用”状态。
	Get(key interface{}) (value *int64, isFound bool)

	// Remove 从缓存中移除
	// 如果包含该键则返回 true
	Remove(key interface{}) (isFound bool)

	// Contains 检查缓存中是否存在某个键
	// 不更新最近性。
	Contains(key interface{}) (ok bool)

	// Keys 返回缓存中键的一部分，从最旧到最新。
	Keys() []interface{}

	// Len 返回缓存中的条目长度
	Len() int

	// Purge 清空缓存中的全部值
	Purge()
}
