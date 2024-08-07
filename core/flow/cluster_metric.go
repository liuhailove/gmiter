package flow

import (
	"errors"
	"github.com/liuhailove/gmiter/core/base"
	"github.com/liuhailove/gmiter/core/stat"
)

// ClusterMetric 集群Metric
type ClusterMetric struct {
	*stat.BaseStatNode
	maxQueueingTimeNs int64
	statIntervalNs    int64
	lastPassedTime    int64
	lastSystemTime    int64
}

// NewClusterMetric 新建集群Metric
func NewClusterMetric(sampleCount, intervalInMs uint32, maxQueueingTimeNs, statIntervalNs int64) *ClusterMetric {
	return &ClusterMetric{stat.NewBaseStatNode(sampleCount, intervalInMs), maxQueueingTimeNs, statIntervalNs, 0, 0}
}

// NewClusterMetricWithCheck 新建集群Metric
func NewClusterMetricWithCheck(sampleCount, intervalInMs uint32) (*ClusterMetric, error) {
	if sampleCount <= 0 {
		return nil, errors.New("sampleCount should be positive")
	}
	if intervalInMs <= 0 {
		return nil, errors.New("interval should be positive")
	}
	if intervalInMs%sampleCount != 0 {
		return nil, errors.New("time span needs to be evenly divided")
	}
	return &ClusterMetric{stat.NewBaseStatNode(sampleCount, intervalInMs), 0, 0, 0, 0}, nil
}

func (c *ClusterMetric) GetAvg(event base.MetricEvent) float64 {
	return c.BaseStatNode.GetAvg(event)
}
