package datasource

import (
	"fmt"
	"github.com/liuhailove/gmiter/core/hotspot"
	"github.com/liuhailove/gmiter/logging"
	"github.com/pkg/errors"
	"strconv"
)

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

type HotspotRule struct {
	// ID is the unique id
	ID string `json:"id,omitempty"`
	// LimitApp 限制应用程序
	// 将受来源限制的应用程序名称。
	// 默认的limitApp是{@code default}，表示允许所有源端应用。
	// 对于权限规则，多个源名称可以用逗号（','）分隔。
	LimitApp string `json:"limitApp"`
	// Resource is the resource name
	Resource string `json:"resource"`
	// MetricType indicates the metric type for checking logic.
	// For Concurrency metric, hotspot module will check the each hot parameter's concurrency,
	//		if concurrency exceeds the Threshold, reject the traffic directly.
	// For QPS metric, hotspot module will check the each hot parameter's QPS,
	//		the ControlBehavior decides the behavior of traffic shaping controller
	MetricType hotspot.MetricType `json:"metricType"`
	// ControlBehavior indicates the traffic shaping behaviour.
	// ControlBehavior only takes effect when MetricType is QPS
	ControlBehavior hotspot.ControlBehavior `json:"controlBehavior"`
	// ParamSource 参数类型
	ParamSource hotspot.ParameterSourceType `json:"paramSource"`
	// ParamIdx is the index in context arguments slice.
	// if ParamIdx is greater than or equals to zero, ParamIdx means the <ParamIdx>-th parameter
	// if ParamIdx is the negative, ParamIdx means the reversed <ParamIdx>-th parameter
	ParamIdx int `json:"paramIdx"`
	// 参数类型
	ParamKind hotspot.ParamKind `json:"paramKind"`
	// ParamKey is the key in EntryContext.Input.Attachments map.
	// ParamKey can be used as a supplement to ParamIdx to facilitate rules to quickly obtain parameter from a large number of parameters
	// ParamKey is mutually exclusive with ParamIdx, ParamKey has the higher priority than ParamIdx
	ParamKey string `json:"paramKey"`
	// Threshold is the threshold to trigger rejection
	Threshold float64 `json:"threshold"`
	// MaxQueueingTimeMs only takes effect when ControlBehavior is Throttling and MetricType is QPS
	MaxQueueingTimeMs int64 `json:"maxQueueingTimeMs"`
	// BurstCount is the silent count
	// BurstCount only takes effect when ControlBehavior is Reject and MetricType is QPS
	BurstCount int64 `json:"burstCount"`
	// DurationInSec is the time interval in statistic
	// DurationInSec only takes effect when MetricType is QPS
	DurationInSec int64 `json:"durationInSec"`
	// ParamsMaxCapacity is the max capacity of cache statistic
	ParamsMaxCapacity int64           `json:"paramsMaxCapacity"`
	ParamFlowItems    []ParamFlowItem `json:"paramFlowItemList"`

	// 集群模式：true，为集群，否则为单机限流
	ClusterMode bool `json:"clusterMode"`
	// ClusterConfig 集群配置
	ClusterConfig *ClusterConfig `json:"clusterConfig"`
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
	KindByte              = "byte"
	KindSum               = ""
)

// ParamFlowItem indicates the specific param, contain the supported param kind and concrete value.
type ParamFlowItem struct {
	ParamKind  ParamKind `json:"paramKind"`  // 参数类型
	ParamValue string    `json:"paramValue"` // 例外参数值，如当参数值为100时
	Threshold  int64     `json:"threshold"`  // 当参数值为Value时的阈值
}

func (s *ParamFlowItem) String() string {
	return fmt.Sprintf("ParamFlowItem: [ParamKind: %+v, Value: %s]", s.ParamKind, s.ParamValue)
}

