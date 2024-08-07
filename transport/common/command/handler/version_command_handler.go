package handler

import (
	"github.com/liuhailove/gmiter/core/config"
	"github.com/liuhailove/gmiter/transport/common/command"
)

var (
	versionCommandHandlerInst = new(versionCommandHandler)
)

func init() {
	command.RegisterHandler(versionCommandHandlerInst.Name(), versionCommandHandlerInst)
}

type versionCommandHandler struct {
}

func (v versionCommandHandler) Name() string {
	return "version"
}

func (v versionCommandHandler) Desc() string {
	return "get sea version"
}

func (v versionCommandHandler) Handle(request command.Request) *command.Response {
	return command.OfSuccess(config.Version())
}
