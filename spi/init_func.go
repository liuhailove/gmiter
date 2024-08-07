package spi

import (
	"git.garena.com/honggang.liu/seamiter-go/constants"
	"git.garena.com/honggang.liu/seamiter-go/core/base"
)

// InitialFunc 初始化Func
type InitialFunc interface {
	//Initial 初始化
	Initial() error
	//Order 排序
	Order() int

	// ImmediatelyLoadOnce 立即加载一次
	ImmediatelyLoadOnce() error

	// GetRegisterType 获取注册类型
	GetRegisterType() constants.RegisterType
}

// InitialTokenServiceFunc 初始化TokenServiceFunc
type InitialTokenServiceFunc interface {
	InitialFunc
	// GetTokenService 获取TokenService服务
	GetTokenService() base.TokenService
	// ReInitial 重新初始化
	ReInitial() error
}

var (
	// 初始化Map
	initFuncMap = make(map[constants.RegisterType]InitialFunc)
)

func Register(initialFunc InitialFunc) {
	var hasRegisteredFunc = initFuncMap[initialFunc.GetRegisterType()]
	if hasRegisteredFunc == nil || hasRegisteredFunc.Order() < initialFunc.Order() {
		initFuncMap[initialFunc.GetRegisterType()] = initialFunc
	}
}

// GetRegisterTokenServiceInst 获取注册实例
func GetRegisterTokenServiceInst(registerType constants.RegisterType) InitialTokenServiceFunc {
	if initialFunc, ok := initFuncMap[registerType]; ok {
		if inst, ok2 := initialFunc.(InitialTokenServiceFunc); ok2 {
			return inst
		}
		return nil
	}
	return nil
}

func GetAllRegisterFunc() []InitialFunc {
	var funs = make([]InitialFunc, 0, len(initFuncMap))
	for _, fun := range initFuncMap {
		funs = append(funs, fun)
	}
	return funs
}
