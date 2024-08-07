package go_scheduler

import gs "git.garena.com/shopee/loan-service/credit_backend/fast-escrow/go-scheduler-executor-go"

type (
	Option  func(*options)
	options struct {
		resourceExtract func(*gs.RunReq) string
	}
)

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	for _, opt := range opts {
		opt(optCopy)
	}
	return optCopy
}

// WithResourceExtractor 设置抽取go-scheduler拦截器资源的方法
func WithResourceExtractor(fn func(req *gs.RunReq) string) Option {
	return func(o *options) {
		o.resourceExtract = fn
	}
}
