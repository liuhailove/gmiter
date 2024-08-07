package hotspot

import "github.com/liuhailove/gmiter/core/hotspot/cache"

const (
	ConcurrencyMaxCount = 4000
	ParamsCapacityBase  = 4000
	ParamsMaxCapacity   = 20000
)

// ParamsMetric 带有频繁（“热点”）参数的实时计数器。
//
// 对于每个缓存映射，键是参数值，而值是计数器。
type ParamsMetric struct {
	// RuleTimeCounter 记录最后添加的令牌时间戳。
	RuleTimeCounter cache.ConcurrentCounterCache
	// RuleTokenCounter 记录令牌的数量。
	RuleTokenCounter cache.ConcurrentCounterCache
	// ConcurrencyCounter 记录实时并发数
	ConcurrentCounter cache.ConcurrentCounterCache
}
