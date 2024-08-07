package flow

import (
	"git.garena.com/honggang.liu/seamiter-go/constants"
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/util"
)

type defaultTokenServiceInitFunc struct {
	isInitialized util.AtomicBool
	tokenService  *DefaultTokenService
}

func (r defaultTokenServiceInitFunc) Initial() error {
	if !r.isInitialized.CompareAndSet(false, true) {
		return nil
	}
	if config.CloseAll() {
		logging.Warn("[apolloDatasourceInitFunc] WARN: Sdk closeAll is set true")
		return nil
	}
	r.tokenService = NewDefaultTokenService()
	return nil
}

func (r defaultTokenServiceInitFunc) Order() int {
	return 100
}

func (r defaultTokenServiceInitFunc) ImmediatelyLoadOnce() error {
	return nil
}

func (r defaultTokenServiceInitFunc) GetRegisterType() constants.RegisterType {
	return constants.DefaultTokenServiceType
}

// GetTokenService 获取TokenService服务
func (r defaultTokenServiceInitFunc) GetTokenService() base.TokenService {
	return r.tokenService
}
func (r defaultTokenServiceInitFunc) Destroy() {
}

// ReInitial 重新初始化
func (r defaultTokenServiceInitFunc) ReInitial() error {
	r.isInitialized.CompareAndSet(true, false)
	return r.Initial()
}
func GetDefaultTokenServiceInst() *defaultTokenServiceInitFunc {
	return new(defaultTokenServiceInitFunc)
}
