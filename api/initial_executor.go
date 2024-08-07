package api

import (
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/core/flow"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/spi"
	"git.garena.com/honggang.liu/seamiter-go/transport/http"
)

func defaultRegister() {
	spi.Register(http.GetCommandCenterInitFuncInst())   //负责接受dashboard请求，推模式，现在不需要使用
	spi.Register(http.GetHeartBeatSenderInitFuncInst()) //与dashboard保持心跳
	spi.Register(http.GetFetchRuleInitFuncInst())       //主动去dashboard 拉去业务规则，拉模式
	spi.Register(http.GetSendRspInitFuncInst())         //上报资源某次最新请求的响应体给dashboard
	// 发送请求体
	spi.Register(http.GetSendRequestInitFuncInst()) //上报资源某次请求的请求体响应给dashboard
	spi.Register(http.GetSendMetricInitFuncInst())  //上报服务指标
	// 默认持久化加载
	spi.Register(http.GetDefaultDatasourceInitFuncInst()) //服务启动时将rule从文件(或其他方式）加载到内存
	// 默认TokenService
	spi.Register(flow.GetDefaultTokenServiceInst())
}

func defaultDestroy() {
	spi.RegisterDestroy(http.GetHeartBeatSenderInitFuncInst(), 0)
}

// doInit 初始化
func doInit() error {
	defaultRegister()
	var funcs []spi.InitialFunc
	for _, v := range spi.GetAllRegisterFunc() {
		funcs = append(funcs, v)
	}
	for _, fun := range funcs {
		err := fun.Initial()
		if err != nil {
			logging.Warn("[InitExecutor] WARN: Initialization failed", "err", err)
			return err
		} else {
			logging.Info("[InitExecutor] Executing {} with order {}", "funName", fun, "order", fun.Order())
		}

	}

	//fun.Initial和fun.ImmediatelyLoadOnce 拆分为两个for 循环 避免相互影响
	for _, fun := range funcs {
		// 如果配置了立即拉取配置文件，则立刻拉取一次，拉取失败将会直接抛出异常
		// 立即加载的原因，是考虑到部分配置在启动前就需要加载，否则会导致不可预期的问题
		if config.ImmediatelyFetch() {
			err := fun.ImmediatelyLoadOnce()
			if err != nil {
				logging.Warn("[InitExecutor] WARN: ImmediatelyLoadOnce failed", "err", err)
				return err
			}
		}
	}

	return nil
}
