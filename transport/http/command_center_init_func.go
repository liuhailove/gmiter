package http

import (
	"errors"
	"github.com/liuhailove/gmiter/constants"
	"github.com/liuhailove/gmiter/core/config"
	"github.com/liuhailove/gmiter/logging"
	_ "github.com/liuhailove/gmiter/transport/common/command/handler" // 强制初始化
	"github.com/liuhailove/gmiter/transport/http/command"
	"github.com/liuhailove/gmiter/util"
)

var (
	commandCenterInitFuncInst = new(commandCenterInitFunc)
)

type commandCenterInitFunc struct {
	isInitialized util.AtomicBool
}

func (c commandCenterInitFunc) Initial() error {
	if !c.isInitialized.CompareAndSet(false, true) {
		return nil
	}
	if config.CloseAll() {
		logging.Warn("[fetchRuleInitFunc] WARN: Sdk closeAll is set true")
		return nil
	}
	var commandCenter = command.GetCommandCenter()
	if commandCenter == nil {
		logging.Warn("[CommandCenterInitFunc] Cannot resolve CommandCenter")
		return errors.New("[CommandCenterInitFunc] Cannot resolve CommandCenter")
	}
	err := commandCenter.BeforeStart()
	if err != nil {
		return err
	}
	err = commandCenter.Start()
	if err != nil {
		return err
	}
	return nil
}

func (c commandCenterInitFunc) Order() int {
	return -1
}

func (c commandCenterInitFunc) ImmediatelyLoadOnce() error {
	return nil
}

// GetRegisterType 获取注册类型
func (c commandCenterInitFunc) GetRegisterType() constants.RegisterType {
	return constants.CommandCenterType
}

func GetCommandCenterInitFuncInst() *commandCenterInitFunc {
	return commandCenterInitFuncInst
}
