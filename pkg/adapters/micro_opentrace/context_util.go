package micro_opentrace

import (
	"context"
	"github.com/micro/go-micro/metadata"
)

// InjectRequestAttr 将key-value注入到context中的metadata中，一般用于请求链路上下文传递
func InjectRequestAttr(ctx context.Context, key string, value string) context.Context {
	data, ok := metadata.FromContext(ctx)
	if !ok {
		ctx = metadata.NewContext(ctx, metadata.Metadata{})
		data, _ = metadata.FromContext(ctx)
	}
	data[key] = value
	return ctx
}

// ExtractRequestAttr 将context中的metadata中的key-value提取出来
func ExtractRequestAttr(ctx context.Context, key string) (value string, ok bool) {
	data, existed := metadata.FromContext(ctx)
	if !existed {
		return "", false
	}
	value, ok = data[key]
	return
}
