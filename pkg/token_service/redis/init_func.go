package redis

import (
	"git.garena.com/honggang.liu/seamiter-go/constants"
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/spi"
	"git.garena.com/honggang.liu/seamiter-go/util"
)

var (
	redisTokenServiceInst *redisTokenServiceInitFunc
)

func init() {
	redisTokenServiceInst = new(redisTokenServiceInitFunc)
	spi.Register(redisTokenServiceInst)
	spi.RegisterDestroy(redisTokenServiceInst, 100)
}

type redisTokenServiceInitFunc struct {
	isInitialized util.AtomicBool
	tokenService  *RedisClusterTokenService
}

func (r *redisTokenServiceInitFunc) Initial() error {
	if !r.isInitialized.CompareAndSet(false, true) {
		return nil
	}
	if config.CloseAll() {
		logging.Warn("[redisTokenServiceInitFunc] WARN: Sdk closeAll is set true")
		return nil
	}
	var err error
	r.tokenService, err = NewRedisClient(&config.RedisClusterConfig{Host: config.RedisClusterHost(), Port: config.RedisClusterPort(), Password: config.RedisClusterPassword(), Database: config.RedisClusterDatabase()})
	if err != nil {
		return err
	}
	return nil
}

func (r *redisTokenServiceInitFunc) Order() int {
	return 100
}

func (r *redisTokenServiceInitFunc) ImmediatelyLoadOnce() error {
	return nil
}

func (r *redisTokenServiceInitFunc) GetRegisterType() constants.RegisterType {
	return constants.RedisTokenServiceType
}

// GetTokenService 获取TokenService服务
func (r *redisTokenServiceInitFunc) GetTokenService() base.TokenService {
	return r.tokenService
}

// ReInitial 重新初始化
func (r *redisTokenServiceInitFunc) ReInitial() error {
	r.isInitialized.CompareAndSet(true, false)
	return r.Initial()
}

func (r *redisTokenServiceInitFunc) Destroy() {
	if r.tokenService != nil {
		r.tokenService.Destroy()
	}
}
