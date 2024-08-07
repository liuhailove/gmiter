package mock

import (
	"github.com/liuhailove/gmiter/core/base"
	"github.com/liuhailove/gmiter/util"
	"time"
)

type waitingTrafficShapingController struct {
	baseTrafficShapingController
}

func (w *waitingTrafficShapingController) PerformCheckingFunc(ctx *base.EntryContext) *base.TokenResult {
	if nanosToWait := w.r.ThenReturnWaitingTimeMs * time.Millisecond.Nanoseconds(); nanosToWait > 0 {
		// Handle waiting action.
		util.Sleep(time.Duration(nanosToWait))
	}
	return nil
}

func (w *waitingTrafficShapingController) PerformCheckingArgs(ctx *base.EntryContext) *base.TokenResult {
	return w.DoInnerCheck(ctx)
}
