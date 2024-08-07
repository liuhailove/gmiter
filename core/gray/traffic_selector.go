package gray

import (
	"github.com/liuhailove/gmiter/core/base"
	"github.com/liuhailove/gmiter/logging"
	"strings"
)

// TrafficSelector 基于灰度规则和实际的流量计算对应的输出资源及地址
type TrafficSelector interface {
	// BoundOwner 归属的流控选择器
	BoundOwner() *TrafficSelectorController

	// CalculateAllowedResource 计算被允许执行的资源接口和地址列表
	CalculateAllowedResource(ctx *base.EntryContext) (reource string, effectiveAddresses string)
}

// TrafficSelectorController 流量选择Controller
type TrafficSelectorController struct {
	// flowCalculator 灰度流量选择器
	flowCalculator TrafficSelector

	// rule 关联规则
	rule *Rule
}

func NewTrafficSelectorController(rule *Rule) (*TrafficSelectorController, error) {
	return &TrafficSelectorController{rule: rule}, nil
}

func (t *TrafficSelectorController) BoundRule() *Rule {
	return t.rule
}

func (t *TrafficSelectorController) FlowSelector() TrafficSelector {
	return t.flowCalculator
}

func (t *TrafficSelectorController) PerformSelecting(ctx *base.EntryContext) *base.TokenResult {
	allowedResource, effectiveAddresses := t.flowCalculator.CalculateAllowedResource(ctx)
	if allowedResource == "" {
		if logging.InfoEnabled() {
			logging.Info("resource no match rule", "resource", ctx.Resource.Name(), "rule", t.rule)
		}
		return nil
	}
	if allowedResource[len(allowedResource)-1] == '*' {
		allowedResource = ctx.Resource.Name()
	}
	newResource := base.NewResourceWrapper(allowedResource, ctx.Resource.Classification(), ctx.Resource.FlowType())
	var grayAddress []string = nil
	if strings.TrimSpace(t.rule.WhiteIpAddresses) != "" {
		grayAddress = strings.Split(t.rule.WhiteIpAddresses, ",")
	}
	if strings.TrimSpace(effectiveAddresses) != "" {
		effectiveAddressArr := strings.Split(effectiveAddresses, ",")
		var grayAddressIntersection = make([]string, 0)
		// 取交集
		if grayAddress != nil {
			for _, address := range grayAddress {
				for _, effectiveAddress := range effectiveAddressArr {
					if address == effectiveAddress {
						grayAddressIntersection = append(grayAddressIntersection, address)
					}
				}
			}
		} else {
			grayAddressIntersection = effectiveAddressArr[:]
		}
		grayAddress = grayAddressIntersection
	}
	return base.NewTokenResultPassWithGrayResource(newResource, t.rule.LinkPass, t.rule.GrayTag, t.rule.Resource, grayAddress)
}
