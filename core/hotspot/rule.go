package hotspot

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"reflect"
	"strconv"
)

// ControlBehavior indicates the traffic shaping behaviour.
type ControlBehavior int32

const (
	Reject ControlBehavior = iota
	Throttling
)

func (t ControlBehavior) String() string {
	switch t {
	case Reject:
		return "Reject"
	case Throttling:
		return "Throttling"
	default:
		return strconv.Itoa(int(t))
	}
}

// MetricType 表示目标度量类型。
type MetricType int32

const (
	// Concurrency 标识并发统计
	Concurrency MetricType = iota
	// QPS 标识每秒的统计数
	QPS
)

// ParameterSourceType 参数来源类型
type ParameterSourceType int

const (
	ParameterTypeUnknown   ParameterSourceType = 0 // 未知类型，为了兼容
	ParameterTypeParameter ParameterSourceType = 1
	ParameterTypeHeader    ParameterSourceType = 2
	ParameterTypeMetadata  ParameterSourceType = 3
)

func (p ParameterSourceType) String() string {
	switch p {
	case ParameterTypeUnknown:
		return "ParameterTypeUnknown"
	case ParameterTypeHeader:
		return "ParameterTypeHeader"
	case ParameterTypeParameter:
		return "ParameterTypeParameter"
	case ParameterTypeMetadata:
		return "ParameterTypeMetadata"
	default:
		return strconv.Itoa(int(p))
	}
}

func (t MetricType) String() string {
	switch t {
	case Concurrency:
		return "Concurrency"
	case QPS:
		return "QPS"
	default:
		return "Undefined"
	}
}

// ParamKind represents the Param kind.
type ParamKind string

const (
	KindInt     ParamKind = "int"
	KindInt32             = "int32"
	KindInt64             = "int64"
	KindString            = "string"
	KindBool              = "bool"
	KindFloat32           = "float32"
	KindFloat64           = "float64"
)

// ClusterStrategy 集群策略
type ClusterStrategy int32

const (
	ThresholdAvgLocal ClusterStrategy = 1
	ThresholdGlobal   ClusterStrategy = 2
)

func (c ClusterStrategy) String() string {
	switch c {
	case ThresholdAvgLocal:
		return "FlowThresholdAvgLocal"
	case ThresholdGlobal:
		return "FlowThresholdGlobal"
	default:
		return "Undefined"
	}
}

// ClusterConfig 集群流控配置
type ClusterConfig struct {
	// FlowId 全局流控ID
	FlowId string `json:"flowId"`

	// FallbackToLocalWhenFail 当向tokenServer请求Token失败时，是否要被回退到本地限流，
	// true: 回退到本地限流
	// false：不回退，直接通过
	FallbackToLocalWhenFail bool `json:"fallbackToLocalWhenFail"`
	// ClusterStrategy 集群策略
	// FLOW_THRESHOLD_AVG_LOCAL-1：全局均摊，此策略每个节点的阈值在server端计算
	// FLOW_THRESHOLD_GLOBAL-2： 全局限流，此策略下向TokenServer获取Token
	ClusterStrategy int32 `json:"clusterStrategy"`
	// ResourceTimeout 如果客户端保持Token的时间超过ResourceTimeout，resourceTimeoutStrategy策略将会生效
	ResourceTimeout int64 `json:"resourceTimeout"`
	// ResourceTimeoutStrategy 资源超时策略， 0-忽略， 1-释放Token
	ResourceTimeoutStrategy int32 `json:"resourceTimeoutStrategy"`
	// ClientOfflineTime 如果一个客户端下线了，tokenServer会在ClientOfflineTime后删除这个client保持的全部token
	ClientOfflineTime int64 `json:"clientOfflineTime"`
	// GlobalThreshold 全局限流值
	GlobalThreshold float64 `json:"globalThreshold"`
	// TokenServerStrategy tokenServer的选主策略，
	// TOKEN_SERVER_STRATEGY_NODE_SELECT_MASTER-1-节点选主 TOKEN_SERVER_STRATEGY_INDEPENDENT_TOKEN_SERVER-2：独立TokenServer
	TokenServerStrategy int32 `json:"tokenServerStrategy"`
	// TokenServerAddress 当为独立TokenServer是的TokenServer地址
	TokenServerAddress string `json:"tokenServerAddress"`
	// DowngradeDurationInMs 当TokenServer不可用的情况下，Client会降级为本地限流，此处指降级的时长，
	// 降级时长到之后才会重新向TokenServer请求Token
	DowngradeDurationInMs int64 `json:"downgradeDurationInMs"`
	// MasterNodeThreshold 主节点阈值，当主节点请求Token的速率超过此值时则降级
	MasterNodeThreshold float64 `json:"masterNodeThreshold"`
	// TokenServerMasterHost 选主的master host
	TokenServerMasterHost string `json:"tokenServerMasterHost"`
	// TokenServerMasterPort 选主的master port
	TokenServerMasterPort int32 `json:"tokenServerMasterPort"`
}

