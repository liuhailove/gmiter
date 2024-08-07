package flow

import (
	"errors"
	"strings"
	"sync"

	"git.garena.com/honggang.liu/seamiter-go/core/config"
)

var (
	_globalQpsLimiterMap = make(map[string]*RequestLimiter)
	_globalLock          = new(sync.RWMutex)
)

func GlobalInitIfAbsent(namespace string) error {
	if strings.TrimSpace(namespace) == "" {
		return errors.New("namespace cannot be empty")
	}
	_globalLock.Lock()
	_globalQpsLimiterMap[namespace] = NewRequestLimiter(config.MaxAllowQps())
	_globalLock.Unlock()
	return nil
}

func GetGlobalRequestLimiter(namespace string) *RequestLimiter {
	if strings.TrimSpace(namespace) == "" {
		return nil
	}
	_globalLock.RLock()
	var rLimiter = _globalQpsLimiterMap[namespace]
	_globalLock.RUnlock()
	if rLimiter == nil {
		_globalLock.Lock()
		rLimiter = NewRequestLimiter(config.MaxAllowQps())
		_globalQpsLimiterMap[namespace] = rLimiter
		_globalLock.Unlock()
	}
	return rLimiter
}

func GetClientRequestLimiter(namespace string) *RequestLimiter {
	if strings.TrimSpace(namespace) == "" {
		return nil
	}
	_globalLock.RLock()
	var rLimiter = _globalQpsLimiterMap[namespace]
	_globalLock.RUnlock()
	if rLimiter == nil {
		_globalLock.Lock()
		rLimiter = NewRequestLimiter(config.ClientMaxAllowQps())
		_globalQpsLimiterMap[namespace] = rLimiter
		_globalLock.Unlock()
	}
	return rLimiter
}

func GlobalTryPass(namespace string) bool {
	var reqLimiter = GetGlobalRequestLimiter(namespace)
	if reqLimiter == nil {
		return true
	}
	return reqLimiter.TryPass()
}

func ClientTryPass(namespace string) bool {
	var reqLimiter = GetClientRequestLimiter(namespace)
	if reqLimiter == nil {
		return true
	}
	return reqLimiter.TryPass()
}

func GetGlobalCurrentQps(namespace string) float64 {
	var reqLimiter = GetGlobalRequestLimiter(namespace)
	if reqLimiter == nil {
		return 0
	}
	return reqLimiter.GetQps()
}

func GetGlobalMaxAllowedQps(namespace string) float64 {
	var reqLimiter = GetGlobalRequestLimiter(namespace)
	if reqLimiter == nil {
		return 0
	}
	return reqLimiter.GetQpsAllowed()
}

func ApplyGlobalMaxQpsChange(maxAllowedQps float64) error {
	if maxAllowedQps <= 0 {
		return errors.New("max allowed QPS should > 0")
	}
	_globalLock.Lock()
	for _, limiter := range _globalQpsLimiterMap {
		limiter.qpsAllowed = maxAllowedQps
	}
	_globalLock.Unlock()
	return nil
}
