package micro_opentrace

import (
	"context"
	"encoding/json"
	jsoniter "github.com/json-iterator/go"
	"github.com/micro/go-micro/metadata"
	"github.com/opentracing/opentracing-go"
)

// 增加链路追踪，主要是为了适配在Mock时依然上报到链路追踪
func addTrace(opts *options, ctx context.Context, endPoint string, reqBody interface{}, rsp interface{}, withErr bool) {
	if opts.tracer == nil {
		return
	}
	metaData, _ := metadata.FromContext(ctx)

	var spanOpts []opentracing.StartSpanOption
	// 首先从context去找
	// 如果不存在，就在metadata找
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		spanOpts = append(spanOpts, opentracing.ChildOf(parentSpan.Context()))
	} else if spanCtx, err := opts.tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(metaData)); err == nil {
		spanOpts = append(spanOpts, opentracing.ChildOf(spanCtx))
	}
	sp := opts.tracer.StartSpan(endPoint, spanOpts...)
	sp.SetTag("source-type", "seamiter-mock")
	if err := sp.Tracer().Inject(sp.Context(), opentracing.TextMap, opentracing.TextMapCarrier(metaData)); err == nil {
		ctx = opentracing.ContextWithSpan(ctx, sp)
		ctx = metadata.NewContext(ctx, metaData)
		// 放入metadata
		var metaDataByte, _ = json.Marshal(metaData)
		sp.LogKV("md", string(metaDataByte))
		// 放入request
		if requestJsonData, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(reqBody); err == nil {
			sp.LogKV("req", string(requestJsonData))
		}
		//报错的时候 将返回写到err_msg中;如果是正常返回 则写到resp中
		if withErr {
			sp.LogKV("err_msg", rsp)
		} else {
			// 放入resp
			sp.LogKV("resp", rsp)
		}
		// finish
		sp.Finish()
	}
}
