package flow

import (
	"github.com/pkg/errors"

	"github.com/liuhailove/gmiter/core/base"
	"github.com/liuhailove/gmiter/logging"
)

const (
	StatSlotOrder = 3000
)

var (
	DefaultStandaloneStatSlot = &StandaloneStatSlot{}
)

type StandaloneStatSlot struct {
}

func (s *StandaloneStatSlot) Order() uint32 {
	return StatSlotOrder
}

// Initial
//
// 初始化，如果有初始化工作放入其中
func (s *StandaloneStatSlot) Initial() {}

func (s *StandaloneStatSlot) OnEntryPassed(ctx *base.EntryContext) {
	res := ctx.Resource.Name()
	for _, tc := range getTrafficControllerListFor(res) {
		if !tc.boundStat.reuseResourceStat {
			if tc.boundStat.writeOnlyMetric != nil {
				// TODO
				tc.boundStat.writeOnlyMetric.AddCount(base.MetricEventPass, int64(ctx.Input.BatchCount))
			} else {
				logging.Error(errors.New("nil independent write statistic"), "Nil statistic for traffic control in StandaloneStatSlot.OnEntryPassed()", "rule", tc.rule)
			}
		}
	}
}

func (s *StandaloneStatSlot) OnEntryBlocked(ctx *base.EntryContext, blockError *base.BlockError) {
	// Do nothing
}

func (s *StandaloneStatSlot) OnCompleted(ctx *base.EntryContext) {
	// Do nothing
}
