package flow

//var (
//	clusterMetricMap = sync.Map{}
//)
//
//func ClearClusterMetric() {
//	clusterMetricMap = sync.Map{}
//}
//
//func PutClusterMetric(resource string, maxQueueingTimeMs uint32, statIntervalInMs uint32) bool {
//	var statIntervalNs int64
//	if statIntervalInMs == 0 {
//		statIntervalNs = 1000 * MillisToNanosOffset
//	} else {
//		statIntervalNs = int64(statIntervalInMs) * MillisToNanosOffset
//	}
//	var maxQueueingTimeNs = int64(maxQueueingTimeMs) * MillisToNanosOffset
//	var oldMetrics, ok = clusterMetricMap.Load(resource)
//	if ok {
//		m := oldMetrics.(*ClusterMetric)
//		// maxQueueTime需要取最小的存入
//		if m.maxQueueingTimeNs <= maxQueueingTimeNs && m.statIntervalNs == statIntervalNs {
//			return false
//		}
//	}
//	var metric = NewClusterMetric(config.MetricStatisticSampleCount(), config.MetricStatisticIntervalMs(), maxQueueingTimeNs, statIntervalNs)
//	clusterMetricMap.Store(resource, metric)
//	return true
//}
//
//func GetClusterMetric(resource string) *ClusterMetric {
//	if m, ok := clusterMetricMap.Load(resource); ok {
//		return m.(*ClusterMetric)
//	}
//	return nil
//}

//func RemoveClusterMetric(resource string) {
//	clusterMetricMap.Delete(resource)
//}

//// GetOrCreateClusterMetric 不存在就创建，或者直接返回创建后的对象
//func GetOrCreateClusterMetric(rule Rule) *ClusterMetric {
//	var statIntervalNs int64
//	if rule.StatIntervalInMs == 0 {
//		statIntervalNs = 1000 * MillisToNanosOffset
//	} else {
//		statIntervalNs = int64(rule.StatIntervalInMs) * MillisToNanosOffset
//	}
//	var clusterMetric interface{}
//	if rule.RelationStrategy == AssociatedResource {
//		clusterMetric, _ = clusterMetricMap.LoadOrStore(rule.RefResource, NewClusterMetric(config.MetricStatisticSampleCount(), config.MetricStatisticIntervalMs(), int64(rule.MaxQueueingTimeMs)*MillisToNanosOffset, statIntervalNs))
//	} else {
//		clusterMetric, _ = clusterMetricMap.LoadOrStore(rule.Resource, NewClusterMetric(config.MetricStatisticSampleCount(), config.MetricStatisticIntervalMs(), int64(rule.MaxQueueingTimeMs)*MillisToNanosOffset, statIntervalNs))
//	}
//	return clusterMetric.(*ClusterMetric)
//}
//
//func ResetClusterMetrics() {
//	clusterMetricMap.Range(func(key, value interface{}) bool {
//		var clusterMetric, _ = NewClusterMetricWithCheck(config.MetricStatisticSampleCount(), config.MetricStatisticIntervalMs())
//		clusterMetricMap.Store(key, clusterMetric)
//		return true
//	})
//}
