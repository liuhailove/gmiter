package etcdv3

import (
	"git.garena.com/honggang.liu/seamiter-go/api"
	"git.garena.com/honggang.liu/seamiter-go/constants"
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/util"
)

var (
	etcdV3DatasourceInitFuncInst = new(etcdV3DatasourceInitFunc)
)

func init() {
	api.Register(etcdV3DatasourceInitFuncInst)
}

type etcdV3DatasourceInitFunc struct {
	isInitialized util.AtomicBool
}

func (e etcdV3DatasourceInitFunc) GetRegisterType() constants.RegisterType {
	return constants.PersistenceDatasourceType
}

func (e etcdV3DatasourceInitFunc) Initial() error {
	if !e.isInitialized.CompareAndSet(false, true) {
		return nil
	}
	if config.CloseAll() {
		logging.Warn("[defaultDatasourceInitFunc] WARN: Sdk closeAll is set true")
		return nil
	}
	// 默认持久化加载
	Initialize()
	return nil
}

func (e etcdV3DatasourceInitFunc) Order() int {
	return 100
}

func (e etcdV3DatasourceInitFunc) ImmediatelyLoadOnce() error {
	return nil
}
