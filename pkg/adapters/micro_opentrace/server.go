package micro_opentrace

import (
	"context"
	"encoding/json"
	"fmt"
	sea "git.garena.com/honggang.liu/seamiter-go/api"
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/ext/micro"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/client/grpc"
	micro_error "github.com/micro/go-micro/errors"
	"github.com/micro/go-micro/metadata"
	"github.com/micro/go-micro/server"
	"github.com/pkg/errors"
	"strings"
)

const (
	DefaultGrpcPort = 0
)

var (
	ErrBlockedByGray = errors.New("error blocked by gray")
)

// NewHandlerWrapper returns a Handler Wrapper with  sea breaker
func NewHandlerWrapper(seaOpts ...Option) server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			if !config.CloseAll() {
				resourceName := req.Service() + "." + req.Endpoint()
				opts := evaluateOptions(seaOpts)
				if opts.serverResourceExtract != nil {
					resourceName = opts.serverResourceExtract(ctx, req)
				}
				metaDataMap := make(map[string]string, 0)
				metaData, ok := metadata.FromContext(ctx)
				// 来源服务名称
				var fromService string
				if ok {
					for k, v := range metaData {
						metaDataMap[k] = v
						if k == "Micro-From-Service" {
							fromService = v
						}
					}
				}
				entry, blockErr := sea.Entry(
					resourceName,
					sea.WithResourceType(base.ResTypeMicro),
					sea.WithTrafficType(base.Inbound),
					sea.WithArgs(req.Body()),
					sea.WithRsps(rsp),
					sea.WithMetaData(metaDataMap),
					sea.WithFromService(fromService))
				if blockErr != nil {
					if blockErr.BlockType() == base.BlockTypeMock {
						if strVal, ok := blockErr.TriggeredValue().(string); ok {
							err := json.Unmarshal([]byte(strVal), rsp)
							if err != nil {
								sea.TraceError(entry, err)
							}
							addTrace(opts, ctx, req.Endpoint(), req.Body(), strVal, false)
							return err
						}
						addTrace(opts, ctx, req.Endpoint(), req.Body(), blockErr, false)
						return blockErr
					}
					if blockErr.BlockType() == base.BlockTypeMockError {
						if strVal, ok := blockErr.TriggeredValue().(string); ok {
							addTrace(opts, ctx, req.Endpoint(), req.Body(), strVal, true)
							return errors.New(strVal)
						}
						addTrace(opts, ctx, req.Endpoint(), req.Body(), blockErr, true)
						return blockErr
					}
					if opts.serverBlockFallback != nil {
						return opts.serverBlockFallback(ctx, req, blockErr)
					}
					return blockErr
				}
				defer entry.Exit()
				// 命中了灰度规则
				if entry.GrayResource() != nil {
					var val, ok = ExtractRequestAttr(ctx, "sea_redirect")
					if ok && val == "sea_redirect" {
						md, _ := metadata.FromContext(ctx)
						// 移除转发标记
						newMd := metadata.Copy(md)
						delete(newMd, "sea_redirect")
						ctx = metadata.NewContext(ctx, newMd)
						// 表示已经被转发过，不应该在走灰度逻辑，只需要在本地处理
						goto HandleLabel
					}
					if entry.LinkPass() {
						md, success := metadata.FromContext(ctx)
						if success {
							newMd := metadata.Copy(md)
							newMd["grayTag"] = entry.GrayTag()
							ctx = metadata.NewContext(ctx, newMd)
						}
					}
					if len(entry.GrayAddress()) > 0 {
						// 判断IP地址和当前地址一致，如果不一致，返回被灰度阻塞错误
						var localAddress string
						if micro.GetGrpcPort() > 0 {
							localAddress = fmt.Sprintf("%s:%d", config.HeartbeatClintIp(), micro.GetGrpcPort())
						} else {
							localAddress = fmt.Sprintf("%s:%d", config.HeartbeatClintIp(), DefaultGrpcPort)
						}
						var eqLocalAddress = false
						for _, grayAddr := range entry.GrayAddress() {
							// 如果本地地址等于灰度地址，则返回异常，以便上游重试到正确的IP地址上
							if localAddress == grayAddr {
								eqLocalAddress = true
								break
							}
						}
						if !eqLocalAddress {
							// 请求转发一次，正常来说应该会路由到正确的节点，
							// 如果不可以则让上游重试
							// 步骤1： 先指定转发IP,如果IP生效，则一次转发就会完成
							var err error
							var clientCallWraps []client.CallWrapper
							if len(opts.clientCallWraps) > 0 {
								clientCallWraps = opts.clientCallWraps
							}
							ctx = InjectRequestAttr(ctx, "sea_redirect", "sea_redirect")
							if req.ContentType() == client.DefaultContentType {
								// 默认使用RPC Client
								newRequest := client.NewRequest(req.Service(), req.Endpoint(), req)
								err = client.Call(ctx, newRequest, rsp, client.WithAddress(randomSort(entry.GrayAddress())...), client.WithCallWrapper(clientCallWraps...))
							} else {
								// 此处为GRPC client
								gClient := grpc.NewClient()
								newRequest := gClient.NewRequest(req.Service(), req.Endpoint(), req.Body())
								err = gClient.Call(ctx, newRequest, rsp, client.WithAddress(randomSort(entry.GrayAddress())...), client.WithCallWrapper(clientCallWraps...))
							}
							// 步骤2： 不指定IP转发，这是为了在规则中的IP失效时，可以按照正常的随机算法进行转发
							// 断言为micro error，为灰度重试
							if microErr, ok := err.(*micro_error.Error); ok {
								if microErr.Code == 500 && strings.Contains(microErr.Detail, "not found") {
									if req.ContentType() == client.DefaultContentType {
										// 默认使用RPC Client
										newRequest := client.NewRequest(req.Service(), req.Endpoint(), req)
										err = client.Call(ctx, newRequest, rsp)
									} else {
										// 此处为GRPC client
										gClient := grpc.NewClient(client.ContentType(req.ContentType()))
										newRequest := gClient.NewRequest(req.Service(), req.Endpoint(), req.Body())
										err = gClient.Call(ctx, newRequest, rsp, client.WithCallWrapper(clientCallWraps...))
									}
								}
							}
							return err
						}
					}
				}
			HandleLabel:
				err := h(ctx, req, rsp)
				if err != nil {
					sea.TraceError(entry, err)
				}
				return err
			}
			return h(ctx, req, rsp)
		}
	}
}

func NewStreamWrapper(seaOpts ...Option) server.StreamWrapper {
	return func(stream server.Stream) server.Stream {
		if !config.CloseAll() {
			resourceName := stream.Request().Service() + "." + stream.Request().Endpoint()
			opts := evaluateOptions(seaOpts)
			if opts.serverResourceExtract != nil {
				resourceName = opts.streamServerResourceExtract(stream)
			}
			entry, blockErr := sea.Entry(resourceName, sea.WithResourceType(base.ResTypeRPC), sea.WithTrafficType(base.Inbound))
			if blockErr != nil {
				if opts.serverBlockFallback != nil {
					return opts.streamServerBlockFallback(stream, blockErr)
				}
				stream.Send(blockErr)
				return stream
			}
			defer entry.Exit()
		}
		return stream
	}
}
