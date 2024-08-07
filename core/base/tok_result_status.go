package base

type TokResultStatus int

const (
	// TokResultStatusBadRequest 不合法的Client请求
	TokResultStatusBadRequest TokResultStatus = -4
	// TokResultStatusTooManyRequest 在server端有太多的请求
	TokResultStatusTooManyRequest TokResultStatus = -2
	// TokResultStatusFail 由于网络传输或者序列化失败导致的Server或者client非预期错误
	TokResultStatusFail TokResultStatus = -1

	// TokResultStatusOk 获得Token
	TokResultStatusOk TokResultStatus = 0
	// TokResultStatusBlocked 获取Token失败（blocked）
	TokResultStatusBlocked TokResultStatus = 1
	// TokResultStatusShouldWait 应该等待下一个桶
	TokResultStatusShouldWait TokResultStatus = 2
	// TokResultStatusNoRuleExists Token获取失败（规则不存在）
	TokResultStatusNoRuleExists TokResultStatus = 3
	// TokResultStatusNoRefRuleExists 获取Token失败（引用的资源不可用）
	TokResultStatusNoRefRuleExists TokResultStatus = 4
	// TokResultStatusNotAvailable Token获取失败（策略不可用）
	TokResultStatusNotAvailable TokResultStatus = 5
	// TokResultStatusReleaseOk Token被成功释放
	TokResultStatusReleaseOk TokResultStatus = 6
	// TokResultStatusAlreadyRelease Token在请求到达前被释放
	TokResultStatusAlreadyRelease TokResultStatus = 7
)
