package microv4_opentrace

//
//import (
//	"context"
//	"encoding/json"
//	sea "github.com/liuhailove/gmiter/api"
//	"github.com/liuhailove/gmiter/core/base"
//	"github.com/liuhailove/gmiter/core/config"
//	"github.com/pkg/errors"
//	"go-micro.dev/v4/metadata"
//	"go-micro.dev/v4/server"
//	"regexp"
//	"strings"
//)
//
//// NewHandlerWrapper returns a Handler Wrapper with  sea breaker
//func NewHandlerWrapper(seaOpts ...Option) server.HandlerWrapper {
//	return func(h server.HandlerFunc) server.HandlerFunc {
//		return func(ctx context.Context, req server.Request, rsp interface{}) error {
//			if !config.CloseAll() {
//				resourceName := req.Service() + "." + req.Endpoint()
//				opts := evaluateOptions(seaOpts)
//				if opts.serverResourceExtract != nil {
//					resourceName = opts.serverResourceExtract(ctx, req)
//				}
//				metaDataMap := make(map[string]string, 0)
//				metaData, ok := metadata.FromContext(ctx)
//				if ok {
//					re := regexp.MustCompile(`\b\w`)
//					for k, v := range metaData {
//						metaDataMap[k] = v
//						// 首字母切换为小写，为了兼容micro的配置
//						metaDataMap[re.ReplaceAllStringFunc(k, strings.ToLower)] = v
//					}
//				}
//				entry, blockErr := sea.Entry(
//					resourceName,
//					sea.WithResourceType(base.ResTypeMicro),
//					sea.WithTrafficType(base.Inbound),
//					sea.WithArgs(req.Body()),
//					sea.WithRsps(rsp),
//					sea.WithMetaData(metaDataMap))
//				if blockErr != nil {
//					if blockErr.BlockType() == base.BlockTypeMock {
//						if strVal, ok := blockErr.TriggeredValue().(string); ok {
//							err := json.Unmarshal([]byte(strVal), rsp)
//							if err != nil {
//								sea.TraceError(entry, err)
//							}
//							addTrace(opts, ctx, req.Endpoint(), req.Body(), strVal, false)
//							return err
//						}
//						addTrace(opts, ctx, req.Endpoint(), req.Body(), blockErr, false)
//						return blockErr
//					}
//					if blockErr.BlockType() == base.BlockTypeMockError {
//						if strVal, ok := blockErr.TriggeredValue().(string); ok {
//							addTrace(opts, ctx, req.Endpoint(), req.Body(), strVal, true)
//							return errors.New(strVal)
//						}
//						addTrace(opts, ctx, req.Endpoint(), req.Body(), blockErr, true)
//						return blockErr
//					}
//					if opts.serverBlockFallback != nil {
//						return opts.serverBlockFallback(ctx, req, blockErr)
//					}
//					return blockErr
//				}
//				defer entry.Exit()
//				err := h(ctx, req, rsp)
//				if err != nil {
//					sea.TraceError(entry, err)
//				}
//				return err
//			}
//			return h(ctx, req, rsp)
//		}
//	}
//}
//
//func NewStreamWrapper(seaOpts ...Option) server.StreamWrapper {
//	return func(stream server.Stream) server.Stream {
//		if !config.CloseAll() {
//			resourceName := stream.Request().Service() + "." + stream.Request().Endpoint()
//			opts := evaluateOptions(seaOpts)
//			if opts.serverResourceExtract != nil {
//				resourceName = opts.streamServerResourceExtract(stream)
//			}
//			entry, blockErr := sea.Entry(resourceName, sea.WithResourceType(base.ResTypeRPC), sea.WithTrafficType(base.Inbound))
//			if blockErr != nil {
//				if opts.serverBlockFallback != nil {
//					return opts.streamServerBlockFallback(stream, blockErr)
//				}
//				stream.Send(blockErr)
//				return stream
//			}
//			defer entry.Exit()
//		}
//		return stream
//	}
//}
