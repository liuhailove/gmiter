package datasource

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/liuhailove/gmiter/core/base"
	cb "github.com/liuhailove/gmiter/core/circuitbreaker"
	"github.com/liuhailove/gmiter/core/flow"
	"github.com/liuhailove/gmiter/core/gray"
	"github.com/liuhailove/gmiter/core/hotspot"
	"github.com/liuhailove/gmiter/core/isolation"
	"github.com/liuhailove/gmiter/core/mock"
	retry "github.com/liuhailove/gmiter/core/retry/rule"
	"github.com/liuhailove/gmiter/core/system"
	"github.com/liuhailove/gmiter/core/weight_router"
)

func checkSrcComplianceJson(src []byte) (bool, error) {
	if len(src) == 0 {
		return false, fmt.Errorf("src data is nil")
	}
	return true, nil
}

// FlowRuleJsonArrayParser provide JSON  as the default serialization for list of flow.Rule
func FlowRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}

	rules := make([]*flow.Rule, 0, 8)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(src, &rules); err != nil {
		desc := fmt.Sprintf("TokenResultStatusFail to convert source bytes to []*flow.Rule, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	}
	return rules, nil
}

// FlowRulesUpdater load the newest []flow.Rule to downstream flow component.
func FlowRulesUpdater(data interface{}) error {
	if data == nil {
		return flow.ClearRules()
	}

	rules := make([]*flow.Rule, 0, 8)
	if val, ok := data.([]flow.Rule); ok {
		for _, v := range val {
			rules = append(rules, &v)
		}
	} else if val, ok := data.([]*flow.Rule); ok {
		rules = val
	} else {
		return NewError(UpdatePropertyError, fmt.Sprintf("TokenResultStatusFail to type assert data to []flow.Rule or []*flow.Rule, in fact, data: %+v", data))
	}
	_, err := flow.LoadRules(rules)
	if err == nil {
		return nil
	}
	return NewError(UpdatePropertyError, fmt.Sprintf("%+v", err))
}

func NewFlowRulesHandler(converter PropertyConverter) PropertyHandler {
	return NewDefaultPropertyHandler(converter, FlowRulesUpdater)
}

// SystemRuleJsonArrayParser provide JSON  as the default serialization for list of system.Rule
func SystemRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}
	rules := make([]*SystemRule, 0, 8)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(src, &rules); err != nil {
		desc := fmt.Sprintf("TokenResultStatusFail to convert source bytes to []*system.Rule, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	}
	// 转换为core中的system.rule
	ruleArr := make([]*system.Rule, 0, 8)
	for _, r := range rules {
		rule := &system.Rule{
			ID: r.ID,
		}
		if r.HighestSystemLoad > 0.0 {
			rule.MetricType = system.Load
			rule.TriggerCount = r.HighestSystemLoad
			ruleArr = append(ruleArr, rule)
		} else if r.HighestCpuUsage > 0.0 {
			rule.MetricType = system.CpuUsage
			rule.TriggerCount = r.HighestCpuUsage
			ruleArr = append(ruleArr, rule)
		} else if r.Qps > 0.0 {
			rule.MetricType = system.InboundQPS
			rule.TriggerCount = r.Qps
			ruleArr = append(ruleArr, rule)
		} else if r.AvgRt > 0.0 {
			rule.MetricType = system.AvgRT
			rule.TriggerCount = r.AvgRt
			ruleArr = append(ruleArr, rule)
		} else if r.MaxThread > 0.0 {
			rule.MetricType = system.Concurrency
			rule.TriggerCount = r.MaxThread
			ruleArr = append(ruleArr, rule)
		}
	}
	return ruleArr, nil
}

// SystemRulesUpdater load the newest []system.Rule to downstream system component.
func SystemRulesUpdater(data interface{}) error {
	if data == nil {
		return system.ClearRules()
	}
	rules := make([]*system.Rule, 0, 8)
	if val, ok := data.([]system.Rule); ok {
		for _, v := range val {
			rules = append(rules, &v)
		}
	} else if val, ok := data.([]*system.Rule); ok {
		rules = val
	} else {
		return NewError(UpdatePropertyError, fmt.Sprintf("TokenResultStatusFail to type assert data to []system.Rule or []*system.Rule, in fact, data: %+v", data))
	}
	_, err := system.LoadRules(rules)
	if err == nil {
		return nil
	}
	return NewError(UpdatePropertyError, fmt.Sprintf("%+v", err))
}

func NewSystemRulesHandler(converter PropertyConverter) *DefaultPropertyHandler {
	return NewDefaultPropertyHandler(converter, SystemRulesUpdater)
}

func CircuitBreakerRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}
	degradeRules := make([]*CircuitBreakerRule, 0, 8)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(src, &degradeRules); err != nil {
		desc := fmt.Sprintf("TokenResultStatusFail to convert source bytes to []*circuitbreaker, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	}
	rules := make([]*cb.Rule, len(degradeRules))
	for idx, dg := range degradeRules {
		rule := new(cb.Rule)
		rule.Id = dg.Id
		rule.LimitApp = dg.LimitApp
		rule.Resource = dg.Resource
		rule.Strategy = dg.Strategy
		rule.RetryTimeoutMs = dg.RetryTimeoutMs
		rule.MinRequestAmount = dg.MinRequestAmount
		rule.StatIntervalMs = dg.StatIntervalMs
		rule.StatSlidingWindowBucketCount = dg.StatSlidingWindowBucketCount
		rule.MaxAllowedRtMs = dg.MaxAllowedRtMs
		rule.Threshold = dg.Threshold
		rule.ProbeNum = dg.ProbeNum
		if dg.Strategy == cb.SlowRequestRatio {
			rule.MaxAllowedRtMs = uint64(dg.Threshold)
			rule.Threshold = dg.SlowRatioThreshold
		}
		rules[idx] = rule
	}
	return rules, nil
}

// CircuitBreakerRulesUpdater load the newest []cb.Rule to downstream circuit breaker component.
func CircuitBreakerRulesUpdater(data interface{}) error {
	if data == nil {
		return cb.ClearRules()
	}
	var rules []*cb.Rule
	if val, ok := data.([]*cb.Rule); ok {
		rules = val
	} else {
		return NewError(UpdatePropertyError, fmt.Sprintf("TokenResultStatusFail to type assert data to []*circuitbreaker.Rule, in fact, data: %+v", data))
	}
	_, err := cb.LoadRules(rules)
	if err == nil {
		return nil
	}
	return NewError(UpdatePropertyError, fmt.Sprintf("%+v", err))
}

// GetCircuitBreakerRules 获取熔断策略
func GetCircuitBreakerRules() []CircuitBreakerRule {
	rules := cb.GetRules()
	return transToCircuitBreakerRule(rules)
}

func NewCircuitBreakerRulesHandler(converter PropertyConverter) *DefaultPropertyHandler {
	return NewDefaultPropertyHandler(converter, CircuitBreakerRulesUpdater)
}

// HotSpotParamRuleJsonArrayParser decodes list of param flow rules from JSON bytes.
func HotSpotParamRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}

	hotspotRules := make([]*HotspotRule, 0, 8)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(src, &hotspotRules); err != nil {
		desc := fmt.Sprintf("TokenResultStatusFail to convert source bytes to []*hotspot.Rule, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	}
	rules := make([]*hotspot.Rule, len(hotspotRules))
	for i, hotspotRule := range hotspotRules {
		rules[i] = &hotspot.Rule{
			ID:                hotspotRule.ID,
			Resource:          hotspotRule.Resource,
			LimitApp:          hotspotRule.LimitApp,
			MetricType:        hotspotRule.MetricType,
			ControlBehavior:   hotspotRule.ControlBehavior,
			ParamIdx:          hotspotRule.ParamIdx,
			ParamSource:       hotspotRule.ParamSource,
			ParamKind:         hotspotRule.ParamKind,
			ParamKey:          hotspotRule.ParamKey,
			Threshold:         hotspotRule.Threshold,
			MaxQueueingTimeMs: hotspotRule.MaxQueueingTimeMs,
			BurstCount:        hotspotRule.BurstCount,
			DurationInSec:     hotspotRule.DurationInSec,
			ParamsMaxCapacity: hotspotRule.ParamsMaxCapacity,
			SpecificItems:     parseSpecificItems(ParamKind(hotspotRule.ParamKind), hotspotRule.ParamFlowItems),
			ClusterMode:       hotspotRule.ClusterMode,
		}
		if hotspotRule.ClusterMode && hotspotRule.ClusterConfig != nil {
			rules[i].ClusterConfig = &hotspot.ClusterConfig{
				FlowId:                  hotspotRule.ClusterConfig.FlowId,
				FallbackToLocalWhenFail: hotspotRule.ClusterConfig.FallbackToLocalWhenFail,
				ClusterStrategy:         hotspotRule.ClusterConfig.ClusterStrategy,
				ResourceTimeout:         hotspotRule.ClusterConfig.ResourceTimeout,
				ResourceTimeoutStrategy: hotspotRule.ClusterConfig.ResourceTimeoutStrategy,
				ClientOfflineTime:       hotspotRule.ClusterConfig.ClientOfflineTime,
				GlobalThreshold:         hotspotRule.ClusterConfig.GlobalThreshold,
				TokenServerStrategy:     hotspotRule.ClusterConfig.TokenServerStrategy,
				TokenServerAddress:      hotspotRule.ClusterConfig.TokenServerAddress,
				DowngradeDurationInMs:   hotspotRule.ClusterConfig.DowngradeDurationInMs,
				MasterNodeThreshold:     hotspotRule.ClusterConfig.MasterNodeThreshold,
				TokenServerMasterHost:   hotspotRule.ClusterConfig.TokenServerMasterHost,
				TokenServerMasterPort:   hotspotRule.ClusterConfig.TokenServerMasterPort,
			}
		}
	}
	return rules, nil
}

// HotSpotParamRuleTrans 对象转换为json字节
func HotSpotParamRuleTrans(rules []hotspot.Rule) ([]byte, error) {
	hotspotRules := make([]*HotspotRule, len(rules))
	for i, rule := range rules {
		hotspotRules[i] = &HotspotRule{
			ID:                rule.ID,
			LimitApp:          rule.LimitApp,
			Resource:          rule.Resource,
			MetricType:        rule.MetricType,
			ControlBehavior:   rule.ControlBehavior,
			ParamIdx:          rule.ParamIdx,
			ParamKey:          rule.ParamKey,
			Threshold:         rule.Threshold,
			MaxQueueingTimeMs: rule.MaxQueueingTimeMs,
			BurstCount:        rule.BurstCount,
			DurationInSec:     rule.DurationInSec,
			ParamsMaxCapacity: rule.ParamsMaxCapacity,
			ParamFlowItems:    transToSpecificItems(rule.SpecificItems),
		}
	}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if data, err := json.Marshal(hotspotRules); err != nil {
		desc := fmt.Sprintf("TokenResultStatusFail to trans rules to bytes, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	} else {
		return data, nil
	}
}

// SystemRuleTrans 对象转换为json字节
func SystemRuleTrans(rules []system.Rule) ([]byte, error) {
	systemRules := transToSystemRule(rules)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if data, err := json.Marshal(systemRules); err != nil {
		desc := fmt.Sprintf("TokenResultStatusFail to trans rules to bytes, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	} else {
		return data, nil
	}
}

// HotSpotParamRulesUpdater loads the provided hot-spot param rules to downstream rule manager.
func HotSpotParamRulesUpdater(data interface{}) error {
	if data == nil {
		return hotspot.ClearRules()
	}

	rules := make([]*hotspot.Rule, 0, 8)
	if val, ok := data.([]hotspot.Rule); ok {
		for _, v := range val {
			rules = append(rules, &v)
		}
	} else if val, ok := data.([]*hotspot.Rule); ok {
		rules = val
	} else {
		return NewError(UpdatePropertyError, fmt.Sprintf("TokenResultStatusFail to type assert data to []hotspot.Rule or []*hotspot.Rule, in fact, data: %+v", data))
	}

	_, err := hotspot.LoadRules(rules)
	if err == nil {
		return nil
	}
	return NewError(UpdatePropertyError, fmt.Sprintf("%+v", err))
}

func NewHotSpotParamRulesHandler(converter PropertyConverter) PropertyHandler {
	return NewDefaultPropertyHandler(converter, HotSpotParamRulesUpdater)
}

func NewMockRulesHandler(converter PropertyConverter) PropertyHandler {
	return NewDefaultPropertyHandler(converter, MockRulesUpdater)
}

func NewRetryRulesHandler(converter PropertyConverter) PropertyHandler {
	return NewDefaultPropertyHandler(converter, RetryRulesUpdater)
}

func NewGrayRulesHandler(converter PropertyConverter) PropertyHandler {
	return NewDefaultPropertyHandler(converter, GrayRulesUpdater)
}

func NewWeightRouterRulesHandler(converter PropertyConverter) PropertyHandler {
	return NewDefaultPropertyHandler(converter, WeightRouterRulesUpdater)
}

// IsolationRuleJsonArrayParser provide JSON  as the default serialization for list of isolation.Rule
func IsolationRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}

	rules := make([]*isolation.Rule, 0, 8)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(src, &rules); err != nil {
		desc := fmt.Sprintf("TokenResultStatusFail to convert source bytes to []*isolation.Rule, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	}
	return rules, nil
}

// IsolationRulesUpdater load the newest []isolation.Rule to downstream system component.
func IsolationRulesUpdater(data interface{}) error {
	if data == nil {
		return isolation.ClearRules()
	}

	rules := make([]*isolation.Rule, 0, 8)
	if val, ok := data.([]isolation.Rule); ok {
		for _, v := range val {
			rules = append(rules, &v)
		}
	} else if val, ok := data.([]*isolation.Rule); ok {
		rules = val
	} else {
		return NewError(
			UpdatePropertyError,
			fmt.Sprintf("TokenResultStatusFail to type assert data to []isolation.Rule or []*isolation.Rule, in fact, data: %+v", data),
		)
	}
	_, err := isolation.LoadRules(rules)
	if err == nil {
		return nil
	}
	return NewError(
		UpdatePropertyError,
		fmt.Sprintf("%+v", err),
	)
}

// WeightRouterRuleJsonArrayParser provide JSON  as the default serialization for list of weightRouter.Rule
func WeightRouterRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}

	rules := make([]*weight_router.Rule, 0, 8)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(src, &rules); err != nil {
		desc := fmt.Sprintf("TokenResultStatusFail to convert source bytes to []*weight_router.Rule, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	}
	return rules, nil
}

// WeightRouterRulesUpdater load the newest []weightRouter.Rule to downstream system component.
func WeightRouterRulesUpdater(data interface{}) error {
	if data == nil {
		return weight_router.ClearRules()
	}

	rules := make([]*weight_router.Rule, 0, 8)
	if val, ok := data.([]weight_router.Rule); ok {
		for _, v := range val {
			rules = append(rules, &v)
		}
	} else if val, ok := data.([]*weight_router.Rule); ok {
		rules = val
	} else {
		return NewError(
			UpdatePropertyError,
			fmt.Sprintf("TokenResultStatusFail to type assert data to []weight_router.Rule or []*weight_router.Rule, in fact, data: %+v", data),
		)
	}
	_, err := weight_router.LoadRules(rules)
	if err == nil {
		return nil
	}
	return NewError(
		UpdatePropertyError,
		fmt.Sprintf("%+v", err),
	)
}

func NewIsolationRulesHandler(converter PropertyConverter) *DefaultPropertyHandler {
	return NewDefaultPropertyHandler(converter, IsolationRulesUpdater)
}

// NodeStatTrans 节点统计转换
func NodeStatTrans(metricItems []*base.MetricItem) ([]byte, error) {
	var metricItemData = transToNode(metricItems)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if data, err := json.Marshal(metricItemData); err != nil {
		desc := fmt.Sprintf("TokenResultStatusFail to trans metric to bytes, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	} else {
		return data, nil
	}
}

// MockRuleJsonArrayParser decodes list of mock rules from JSON bytes.
func MockRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}
	rules := make([]*mock.Rule, 0, 8)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(src, &rules); err != nil {
		desc := fmt.Sprintf("TokenResultStatusFail to convert source bytes to []*mock.Rule, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	}
	return rules, nil
}

// MockRulesUpdater loads the provided mock rules to downstream rule manager.
func MockRulesUpdater(data interface{}) error {
	if data == nil {
		return mock.ClearRules()
	}

	rules := make([]*mock.Rule, 0, 8)
	if val, ok := data.([]mock.Rule); ok {
		for _, v := range val {
			rules = append(rules, &v)
		}
	} else if val, ok := data.([]*mock.Rule); ok {
		rules = val
	} else {
		return NewError(UpdatePropertyError, fmt.Sprintf("TokenResultStatusFail to type assert data to []mock.Rule or []*mock.Rule, in fact, data: %+v", data))
	}

	_, err := mock.LoadRules(rules)
	if err == nil {
		return nil
	}
	return NewError(UpdatePropertyError, fmt.Sprintf("%+v", err))
}

// RetryRuleJsonArrayParser decodes list of retry rules from JSON bytes.
func RetryRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}
	rules := make([]*retry.Rule, 0, 8)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(src, &rules); err != nil {
		desc := fmt.Sprintf("TokenResultStatusFail to convert source bytes to []*retry.Rule, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	}
	return rules, nil
}

// GrayRuleJsonArrayParser 从src JSON 字节解码Gray规则列表
func GrayRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}
	rules := make([]*gray.Rule, 0, 8)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(src, &rules); err != nil {
		desc := fmt.Sprintf("TokenResultStatusFail to convert source bytes to []*gray.Rule, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	}
	return rules, nil
}

// RetryRulesUpdater loads the provided retry rules to downstream rule manager.
func RetryRulesUpdater(data interface{}) error {
	if data == nil {
		return retry.ClearRules()
	}

	rules := make([]*retry.Rule, 0, 8)
	if val, ok := data.([]retry.Rule); ok {
		for _, v := range val {
			rules = append(rules, &v)
		}
	} else if val, ok := data.([]*retry.Rule); ok {
		rules = val
	} else {
		return NewError(UpdatePropertyError, fmt.Sprintf("TokenResultStatusFail to type assert data to []retry.Rule or []*retry.Rule, in fact, data: %+v", data))
	}

	_, err := retry.LoadRules(rules)
	if err == nil {
		return nil
	}
	return NewError(UpdatePropertyError, fmt.Sprintf("%+v", err))
}

// GrayRulesUpdater loads the provided gray rules to downstream rule manager.
func GrayRulesUpdater(data interface{}) error {
	if data == nil {
		return gray.ClearRules()
	}

	rules := make([]*gray.Rule, 0, 8)
	if val, ok := data.([]gray.Rule); ok {
		for _, v := range val {
			rules = append(rules, &v)
		}
	} else if val, ok := data.([]*gray.Rule); ok {
		rules = val
	} else {
		return NewError(UpdatePropertyError, fmt.Sprintf("TokenResultStatusFail to type assert data to []gray.Rule or []*gray.Rule, in fact, data: %+v", data))
	}

	_, err := gray.LoadRules(rules)
	if err == nil {
		return nil
	}
	return NewError(UpdatePropertyError, fmt.Sprintf("%+v", err))
}

// Publish 规则发布
type Publish struct {
	// App 应用名称
	App string `json:"app"`
	// 规则类型
	//（1：流控规则-FLOW_RULE，
	//  2：降级规则-DEGRADE_RULE
	//  3：热点规则-HOT_PARAM_RULE，
	//  4：Mock规则-MOCK_RULE，
	//  5：系统规则-SYSTEM_RULE,
	//  6: 授权规则-AUTHORITY_RULE
	//  7: 接口重试规则-RETRY_RULE
	//  8: 待定)

	RuleType int32  `json:"ruleType"`
	Version  string `json:"version"`
}

// RulePublishJsonArrayParser decodes list of rule publish from JSON bytes.
func RulePublishJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}
	rules := make([]*Publish, 0, 8)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(src, &rules); err != nil {
		desc := fmt.Sprintf("TokenResultStatusFail to convert source bytes to []*mock.Rule, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	}
	return rules, nil
}
