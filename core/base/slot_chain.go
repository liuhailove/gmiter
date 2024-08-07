package base

import (
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"github.com/pkg/errors"
	"sort"
	"sync"
)

type BaseSlot interface {
	// Order returns the sort value of the slot.
	// SlotChain will sort all it's slots by ascending sort value in each bucket
	// (StatPrepareSlot bucket、RuleCheckSlot bucket and StatSlot bucket)
	Order() uint32

	// Initial
	//
	// 初始化，如果有初始化工作放入其中
	Initial()
}

// StatPrepareSlot is responsible for some preparation before statistic
// For example: init structure and so on
type StatPrepareSlot interface {
	BaseSlot
	// Prepare function do some initialization
	// Such as: init statistic structure、node and etc
	// The result of preparing would store in EntryContext
	// All StatPrepareSlots execute in sequence
	// Prepare function should not throw panic.
	Prepare(ctx *EntryContext)
}

// RuleCheckSlot is rule based checking strategy
// All checking rule must implement this interface.
type RuleCheckSlot interface {
	BaseSlot

	// Check function do some validation
	// It can break off the slot pipeline
	// Each TokenResult will return check result
	// The upper logic will control pipeline according to SlotResult.
	Check(ctx *EntryContext) *TokenResult
}

// RouterSlot 流量路由策略
// 所有的路由规则都必须实现这个接口.
type RouterSlot interface {
	BaseSlot

	// Router Check function do some validation
	// It can break off the slot pipeline
	// Each TokenResult will return check result
	// The upper logic will control pipeline according to SlotResult.
	Router(ctx *EntryContext) *TokenResult
}

// StatSlot is responsible for counting all custom biz metrics.
// StatSlot would not handle any panic, and pass up all panic to slot chain
type StatSlot interface {
	BaseSlot
	// OnEntryPassed function will be invoked when StatPrepareSlots and RuleCheckSlots execute pass
	// StatSlots will do some statistic logic, such as QPS、log、etc
	OnEntryPassed(ctx *EntryContext)
	// OnEntryBlocked function will be invoked when StatPrepareSlots and RuleCheckSlots fail to execute
	// It may be inbound flow control or outbound cir
	// StatSlots will do some statistic logic, such as QPS、log、etc
	// blockError introduce the block detail
	OnEntryBlocked(ctx *EntryContext, blockError *BlockError)
	// OnCompleted function will be invoked when chain exits.
	// The semantics of OnCompleted is the entry passed and completed
	// Note: blocked entry will not call this function
	OnCompleted(ctx *EntryContext)
}

// SlotChain hold all system slots and customized slot.
// SlotChain support plug-in slots developed by developer.
type SlotChain struct {
	// statPres is in ascending order by StatPrepareSlot.Order() value.
	statPres []StatPrepareSlot
	// ruleChecks is in ascending order by RuleCheckSlot.Order() value.
	ruleChecks []RuleCheckSlot
	// stats is in ascending order by StatSlot.Order() value.
	stats []StatSlot

	// 流量路由策略
	routers []RouterSlot

	// EntryContext Pool, used for reuse EntryContext object
	ctxPool *sync.Pool
}

var (
	ctxPool = &sync.Pool{
		New: func() interface{} {
			ctx := NewEmptyEntryContext()
			ctx.RuleCheckResult = NewTokenResultPass()
			ctx.Data = make(map[interface{}]interface{})
			ctx.Input = &seaInput{
				BatchCount:  1,
				Flag:        0,
				Args:        make([]interface{}, 0),
				Attachments: make(map[interface{}]interface{}),
				Headers:     make(map[string][]string),
				Cookies:     make(map[string][]string),
				Body:        make(map[string][]string),
				MetaData:    make(map[string]string, 0),
			}
			ctx.Output = &seaOutput{Rsps: make([]interface{}, 0)}
			return ctx
		},
	}
)

func NewSlotChain() *SlotChain {
	return &SlotChain{
		statPres:   make([]StatPrepareSlot, 0, 8),
		ruleChecks: make([]RuleCheckSlot, 0, 8),
		stats:      make([]StatSlot, 0, 8),
		ctxPool:    ctxPool,
	}
}

// GetPooledContext 从 EntryContext ctxPool 获取一个 EntryContext，如果 ctxPool 没有足够的 EntryContext 则新建一个.
func (sc *SlotChain) GetPooledContext() *EntryContext {
	ctx := sc.ctxPool.Get().(*EntryContext)
	ctx.startTime = util.CurrentTimeMillis()
	return ctx
}

func (sc *SlotChain) RefurbishContext(c *EntryContext) {
	if c != nil {
		c.Reset()
		sc.ctxPool.Put(c)
	}
}

// AddStatPrepareSlot 将 StatPrepareSlot 插槽添加到 SlotChain 的 StatPrepareSlot 列表中。
// 列表中的所有StatPrepareSlot将根据StatPrepareSlot.Order()升序排序。
// AddStatPrepareSlot 是非线程安全的，
// 并发场景下，AddStatPrepareSlot 必须由 SlotChain.RWMutex#Lock 守护
func (sc *SlotChain) AddStatPrepareSlot(s StatPrepareSlot) {
	sc.statPres = append(sc.statPres, s)
	sort.SliceStable(sc.statPres, func(i, j int) bool {
		return sc.statPres[i].Order() < sc.statPres[j].Order()
	})
}

