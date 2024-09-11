package microv4_opentrace

import (
	"errors"
	"github.com/liuhailove/gmiter/core/weight_router"
	"github.com/liuhailove/gmiter/logging"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/selector"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Random is a random strategy algorithm for node selection
func Random(services []*registry.Service) selector.Next {
	nodes := make([]*registry.Node, 0, len(services))

	for _, service := range services {
		nodes = append(nodes, service.Nodes...)
	}

	return func() (*registry.Node, error) {
		if len(nodes) == 0 {
			return nil, selector.ErrNoneAvailable
		}

		i := rand.Int() % len(nodes)
		return nodes[i], nil
	}
}

// WeightSelect 带权重策略的节点选择算法
func WeightSelect(services []*registry.Service) selector.Next {
	nodes := make([]*registry.Node, 0, len(services))
	for _, service := range services {
		nodes = append(nodes, service.Nodes...)
	}

	// weightNodes 用于gmiter节点权重计算
	weightNodes := make([]*weight_router.WeightNode, 0, len(services))
	for _, node := range nodes {
		weightNode := &weight_router.WeightNode{Address: node.Address}
		weightNodes = append(weightNodes, weightNode)
	}

	validServiceName := ""
	if ok, serviceName := checkAndGetServiceName(services); ok {
		validServiceName = serviceName
	} else {
		//理论上不会出现服务名不一致的情况，这里做了兜底处理,若出现非预期情况则回滚到随机的策略
		logging.Error(errors.New("serviceName invalid"), "WeightSelect fail, rollback to random strategy", "error", "serviceName invalid")
		return Random(services)
	}

	return func() (*registry.Node, error) {
		if len(nodes) == 0 {
			return nil, selector.ErrNoneAvailable
		}

		//传入下游服务名和节点列表 经过权重规则计算后返回目标节点序号
		index, err := weight_router.GetTargetNodeByWeightRule(validServiceName, weightNodes)
		if err != nil {
			//遇到非预期错误 回滚到随机算法模式
			logging.Error(err, "WeightSelect GetTargetNodeByWeightRule fail, rollback to random strategy", "error", err.Error())
			i := rand.Int() % len(nodes)
			return nodes[i], nil
		}
		if index < 0 || index >= len(nodes) {
			//index非法 回滚到随机算法模式
			logging.Error(err, "index invalid, rollback to random strategy", "index", index)
			i := rand.Int() % len(nodes)
			return nodes[i], nil
		}

		return nodes[index], nil
	}
}

// 校验所有services的服务名是否一致 并返回一致的服务名
func checkAndGetServiceName(services []*registry.Service) (isValid bool, serviceName string) {
	if len(services) < 1 {
		return false, ""
	}
	serviceName = services[0].Name
	for _, service := range services {
		if serviceName != service.Name {
			return false, ""
		}
	}
	return true, serviceName
}

func GenStrategyWithRouterRules(rules []weight_router.Rule) selector.Strategy {
	return func(services []*registry.Service) selector.Next {
		nodes := make([]*registry.Node, 0, len(services))
		for _, service := range services {
			nodes = append(nodes, service.Nodes...)
		}

		// weightNodes 用于gmiter节点权重计算
		weightNodes := make([]*weight_router.WeightNode, 0, len(services))
		for _, node := range nodes {
			weightNode := &weight_router.WeightNode{Address: node.Address}
			weightNodes = append(weightNodes, weightNode)
		}

		validServiceName := ""
		if ok, serviceName := checkAndGetServiceName(services); ok {
			validServiceName = serviceName
		} else {
			//理论上不会出现服务名不一致的情况，这里做了兜底处理,若出现非预期情况则回滚到随机的策略
			logging.Error(errors.New("serviceName invalid"), "WeightSelect fail, rollback to random strategy", "error", "serviceName invalid")
			return Random(services)
		}

		return func() (*registry.Node, error) {
			if len(nodes) == 0 {
				return nil, selector.ErrNoneAvailable
			}

			//传入下游服务名和节点列表和权重规则 经过计算后返回目标节点序号
			index, err := weight_router.GetTargetIndexByWeightRule(validServiceName, weightNodes, rules)
			if err != nil {
				//遇到非预期错误 回滚到随机算法模式
				logging.Error(err, "WeightSelect GetTargetNodeByWeightRule fail, rollback to random strategy", "error", err.Error())
				i := rand.Int() % len(nodes)
				return nodes[i], nil
			}
			if index < 0 || index >= len(nodes) {
				//index非法 回滚到随机算法模式
				logging.Error(err, "index invalid, rollback to random strategy", "index", index)
				i := rand.Int() % len(nodes)
				return nodes[i], nil
			}

			return nodes[index], nil
		}
	}
}
