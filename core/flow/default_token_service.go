package flow

import (
	jsoniter "github.com/json-iterator/go"

	"github.com/liuhailove/gmiter/core/base"
)

var (
	jsonTraffic = jsoniter.ConfigCompatibleWithStandardLibrary
)

// DefaultTokenService 默认TokenService集群实现
type DefaultTokenService struct {
}

func NewDefaultTokenService() *DefaultTokenService {
	return &DefaultTokenService{}
}

func (d *DefaultTokenService) RequestToken(rule string, acquireCount uint32, prioritized int32) *base.TokResult {
	var ru = new(Rule)
	var err = jsonTraffic.Unmarshal([]byte(rule), ru)
	if err != nil {
		return base.BadResult
	}
	return acquireClusterToken(ru, acquireCount, prioritized)
}

func (d *DefaultTokenService) RequestParamToken(rule string, acquireCount uint32, params []interface{}) *base.TokResult {
	panic("implement me")
}

func (d *DefaultTokenService) RequestConcurrentToken(rule string, acquireCount uint32) *base.TokResult {
	panic("implement me")
}

func (d *DefaultTokenService) ReleaseConcurrentToken(rule string, tokenId string) {
	//TODO implement me
	panic("implement me")
}

func (d *DefaultTokenService) notValidRequestSimple(id string, count uint32) bool {
	return id == "" || count <= 0
}

func (d *DefaultTokenService) notValidRequest(address string, id string, count uint32) bool {
	return address == "" || id == "" || count <= 0
}