// AddRuleCheckSlot 将RuleCheckSlot添加到SlotChain的RuleCheckSlot列表中。
// 列表中的所有RuleCheckSlot将根据RuleCheckSlot.Order()升序排序。
// AddRuleCheckSlot 是非线程安全的，
// 并发场景下，AddRuleCheckSlot必须由SlotChain.RWMutex#Lock守护
func (sc *SlotChain) AddRuleCheckSlot(s RuleCheckSlot) {
	// 初始化
	s.Initial()
	sc.ruleChecks = append(sc.ruleChecks, s)
	sort.SliceStable(sc.ruleChecks, func(i, j int) bool {
		return sc.ruleChecks[i].Order() < sc.ruleChecks[j].Order()
	})
}

// AddStatSlot 将 StatSlot 添加到 SlotChain 的 StatSlot 列表中。
// 列表中的所有StatSlot将根据StatSlot.Order()升序排序。
// AddStatSlot 是非线程安全的，
// 并发场景下，AddStatSlot必须由SlotChain.RWMutex#Lock守护
func (sc *SlotChain) AddStatSlot(s StatSlot) {
	sc.stats = append(sc.stats, s)
	sort.SliceStable(sc.stats, func(i, j int) bool {
		return sc.stats[i].Order() < sc.stats[j].Order()
	})
}

// AddRouterSlot 将 RouterSlot 添加到 SlotChain 的 StatSlot 列表中。
// 列表中的所有 RouterSlot 将根据 RouterSlot.Order() 升序排序。
// AdRouterSlot 是非线程安全的，
// 并发场景下，AddStatSlot必须由SlotChain.RWMutex#Lock守护
func (sc *SlotChain) AddRouterSlot(s RouterSlot) {
	sc.routers = append(sc.routers, s)
	sort.SliceStable(sc.routers, func(i, j int) bool {
		return sc.routers[i].Order() < sc.routers[j].Order()
	})
}

// Entry 规则槽职责链入口
// 返回 TokenResult ，如果内部panic返回空
func (sc *SlotChain) Entry(ctx *EntryContext) *TokenResult {
	defer func() {
		// 不应该执行到此块代码，除非SDK内部有错误。
		// 如果发生了，把结果加入到 EntryContext 上下文
		if err := recover(); err != nil {
			logging.Error(errors.Errorf("%+v", err), "sea internal panic in SlotChain.Entry()")
			ctx.SetError(errors.Errorf("%+v", err))
			return
		}
	}()

	// 执行准备的规则槽
	sps := sc.statPres
	if len(sps) > 0 {
		for _, s := range sps {
			s.Prepare(ctx)
		}
	}

	// 执行规则检查
	rcs := sc.ruleChecks
	var ruleCheckRet *TokenResult
	if len(rcs) > 0 {
		for _, s := range rcs {
			sr := s.Check(ctx)
			if sr == nil {
				// 空意味着通过
				continue
			}
			// 检查返回结果
			if sr.IsBlocked() {
				ruleCheckRet = sr
				break
			}
		}
	}
	if ruleCheckRet == nil {
		ctx.RuleCheckResult.ResetToPass()
	} else {
		ctx.RuleCheckResult = ruleCheckRet
	}

	// 路由Check
	rs := sc.routers
	if len(rs) > 0 {
		// 目前只有一个
		for _, s := range rs {
			var routerResult = s.Router(ctx)
			if routerResult == nil {
				continue
			}
			// 检查是否发生阻塞
			if routerResult.IsBlocked() {
				ctx.RuleCheckResult.blockErr = routerResult.blockErr
				ctx.RuleCheckResult.status = ResultStatusBlocked
				break
			}
			ctx.RuleCheckResult.grayRes = routerResult.grayRes
			ctx.RuleCheckResult.grayTag = routerResult.grayTag
			ctx.RuleCheckResult.linkPass = routerResult.linkPass
			ctx.RuleCheckResult.grayAddress = routerResult.grayAddress
			ctx.RuleCheckResult.grayMatchRes = routerResult.grayMatchRes
		}
	}

	// 执行统计槽
	ss := sc.stats
	ruleCheckRet = ctx.RuleCheckResult
	if len(ss) > 0 {
		for _, s := range ss {
			// 指示基于规则的检查槽的结果。
			if !ruleCheckRet.IsBlocked() {
				s.OnEntryPassed(ctx)
			} else {
				// block error不应该为空
				s.OnEntryBlocked(ctx, ruleCheckRet.blockErr)
			}
		}
	}
	return ruleCheckRet
}

func (sc *SlotChain) exit(ctx *EntryContext) {
	if ctx == nil || ctx.Entry() == nil {
		logging.Error(errors.New("entryContext or seantry is nil"),
			"EntryContext or SeaEntry is nil in SlotChain.exit()", "ctx", ctx)
		return
	}
	// 如果被阻塞直接返回，不在调用 OnCompleted
	if ctx.IsBlocked() {
		return
	}
	// OnCompleted 只有在通过时才调用
	for _, s := range sc.stats {
		s.OnCompleted(ctx)
	}
	// 释放上下文
}
