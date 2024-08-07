package base

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const tokPartSeparator = "|"

var (
	BlockedResult               = &TokResult{Status: int(TokResultStatusBlocked), WaitInMs: 0}
	BadResult                   = &TokResult{Status: int(TokResultStatusBadRequest), WaitInMs: 0}
	NoRuleExistsResult          = &TokResult{Status: int(TokResultStatusNoRuleExists), WaitInMs: 0}
	FailResult                  = &TokResult{Status: int(TokResultStatusFail), WaitInMs: 0}
	TooManyRequestResult        = &TokResult{Status: int(TokResultStatusTooManyRequest), WaitInMs: 0}
	TokenResultStatusFailResult = &TokResult{Status: int(TokResultStatusTooManyRequest), WaitInMs: 0}
	StatusOkResult              = &TokResult{Status: int(TokResultStatusOk), WaitInMs: 0}
)

// TokResult 集群模式下的流控实体
type TokResult struct {
	// Status 状态
	Status int `json:"status"`

	// WaitInMs 休眠时间
	WaitInMs int `json:"waitInMs"`

	// 全局TokenId
	TokenId int64 `json:"tokenId"`

	// Attachments 附加参数
	Attachments map[string]string `json:"attachments"`
}

func (t *TokResult) String() string {
	return "{" +
		"\"status\":" + strconv.Itoa(t.Status) +
		",\"waitInMs\":" + strconv.Itoa(t.WaitInMs) +
		",\"tokenId\":" + strconv.FormatInt(t.TokenId, 10) +
		"}"
}
func (t *TokResult) TokenToThinString() string {
	b := strings.Builder{}
	// All "|" in the resource name will be replaced with "_"
	_, _ = fmt.Fprintf(&b, "%d|%d",
		t.Status, t.WaitInMs)
	return b.String()
}

func ThinStringToToken(line string) (*TokResult, error) {
	if len(line) == 0 {
		return nil, errors.New("invalid token line: empty string")
	}
	token := &TokResult{}
	arr := strings.Split(line, tokPartSeparator)
	if len(arr) < 2 {
		return nil, errors.New("invalid token line: invalid format")
	}
	status, err := strconv.ParseInt(arr[0], 10, 32)
	if err != nil {
		return nil, err
	}
	token.Status = int(status)
	waitInMs, err := strconv.ParseInt(arr[1], 10, 32)
	if err != nil {
		return nil, err
	}
	token.WaitInMs = int(waitInMs)
	return token, nil
}
