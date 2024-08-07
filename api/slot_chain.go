package api

import (
	"github.com/liuhailove/gmiter/core/base"
	"github.com/liuhailove/gmiter/core/circuitbreaker"
	"github.com/liuhailove/gmiter/core/flow"
	"github.com/liuhailove/gmiter/core/gray"
	"github.com/liuhailove/gmiter/core/hotspot"
	"github.com/liuhailove/gmiter/core/isolation"
	"github.com/liuhailove/gmiter/core/log"
	"github.com/liuhailove/gmiter/core/mock"
	"github.com/liuhailove/gmiter/core/stat"
	"github.com/liuhailove/gmiter/core/system"
)

var globalSlotChain = BuildDefaultSlotChain()

func GlobalSlotChain() *base.SlotChain {
	return globalSlotChain
}

func BuildDefaultSlotChain() *base.SlotChain {
	sc := base.NewSlotChain()
	sc.AddStatPrepareSlot(stat.DefaultResourceNodePrepareSlot)

	sc.AddRuleCheckSlot(system.DefaultAdaptiveSlot)
	sc.AddRuleCheckSlot(flow.DefaultSlot)
	sc.AddRuleCheckSlot(isolation.DefaultSlot)
	sc.AddRuleCheckSlot(hotspot.DefaultSlot)
	sc.AddRuleCheckSlot(circuitbreaker.DefaultSlot)
	// 数据Mock Check
	sc.AddRuleCheckSlot(mock.DefaultSlot)

	sc.AddStatSlot(stat.DefaultSlot)
	sc.AddStatSlot(log.DefaultSlot)
	sc.AddStatSlot(flow.DefaultStandaloneStatSlot)
	sc.AddStatSlot(hotspot.DefaultConcurrencyStatSlot)
	sc.AddStatSlot(circuitbreaker.DefaultMetricStatSlot)

	// 增加灰度路由策略
	sc.AddRouterSlot(gray.DefaultSlot)
	return sc
}
