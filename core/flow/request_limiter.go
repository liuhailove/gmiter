package flow

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/core/stat"
)

type RequestLimiter struct {
	*stat.BaseStatNode
	qpsAllowed float64
}

// NewRequestLimiter 构造请求Limiter
func NewRequestLimiter(qpsAllowed float64) *RequestLimiter {
	return &RequestLimiter{stat.NewBaseStatNode(config.MetricStatisticSampleCount(), config.MetricStatisticIntervalMs()), qpsAllowed}
}

func (r *RequestLimiter) Add(x int64) {
	r.AddCount(base.MetricEventPass, x)
}

func (r *RequestLimiter) GetQps() float64 {
	return r.GetQPS(base.MetricEventPass)
}

func (r *RequestLimiter) GetQpsAllowed() float64 {
	return r.qpsAllowed
}

func (r *RequestLimiter) CanPass() bool {
	return r.GetQPS(base.MetricEventPass)+1 <= r.qpsAllowed
}

func (r *RequestLimiter) TryPass() bool {
	if r.CanPass() {
		r.Add(1)
		return true
	}
	return false
}
