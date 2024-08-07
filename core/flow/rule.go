package flow

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"

	"git.garena.com/honggang.liu/seamiter-go/util"
)

// RelationStrategy 表示基于调用关系的流控策略。
type RelationStrategy int32

const (
	// CurrentResource 表示直接通过当前资源进行流量控制。
	CurrentResource RelationStrategy = iota
	// AssociatedResource 表示由关联资源而不是当前资源进行流量控制。
	AssociatedResource
)

// TokenServerStrategy TokenServer策略
type TokenServerStrategy int32

const (
	TokenServerStrategyNodeSelectMaster       TokenServerStrategy = 1
	TokenServerStrategyIndependentTokenServer TokenServerStrategy = 2
)

func (s RelationStrategy) String() string {
	switch s {
	case CurrentResource:
		return "CurrentResource"
	case AssociatedResource:
		return "AssociatedResource"
	default:
		return "Undefined"
	}
}

type TokenCalculateStrategy int32

const (
	Direct TokenCalculateStrategy = iota
	WarmUp
	MemoryAdaptive
)

func (s TokenCalculateStrategy) String() string {
	switch s {
	case Direct:
		return "Direct"
	case WarmUp:
		return "WarmUp"
	case MemoryAdaptive:
		return "MemoryAdaptive"
	default:
		return "Undefined"
	}
}

// ControlBehavior 定义请求达到资源容量时的行为。
type ControlBehavior int32

const (
	Reject ControlBehavior = iota
	// Throttling 表示挂起的请求将被限制，在队列中等待（直到有可用容量）
	Throttling
)

func (s ControlBehavior) String() string {
	switch s {
	case Reject:
		return "Reject"
	case Throttling:
		return "Throttling"
	default:
		return "Undefined"
	}
}

// ClusterStrategy 集群策略
type ClusterStrategy int32

