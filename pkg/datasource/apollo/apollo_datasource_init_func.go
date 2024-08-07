package apollo

import (
	"github.com/liuhailove/gmiter/api"
	"github.com/liuhailove/gmiter/constants"
	"github.com/liuhailove/gmiter/core/config"
	"github.com/liuhailove/gmiter/logging"
	"github.com/liuhailove/gmiter/util"
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
