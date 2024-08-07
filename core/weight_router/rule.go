package weight_router

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
)

const (
	// DefaultWightForNormalNode 正常节点的默认缺省权重为10000
	DefaultWightForNormalNode = 10000
)

type TypeForWeightRule int32

const (
	InitWeightRuleType   TypeForWeightRule = 0 //初始值 默认不使用
	ClientWeightRuleType TypeForWeightRule = 1 //客户端权重路由
	ServerWeightRuleType TypeForWeightRule = 2 //服务端权重路由
)

// Rule 描述应用路由规则
type Rule struct {
	// ID 规则唯一ID（可选）
	ID string `json:"id,omitempty"`
	// 应用名称
	App string `json:"app"`
	//客户端应用名称
	ClientAppName string `json:"clientAppName"`
	//服务端应用名称
	ServerAppName string `json:"serverAppName"`
	// 服务端服务名称
	ServerServiceName string `json:"serverServiceName"`
	// 目标地址
	TargetAddress string `json:"targetAddress"`
	// 节点权重
	Weight int64 `json:"weight"`
	// 权重规则类型，0为初始值，1为客户端权重规则，2为服务端权重规则
	WeightRuleType TypeForWeightRule `json:"weightRuleType"`
}

func (r *Rule) String() string {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b, err := json.Marshal(r)
	if err != nil {
		// Return the fallback string
		return fmt.Sprintf("{Id=%s, App=%s,ClientAppName=%s, ServerAppName=%s,ServerServiceName=%s,TargetAddress=%s, Weight=%d, WeightRuleType=%d}", r.ID, r.App, r.ClientAppName, r.ServerAppName, r.ServerServiceName, r.TargetAddress, r.Weight, r.WeightRuleType)
	}
	return string(b)
}

func (r *Rule) ResourceName() string {
	return r.TargetAddress
}

// WeightNode 用于seamiter节点权重计算
type WeightNode struct {
	Address string `json:"address"`
}
