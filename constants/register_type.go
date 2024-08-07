package constants

// RegisterType 注册InitFunc的类型，用于启动初始化
type RegisterType string

const (
	CommandCenterType         RegisterType = "commandCenterType"
	HeartBeatSenderType       RegisterType = "heartBeatSender"
	FetchRuleType             RegisterType = "fetchRuleType"
	SendRspType               RegisterType = "sendRspType"
	SendRequestType           RegisterType = "sendRequestType"
	SendMetricType            RegisterType = "sendMetricType"
	PersistenceDatasourceType RegisterType = "persistenceDatasourceType"
	RedisTokenServiceType     RegisterType = "redisTokenServiceType"
	DefaultTokenServiceType   RegisterType = "defaultTokenServiceType"
)