const (
	ThresholdAvgLocal    ClusterStrategy = 1
	ThresholdGlobal      ClusterStrategy = 2
	ThresholdGlobalRedis ClusterStrategy = 3
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
	// FLOW_THRESHOLD_GLOBAL_REDIS-3： Redis全局限流，此策略下向RedisTokenServer获取Token
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

// Rule 描述流控策略，流控策略基于QPS统计指标
type Rule struct {
	// ID 表示规则的唯一 ID（可选）。
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
	Resource               string                 `json:"resource"`
	TokenCalculateStrategy TokenCalculateStrategy `json:"tokenCalculateStrategy"`
	ControlBehavior        ControlBehavior        `json:"controlBehavior"`
	// Threshold means the threshold during StatIntervalInMs
	// If StatIntervalInMs is 1000(1 second), Threshold means QPS
	Threshold        float64          `json:"threshold"`
	RelationStrategy RelationStrategy `json:"relationStrategy"`
	RefResource      string           `json:"refResource"`
	// MaxQueueingTimeMs only takes effect when ControlBehavior is Throttling.
	// When MaxQueueingTimeMs is 0, it means Throttling only controls interval of requests,
	// and requests exceeding the threshold will be rejected directly.
	MaxQueueingTimeMs uint32 `json:"maxQueueingTimeMs"`
	// 预热时间
	WarmUpPeriodSec uint32 `json:"warmUpPeriodSec"`
	// 预热期内的令牌生产减缓因子，固定值3
	WarmUpColdFactor uint32 `json:"warmUpColdFactor"`
	// StatIntervalInMs indicates the statistic interval and it's the optional setting for flow Rule.
	// If user doesn't set StatIntervalInMs, that means using default metric statistic of resource.
	// If the StatIntervalInMs user specifies can not reuse the global statistic of resource,
	// sea will generate independent statistic structure for this rule.
	StatIntervalInMs uint32 `json:"statIntervalInMs"`

	// adaptive flow control algorithm related parameters'
	// limitation: LowMemUsageThreshold > HighMemUsageThreshold && MemHighWaterMarkBytes > MemLowWaterMarkBytes
	// if the current memory usage is less than or equals to MemLowWaterMarkBytes, threshold == LowMemUsageThreshold
	// if the current memory usage is more than or equals to MemHighWaterMarkBytes, threshold == HighMemUsageThreshold
	// if  the current memory usage is in (MemLowWaterMarkBytes, MemHighWaterMarkBytes), threshold is in (HighMemUsageThreshold, LowMemUsageThreshold)
	LowMemUsageThreshold  float64 `json:"lowMemUsageThreshold"`
	HighMemUsageThreshold float64 `json:"highMemUsageThreshold"`
	MemLowWaterMarkBytes  int64   `json:"memLowWaterMarkBytes"`
	MemHighWaterMarkBytes int64   `json:"memHighWaterMarkBytes"`

	// 集群模式：true，为集群，否则为单机限流
	ClusterMode bool `json:"clusterMode"`
	// ClusterConfig 集群配置
	ClusterConfig *ClusterConfig `json:"clusterConfig"`
}

func (r *Rule) isEqualsTo(newRule *Rule) bool {
	if newRule == nil {
		return false
	}
	if !(r.Resource == newRule.Resource &&
		r.RelationStrategy == newRule.RelationStrategy &&
		r.RefResource == newRule.RefResource &&
		r.StatIntervalInMs == newRule.StatIntervalInMs &&
		r.TokenCalculateStrategy == newRule.TokenCalculateStrategy &&
		r.ControlBehavior == newRule.ControlBehavior &&
		util.Float64Equals(r.Threshold, newRule.Threshold) &&
		r.MaxQueueingTimeMs == newRule.MaxQueueingTimeMs &&
		r.WarmUpPeriodSec == newRule.WarmUpPeriodSec &&
		r.WarmUpColdFactor == newRule.WarmUpColdFactor &&
		r.LowMemUsageThreshold == newRule.LowMemUsageThreshold &&
		r.HighMemUsageThreshold == newRule.HighMemUsageThreshold &&
		r.MemLowWaterMarkBytes == newRule.MemLowWaterMarkBytes &&
		r.MemHighWaterMarkBytes == newRule.MemHighWaterMarkBytes &&
		r.LimitApp == newRule.LimitApp) {
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
	return true
}

func (r *Rule) isStatReusable(newRule *Rule) bool {
	if newRule == nil {
		return false
	}
	return r.Resource == newRule.Resource &&
		r.RelationStrategy == newRule.RelationStrategy &&
		r.RefResource == newRule.RefResource &&
		r.StatIntervalInMs == newRule.StatIntervalInMs &&
		r.needStatistic() && newRule.needStatistic()
}

func (r *Rule) needStatistic() bool {
	return r.TokenCalculateStrategy == WarmUp || r.ControlBehavior == Reject
}

func (r *Rule) isClusterMode() bool {
	return r.ClusterMode
}

func (r *Rule) String() string {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b, err := json.Marshal(r)
	if err != nil {
		// Return the fallback string
		return fmt.Sprintf("Rule{Resource=%s, TokenCalculateStrategy=%s, ControlBehavior=%s, "+
			"Count=%.2f, RelationStrategy=%s, RefResource=%s, MaxQueueingTimeMs=%d, WarmUpPeriodSec=%d, WarmUpColdFactor=%d, StatIntervalInMs=%d, "+
			"LowMemUsageThreshold=%v, HighMemUsageThreshold=%v, MemLowWaterMarkBytes=%v, MemHighWaterMarkBytes=%v,"+
			"ClusterMode=%v,ClusterConfig=%v}",
			r.Resource, r.TokenCalculateStrategy, r.ControlBehavior, r.Threshold, r.RelationStrategy, r.RefResource,
			r.MaxQueueingTimeMs, r.WarmUpPeriodSec, r.WarmUpColdFactor, r.StatIntervalInMs,
			r.LowMemUsageThreshold, r.HighMemUsageThreshold, r.MemLowWaterMarkBytes, r.MemHighWaterMarkBytes,
			r.ClusterMode, r.ClusterConfig)
	}
	return string(b)
}

func (r *Rule) ResourceName() string {
	return r.Resource
}
