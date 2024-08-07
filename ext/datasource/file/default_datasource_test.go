package file

import (
	"git.garena.com/honggang.liu/seamiter-go/ext/datasource/util"
	"testing"
)

func TestName(t *testing.T) {
	Initialize()
	util.GetSystemSource().Write([]byte(TestSystemRules))
}
