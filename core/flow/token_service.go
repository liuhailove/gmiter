package flow

//// TokenService 流控服务接口
//type TokenService interface {
//
//	// RequestToken 从远程TokenServer请求tokens
//	//
//	// @param ruleId 唯一规则ID
//	// @param acquireCount 请求token的数量
//	// @param prioritized 请求是否需要优先处理
//	// @return token请求处理结果
//	//
//	RequestToken(tokenServerIp string, tokenServerPort int32, ruleId string, acquireCount uint32, prioritized int32) *base.TokenResult
//
//	// RequestParamToken 从远程TokenServer为一个具体的参数请求Token
//	//
//	// @param ruleId 全局唯一规则ID
//	// @param acquireCount 请求token的数量
//	// @param params 参数列表
//	// @return token请求处理结果
//	//
//	RequestParamToken(tokenServerIp string, tokenServerPort int32, ruleId string, acquireCount uint32, params []interface{}) *base.TokenResult
//
//	// RequestConcurrentToken 从远程TokenServer获取的并发token数
//	//
//	// @param clientAddress 请求方的地址.
//	// @param ruleId 唯一规则ID
//	// @param acquireCount 并发获取的token数
//	// @return token请求处理结果
//	//
//	RequestConcurrentToken(tokenServerIp string, tokenServerPort int32, ruleId string, acquireCount uint32) *base.TokenResult
//
//	// ReleaseConcurrentToken 远程TokenServer异步释放Token数
//	//
//	// @param tokenServerIp tokenServer的IP
//	// @param tokenServerIp tokenServer的Port
//	// @param tokenId 全局唯一tokenId
//	ReleaseConcurrentToken(tokenServerIp string, tokenServerPort int32, tokenId string)
//}
