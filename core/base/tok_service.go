package base

// TokenService 流控服务接口
type TokenService interface {

	// RequestToken 从远程TokenServer请求tokens
	//
	// @param rule 规则信息
	// @param acquireCount 请求token的数量
	// @param prioritized 请求是否需要优先处理
	// @return token请求处理结果
	//
	RequestToken(rule string, acquireCount uint32, prioritized int32) *TokResult

	// RequestParamToken 从远程TokenServer为一个具体的参数请求Token
	//
	// @param rule 规则信息
	// @param acquireCount 请求token的数量
	// @param params 参数列表
	// @return token请求处理结果
	//
	RequestParamToken(rule string, acquireCount uint32, params []interface{}) *TokResult

	// RequestConcurrentToken 从远程TokenServer获取的并发token数
	//
	// @param rule 规则信息
	// @param acquireCount 并发获取的token数
	// @return token请求处理结果
	//
	RequestConcurrentToken(rule string, acquireCount uint32) *TokResult

	// ReleaseConcurrentToken 远程TokenServer异步释放Token数
	//
	// @param rule 规则信息
	// @param tokenId 全局唯一tokenId
	ReleaseConcurrentToken(rule string, tokenId string)
}
