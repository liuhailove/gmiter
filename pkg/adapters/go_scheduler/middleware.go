package go_scheduler

import (
	"context"
	sea "git.garena.com/honggang.liu/seamiter-go/api"
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	gs "git.garena.com/shopee/loan-service/credit_backend/fast-escrow/go-scheduler-executor-go"
)

// SeaMiddleware 返回 gs.TaskWrapper
// 默认以jobId进行限流，如果你的限流方式不一样，可以自定义资源抽取方法
func SeaMiddleware(opts ...Option) gs.TaskWrapper {
	options := evaluateOptions(opts)
	return func(taskFunc gs.TaskFunc) gs.TaskFunc {
		return func(ctx context.Context, param *gs.RunReq) ([]string, error) {
			if !config.CloseAll() {
				resourceName := "JobId:" + gs.Int64ToStr(param.JobID)
				if options.resourceExtract != nil {
					resourceName = options.resourceExtract(param)
				}
				var entry, blockErr = sea.Entry(
					resourceName,
					sea.WithResourceType(base.ResTypeTask),
					sea.WithTrafficType(base.Inbound),
				)
				if blockErr != nil {
					return nil, blockErr
				}
				defer entry.Exit()
				return taskFunc(ctx, param)
			}
			return taskFunc(ctx, param)
		}
	}
}