// parseSpecificItems parses the ParamFlowItem as real value.
func parseSpecificItems(paramKind ParamKind, source []ParamFlowItem) map[interface{}]int64 {
	ret := make(map[interface{}]int64, len(source))
	if len(source) == 0 {
		return ret
	}
	for _, item := range source {
		switch paramKind {
		case KindInt:
			realVal, err := strconv.Atoi(item.ParamValue)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for int specific item", "itemValKind", item.ParamKind, "itemValStr", item.ParamValue, "itemThreshold", item.Threshold)
				continue
			}
			ret[realVal] = item.Threshold
		case KindInt32:
			realVal, err := strconv.ParseInt(item.ParamValue, 10, 32)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for int specific item", "itemValKind", item.ParamKind, "itemValStr", item.ParamValue, "itemThreshold", item.Threshold)
				continue
			}
			ret[realVal] = item.Threshold
		case KindInt64:
			realVal, err := strconv.ParseInt(item.ParamValue, 10, 64)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for int specific item", "itemValKind", item.ParamKind, "itemValStr", item.ParamValue, "itemThreshold", item.Threshold)
				continue
			}
			ret[realVal] = item.Threshold

		case KindString:
			ret[item.ParamValue] = item.Threshold
		case KindBool:
			realVal, err := strconv.ParseBool(item.ParamValue)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for bool specific item", "itemValKind", item.ParamKind, "itemValStr", item.ParamValue, "itemThreshold", item.Threshold)
				continue
			}
			ret[realVal] = item.Threshold
		case KindFloat32:
			realVal, err := strconv.ParseFloat(item.ParamValue, 32)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for float specific item", "itemValKind", item.ParamKind, "itemValStr", item.ParamValue, "itemThreshold", item.Threshold)
				continue
			}
			ret[realVal] = item.Threshold
		case KindFloat64:
			realVal, err := strconv.ParseFloat(item.ParamValue, 64)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for float specific item", item.ParamKind, "itemValKind", item.ParamKind, "itemValStr", item.ParamValue, "itemThreshold", item.Threshold)
				continue
			}
			ret[realVal] = item.Threshold
		default:
			logging.Error(errors.New("Unsupported kind for specific item"), "", item.ParamKind, "itemValKind", item.ParamKind, "itemValStr", item.ParamValue, "itemThreshold", item.Threshold)
		}
	}
	return ret
}

// transToSpecificItems trans to the ParamFlowItem as real value.
func transToSpecificItems(source map[interface{}]int64) []ParamFlowItem {
	var ret = make([]ParamFlowItem, 0)
	if len(source) == 0 {
		return ret
	}
	for key, value := range source {
		switch key.(type) {
		case int:
			param := ParamFlowItem{ParamKind: KindInt, ParamValue: key.(string), Threshold: value}
			ret = append(ret, param)
		case int32:
			param := ParamFlowItem{ParamKind: KindInt32, ParamValue: key.(string), Threshold: value}
			ret = append(ret, param)
		case int64:
			param := ParamFlowItem{ParamKind: KindInt64, ParamValue: key.(string), Threshold: value}
			ret = append(ret, param)
		case string:
			param := ParamFlowItem{ParamKind: KindString, ParamValue: key.(string), Threshold: value}
			ret = append(ret, param)
		case float32:
			param := ParamFlowItem{ParamKind: KindFloat32, ParamValue: key.(string), Threshold: value}
			ret = append(ret, param)
		case float64:
			param := ParamFlowItem{ParamKind: KindFloat64, ParamValue: key.(string), Threshold: value}
			ret = append(ret, param)
		case byte:
			param := ParamFlowItem{ParamKind: KindByte, ParamValue: key.(string), Threshold: value}
			ret = append(ret, param)
		case bool:
			param := ParamFlowItem{ParamKind: KindBool, ParamValue: key.(string), Threshold: value}
			ret = append(ret, param)
		default:
			logging.Error(errors.New("Unsupported kind for specific item"), "", key)
		}
	}
	return ret
}
