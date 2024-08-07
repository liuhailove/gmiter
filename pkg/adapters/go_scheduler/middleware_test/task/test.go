package task

import (
	"context"
	xxl "git.garena.com/shopee/loan-service/credit_backend/fast-escrow/go-scheduler-executor-go"
)

func Test(ctx context.Context, param *xxl.RunReq) (msg []string, err error) {
	return []string{"test 1 done", "test 2 done"}, nil
}