// Rule 代表热点（频繁）参数流控规则
type Rule struct {
	// ID 唯一ID
	ID string `json:"id,omitempty"`
	// App 规则归属的应用名称
	App string `json:"app,omitempty"`

	// RuleName 规则名称
	RuleName string `json:"ruleName,omitempty"`
	// LimitApp 限制应用程序
	// 将受来源限制的应用程序名称。
	// 默认的limitApp是{@code default}，表示允许所有源端应用。
	// 对于权限规则，多个源名称可以用逗号（','）分隔。
	LimitApp string `json:"limitApp"`
	// Resource 资源名称
	Resource string `json:"resource"`
	// MetricType 表示检查逻辑的metric类型。
	// 对于 Concurrency 指标，热点模块将检查每个热点参数的并发度，
	//     如果并发超过Threshold，则直接拒绝流量。
	// 对于 QPS 指标，热点模块会检查每个热点参数的QPS，
	//     ControlBehavior 决定流量整形控制器的行为
	MetricType MetricType `json:"metricType"`
	// ControlBehavior 标识流量整形行为。
	// 仅仅当MetricType是QPS时才会生效
	ControlBehavior ControlBehavior `json:"controlBehavior"`
	// ParamSource 参数类型
	ParamSource ParameterSourceType `json:"paramSource"`
	// ParamIdx 是上下文参数切片中的索引。
	// 如果 ParamIdx 大于或等于 0，ParamIdx 表示第 <ParamIdx> 参数
	// 如果ParamIdx为负数，则ParamIdx表示反转的第<ParamIdx>参数
	ParamIdx int `json:"paramIdx"`
	// ParamKey 是 EntryContext.Input.Attachments 映射中的键。
	// ParamKey可以作为ParamIdx的补充，方便规则从大量参数中快速获取参数
	// ParamKey与ParamIdx互斥，ParamKey比ParamIdx优先级高
	ParamKey string `json:"paramKey"`
	// 参数类型
	ParamKind ParamKind `json:"paramKind"`
	// Threshold是触发拒绝的阈值
	Threshold float64 `json:"threshold"`
	// MaxQueueingTimeMs 仅在ControlBehavior为Throttling且MetricType为QPS时生效
	MaxQueueingTimeMs int64 `json:"maxQueueingTimeMs"`
	// BurstCount 是静默计数
	// BurstCount 仅在ControlBehavior为Reject且MetricType为QPS时生效
	BurstCount int64 `json:"burstCount"`
	// DurationInSec 为统计的时间间隔
	// DurationInSec 仅在MetricType为QPS时生效
	DurationInSec int64 `json:"durationInSec"`
	// ParamsMaxCapacity 是缓存统计的最大容量
	ParamsMaxCapacity int64 `json:"ParamsMaxCapacity"`
	// SpecificItems 表示特定值的特殊阈值
	SpecificItems map[interface{}]int64 `json:"specificItems"`

	// 集群模式：true，为集群，否则为单机限流
	ClusterMode bool `json:"clusterMode"`
	// ClusterConfig 集群配置
	ClusterConfig *ClusterConfig `json:"clusterConfig"`
}

