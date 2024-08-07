package weight_router

import (
	"fmt"
	"github.com/liuhailove/gmiter/logging"
	"github.com/liuhailove/gmiter/util"
	"github.com/pkg/errors"
	"math"
	"math/rand"
	"reflect"
	"sync"
	"time"
)

var (
	ruleList      = make([]*Rule, 0)
	rwMux         = new(sync.RWMutex)
	updateRuleMux = new(sync.Mutex)
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// LoadRules loads the given weight router rules to the rule manager, while all previous rules will be replaced.
// the first returned value indicates whether you do real load operation, if the rules is the same with previous rules, return false
func LoadRules(rules []*Rule) (bool, error) {
	updateRuleMux.Lock()
	defer updateRuleMux.Unlock()
	isEqual := reflect.DeepEqual(ruleList, rules)
	if isEqual {
		logging.Info("[weight router] Load rules is the same with current rules, so ignore load operation.")
		return false, nil
	}
	err := onRuleUpdate(rules)
	return true, err
}

func onRuleUpdate(rawResRules []*Rule) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%+v", r)
			}
		}
	}()

	// ignore invalid rules
	validResRules := make([]*Rule, 0)
	for _, rule := range rawResRules {
		if err := IsValidRule(rule); err != nil {
			logging.Warn("[weightRouter onRuleUpdate] Ignoring invalid weight router rule", "rule", rule, "reason", err.Error())
			continue
		}
		validResRules = append(validResRules, rule)
	}

	start := util.CurrentTimeNano()
	rwMux.Lock()
	ruleList = validResRules
	rwMux.Unlock()

	if logging.DebugEnabled() {
		logging.Debug("[weightRouter onRuleUpdate] Time statistic(ns) for updating weightRouter rule", "timeCost", util.CurrentTimeNano()-start)
	}
	logRuleUpdate(validResRules)
	return
}

// ClearRules clears all the rules in weight router module.
func ClearRules() error {
	_, err := LoadRules(nil)
	return err
}

// getRules returns all the rules。Any changes of rules take effect for weight router module
// getRules is an internal interface.
func getRules() []Rule {
	rwMux.RLock()
	defer rwMux.RUnlock()

	//给予规则的拷贝 避免对规则的错误修改影响到原有规则
	allRuleList := make([]Rule, 0, len(ruleList))
	for _, rule := range ruleList {
		allRuleList = append(allRuleList, *rule)
	}
	return allRuleList
}

func logRuleUpdate(m []*Rule) {
	if len(m) == 0 {
		logging.Info("[WeightRouterRuleManager] weightRouter rules were cleared")
	} else {
		logging.Info("[WeightRouterRuleManager] weightRouter rules were loaded", "rules", m)
	}
}

// IsValidRule 校验规则数据是否合法.
func IsValidRule(r *Rule) error {
	if r == nil {
		return errors.New("nil weightRouter rule")
	}
	if len(r.TargetAddress) == 0 {
		return errors.New("empty Address of weightRouter rule")
	}
	if r.Weight < 0 {
		return errors.New("invalid weight")
	}
	//当前仅接受客户端权重路由和服务端权重路由
	if r.WeightRuleType != ClientWeightRuleType && r.WeightRuleType != ServerWeightRuleType {
		return errors.New("invalid weight rule type")
	}
	return nil
}

// GetActualRules 获得实际的权重规则列表，该权重规则列表根据优先级判断 过滤掉相同下游的、优先级低的权重规则
func GetActualRules() []Rule {
	allRuleList := getRules()

	//1. 判断当前resource有没有接口级别的路由规则，若有则应用接口级别的路由规则
	//... todo 后续根据需求再实现

	//2. 若没有接口级别的路由规则，则应用节点权重路由规则

	//按照优先级策略 选择目标权重规则
	//如针对同一下游的权重规则，客户端权重优先级>服务端权重优先级，因此需要过滤该服务端权重规则，保留该客户端权重规则
	actualRuleMap := make(map[string]Rule)
	for _, rule := range allRuleList {
		key := rule.ServerServiceName + "-" + rule.TargetAddress
		//如果rule map中已存在该key，则判断对应的规则类型值是否为客户端权重规则，若是则保留原规则，不更新，若不是则新增或覆盖（保持客户端权重优先）
		if actualRule, ok := actualRuleMap[key]; ok {
			if actualRule.WeightRuleType == ClientWeightRuleType {
				continue
			}
		}
		actualRuleMap[key] = rule
	}

	actualRuleList := make([]Rule, 0, len(actualRuleMap))
	for _, rule := range actualRuleMap {
		actualRuleList = append(actualRuleList, rule)
	}

	return actualRuleList
}

