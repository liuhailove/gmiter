package apollo

import (
	"git.garena.com/honggang.liu/seamiter-go/api"
	"git.garena.com/honggang.liu/seamiter-go/constants"
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/util"
)

var (
	apolloDatasourceInitFuncInst = new(apolloDatasourceInitFunc)
)

func init() {
	api.Register(apolloDatasourceInitFuncInst)
}

type apolloDatasourceInitFunc struct {
	isInitialized util.AtomicBool
}

func (a apolloDatasourceInitFunc) GetRegisterType() constants.RegisterType {
	return constants.PersistenceDatasourceType
}

func (a apolloDatasourceInitFunc) Initial() error {
	if !a.isInitialized.CompareAndSet(false, true) {
		return nil
	}
	if config.CloseAll() {
		logging.Warn("[apolloDatasourceInitFunc] WARN: Sdk closeAll is set true")
		return nil
	}
	// 默认持久化加载
	Initialize()
	return nil
}

func (a apolloDatasourceInitFunc) Order() int {
	return 101
}

func (a apolloDatasourceInitFunc) ImmediatelyLoadOnce() error {
	return nil
}