func (r *Rule) String() string {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b, err := json.Marshal(r)
	if err != nil {
		// Return the fallback string
		return fmt.Sprintf("{Id:%s, Resource:%s, MetricType:%+v, ControlBehavior:%+v, ParamSource:%+v, ParamKind:%+v, ParamIdx:%d, ParamKey:%s, Count:%f, MaxQueueingTimeMs:%d, BurstCount:%d, DurationInSec:%d, ParamsMaxCapacity:%d, ParamFlowItems:%+v，ClusterMode=%v,ClusterConfig=%v}",
			r.ID, r.Resource, r.MetricType, r.ControlBehavior, r.ParamSource, r.ParamKind, r.ParamIdx, r.ParamKey, r.Threshold, r.MaxQueueingTimeMs, r.BurstCount, r.DurationInSec, r.ParamsMaxCapacity, r.SpecificItems, r.ClusterMode, r.ClusterConfig)
	}
	return string(b)
}

func (r *Rule) ResourceName() string {
	return r.Resource
}

// IsStatReusable checks whether current rule is "statistically" equal to the given rule.
func (r *Rule) IsStatReusable(newRule *Rule) bool {
	return r.Resource == newRule.Resource && r.ControlBehavior == newRule.ControlBehavior &&
		r.ParamsMaxCapacity == newRule.ParamsMaxCapacity && r.DurationInSec == newRule.DurationInSec &&
		r.MetricType == newRule.MetricType && r.ParamSource == newRule.ParamSource && r.ParamKind == newRule.ParamKind
}

// Equals checks whether current rule is consistent with the given rule.
func (r *Rule) Equals(newRule *Rule) bool {
	baseCheck := r.Resource == newRule.Resource && r.MetricType == newRule.MetricType && r.ControlBehavior == newRule.ControlBehavior && r.ParamsMaxCapacity == newRule.ParamsMaxCapacity && r.ParamIdx == newRule.ParamIdx && r.ParamKey == newRule.ParamKey && r.Threshold == newRule.Threshold && r.DurationInSec == newRule.DurationInSec && r.LimitApp == newRule.LimitApp && reflect.DeepEqual(r.SpecificItems, newRule.SpecificItems) &&
		r.ParamSource == newRule.ParamSource &&
		r.ParamKind == newRule.ParamKind
	if !baseCheck {
		return false
	}

	if r.ClusterMode != newRule.ClusterMode {
		return false
	}

	if r.ClusterMode {
		var rClusterConfig = r.ClusterConfig
		var nClusterConfig = newRule.ClusterConfig
		if !(rClusterConfig.FlowId == nClusterConfig.FlowId &&
			rClusterConfig.FallbackToLocalWhenFail == nClusterConfig.FallbackToLocalWhenFail &&
			rClusterConfig.ClusterStrategy == nClusterConfig.ClusterStrategy &&
			rClusterConfig.ResourceTimeout == nClusterConfig.ResourceTimeout &&
			rClusterConfig.ResourceTimeoutStrategy == nClusterConfig.ResourceTimeoutStrategy &&
			rClusterConfig.ClientOfflineTime == nClusterConfig.ClientOfflineTime &&
			rClusterConfig.GlobalThreshold == nClusterConfig.GlobalThreshold &&
			rClusterConfig.TokenServerStrategy == nClusterConfig.TokenServerStrategy &&
			rClusterConfig.TokenServerAddress == nClusterConfig.TokenServerAddress &&
			rClusterConfig.DowngradeDurationInMs == nClusterConfig.DowngradeDurationInMs &&
			rClusterConfig.MasterNodeThreshold == nClusterConfig.MasterNodeThreshold &&
			rClusterConfig.TokenServerMasterHost == nClusterConfig.TokenServerMasterHost &&
			rClusterConfig.TokenServerMasterPort == nClusterConfig.TokenServerMasterPort) {
			return false
		}
	}

	if r.ControlBehavior == Reject {
		return r.BurstCount == newRule.BurstCount
	}
	if r.ControlBehavior == Throttling {
		return r.MaxQueueingTimeMs == newRule.MaxQueueingTimeMs
	}
	return false
}