func GetTargetNodeByWeightRule(validServiceName string, weightNodes []*WeightNode) (index int, err error) {
	weightRouterRules := GetActualRules()
	weightList := make([]int64, 0, len(weightNodes))
	for _, node := range weightNodes {
		//判断是否有权重规则生效,若生效则使用权重规则的权重，若不生效则使用默认权重
		if ok, weight := isWeightRuleInEffect(node, validServiceName, weightRouterRules); ok {
			weightList = append(weightList, weight)
		} else {
			weightList = append(weightList, DefaultWightForNormalNode)
		}
	}

	//计算权重桶
	weightBuckets := make([]int64, 0, len(weightNodes))
	totalWeight := int64(0)
	for _, weight := range weightList {
		//判断是否溢出 超过最大值 （理论上不会有溢出的场景）
		if weight > math.MaxInt64-totalWeight {
			err = errors.New("total weight overflow")
			logging.Error(err, "total weight overflow ,need to rollback to random strategy")
			return 0, err
		}
		totalWeight += weight
		weightBuckets = append(weightBuckets, totalWeight)
	}

	var hashVal = rand.Int63()
	// 将比例分成weight份，看每次请求落在某份上
	var bucket int64
	if totalWeight > 0 {
		bucket = hashVal%totalWeight + 1
	} else {
		bucket = 1
	}

	for i := 0; i < len(weightBuckets); i++ {
		if bucket <= weightBuckets[i] {
			return i, nil
		}
	}
	//当所有节点权重都为0时，回滚到随机算法模式
	logging.Error(err, "totalWeight=0, rollback to random strategy", "error", "totalWeight=0")
	return 0, errors.New("totalWeight=0, rollback to random strategy")
}

// isWeightRuleInEffect 判断权重规则是否实际生效
func isWeightRuleInEffect(node *WeightNode, serviceName string, rules []Rule) (isValid bool, weight int64) {
	//权重规则生效的条件 服务名称相等&&IP 端口地址相等
	for _, rule := range rules {
		if serviceName == rule.ServerServiceName && node.Address == rule.TargetAddress {
			return true, rule.Weight
		}
	}
	return false, 0
}

func GetTargetIndexByWeightRule(validServiceName string, weightNodes []*WeightNode, weightRouterRules []Rule) (index int, err error) {
	weightList := make([]int64, 0, len(weightNodes))
	for _, node := range weightNodes {
		//判断是否有权重规则生效,若生效则使用权重规则的权重，若不生效则使用默认权重
		if ok, weight := isWeightRuleInEffect(node, validServiceName, weightRouterRules); ok {
			weightList = append(weightList, weight)
		} else {
			weightList = append(weightList, DefaultWightForNormalNode)
		}
	}

	//计算权重桶
	weightBuckets := make([]int64, 0, len(weightNodes))
	totalWeight := int64(0)
	for _, weight := range weightList {
		//判断是否溢出 超过最大值 （理论上不会有溢出的场景）
		if weight > math.MaxInt64-totalWeight {
			err = errors.New("total weight overflow")
			logging.Error(err, "total weight overflow ,need to rollback to random strategy")
			return 0, err
		}
		totalWeight += weight
		weightBuckets = append(weightBuckets, totalWeight)
	}

	var hashVal = rand.Int63()
	// 将比例分成weight份，看每次请求落在某份上
	var bucket int64
	if totalWeight > 0 {
		bucket = hashVal%totalWeight + 1
	} else {
		bucket = 1
	}

	for i := 0; i < len(weightBuckets); i++ {
		if bucket <= weightBuckets[i] {
			return i, nil
		}
	}
	//当所有节点权重都为0时，回滚到随机算法模式
	logging.Error(err, "totalWeight=0, rollback to random strategy", "error", "totalWeight=0")
	return 0, errors.New("totalWeight=0, rollback to random strategy")
}
