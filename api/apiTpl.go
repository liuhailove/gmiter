package api

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
)

// WithEntry 通用entry入口
// @param resource 资源名称，可以是接口名称或者自定义的代码段定义
// @param opts 附加参数，比如请求头信息、资源类型等
func WithEntry(resource string, opts ...EntryOption) *base.BlockError {
	ety, blockErr := Entry(
		resource,
		opts...,
	)
	if blockErr != nil {
		return blockErr
	}
	defer ety.Exit()
	return nil
}

// WithBlockHandlerEntry 通用entry入口
// @param resource 资源名称，可以是接口名称或者自定义的代码段定义
// @param opts 附加参数，比如请求头信息、资源类型等
func WithBlockHandlerEntry(resource string, blockErrHandler func(*base.BlockError) *base.BlockError, opts ...EntryOption) *base.BlockError {
	ety, blockErr := Entry(
		resource,
		opts...,
	)
	if blockErr != nil {
		if blockErrHandler != nil {
			return blockErrHandler(blockErr)
		}
		return blockErr
	}
	defer ety.Exit()
	return nil
}
