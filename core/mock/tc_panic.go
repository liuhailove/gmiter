package mock

import (
	"github.com/liuhailove/gmiter/core/base"
)

type panicTrafficShapingController struct {
	baseTrafficShapingController
}

func (p *panicTrafficShapingController) PerformCheckingFunc(ctx *base.EntryContext) *base.TokenResult {
	return base.NewTokenResultBlockedWithCause(base.BlockTypeMockError, "", p.BoundRule(), p.BoundRule().ThenThrowMsg)
}

func (p *panicTrafficShapingController) PerformCheckingArgs(ctx *base.EntryContext) *base.TokenResult {
	return p.DoInnerCheck(ctx)
}
