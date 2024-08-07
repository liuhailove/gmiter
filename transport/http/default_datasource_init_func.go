package http

import (
	"github.com/liuhailove/gmiter/constants"
	"github.com/liuhailove/gmiter/core/config"
	"github.com/liuhailove/gmiter/ext/datasource/file"
	"github.com/liuhailove/gmiter/logging"
	"github.com/liuhailove/gmiter/util"
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
