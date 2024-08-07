package http

import (
	"git.garena.com/honggang.liu/seamiter-go/constants"
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/ext/datasource/file"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/util"
)

var (
	defaultDatasourceInitFuncInst = new(defaultDatasourceInitFunc)
)

type defaultDatasourceInitFunc struct {
	isInitialized util.AtomicBool
}

func (d defaultDatasourceInitFunc) Initial() error {
	if !d.isInitialized.CompareAndSet(false, true) {
		return nil
	}
	if config.CloseAll() {
		logging.Warn("[defaultDatasourceInitFunc] WARN: Sdk closeAll is set true")
		return nil
	}
	// 默认持久化加载
	file.Initialize()
	return nil
}

func (d defaultDatasourceInitFunc) Order() int {
	return 10
}

func (d defaultDatasourceInitFunc) ImmediatelyLoadOnce() error {
	return nil
}

// GetRegisterType 获取注册类型
func (d defaultDatasourceInitFunc) GetRegisterType() constants.RegisterType {
	return constants.PersistenceDatasourceType
}
func GetDefaultDatasourceInitFuncInst() *defaultDatasourceInitFunc {
	return defaultDatasourceInitFuncInst
}
