package http

import (
	"github.com/liuhailove/gmiter/constants"
	"github.com/liuhailove/gmiter/core/config"
	"github.com/liuhailove/gmiter/logging"
	"github.com/liuhailove/gmiter/transport/http/metric"
	"github.com/liuhailove/gmiter/util"
	"github.com/pkg/errors"
	"runtime"
	"strconv"
	"time"
)

var (
	sendMetricInitFuncInst = new(sendMetricInitFunc)
)

type sendMetricInitFunc struct {
	isInitialized util.AtomicBool
}

func (f sendMetricInitFunc) Initial() error {
	if !f.isInitialized.CompareAndSet(false, true) {
		return nil
	}
	if config.CloseAll() {
		logging.Warn("[fetchRuleInitFunc] WARN: Sdk closeAll is set true")
		return nil
	}
	if !config.OpenConnectDashboard() {
		logging.Warn("[SendMetricInitFunc] WARN: OpenConnectDashboard is false")
		return nil
	}
	metricSender := metric.NewSimpleHttpMetricSender()
	if metricSender == nil {
		logging.Warn("[SendMetricInitFunc] WARN: No RuleCenter loaded")
		return errors.New("[SendMetricInitFunc] WARN: No RuleCenter loaded")
	}
	metricSender.BeforeStart()
	interval := f.retrieveInterval()
	//延迟5s执行，等待配置文件的初始化
	var metricTimer = time.NewTimer(time.Millisecond * 5)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				var buf [4096]byte
				n := runtime.Stack(buf[:], false)
				logging.Warn("sendMetricInitFunc worker exit from panic", "err", string(buf[:n]))
			}
		}()
		for {
			<-metricTimer.C
			metricTimer.Reset(time.Millisecond * time.Duration(interval)) //interval秒心跳防止过期
			util.Try(func() {
				_, err := metricSender.SendMetric()
				if err != nil {
					logging.Warn("[SendMetricInitFunc] WARN: SendMetric error", "err", err.Error())
				}
			}).CatchAll(func(err error) {
				logging.Error(err, "[SendMetricInitFunc] WARN: SendMetric error", "err", err.Error())
			})

		}
	}()
	return nil
}

func (f sendMetricInitFunc) Order() int {
	return 1
}

func (f sendMetricInitFunc) ImmediatelyLoadOnce() error {
	return nil
}

// GetRegisterType 获取注册类型
func (f sendMetricInitFunc) GetRegisterType() constants.RegisterType {
	return constants.SendMetricType
}

func (f sendMetricInitFunc) retrieveInterval() uint64 {
	intervalInConfig := config.SendMetricIntervalMs()
	if intervalInConfig > 0 {
		logging.Info("[SendMetricInitFunc] Using fetch rule interval in sea config property: " + strconv.FormatUint(intervalInConfig, 10))
		return intervalInConfig
	}
	logging.Info("[SendMetricInitFunc] Fetch interval not configured in config property or invalid, using sender default: " + strconv.FormatUint(config.DefaultFetchRuleIntervalMs, 10))
	return config.DefaultSendIntervalMs
}

func GetSendMetricInitFuncInst() *sendMetricInitFunc {
	return sendMetricInitFuncInst
}
