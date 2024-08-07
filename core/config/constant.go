package config

const (
	// UnknownProjectName represents the "default" value
	// that indicates the project name is absent.
	UnknownProjectName = "unknown_go_service"

	ConfFilePathEnvKey = "SEA_CONFIG_FILE_PATH"
	AppNameEnvKey      = "SEA_APP_NAME"
	AppTypeEnvKey      = "SEA_APP_TYPE"
	LogDirEnvKey       = "SEA_LOG_DIR"
	LogNamePidEnvKey   = "SEA_LOG_USE_PID"

	DefaultConfigFilename       = "sea.yaml"
	DefaultAppType        int32 = 0

	DefaultMetricLogFlushIntervalSec   uint32 = 1
	DefaultMetricLogSingleFileMaxSize  uint64 = 1024 * 1024 * 50
	DefaultMetricLogMaxFileAmount      uint32 = 8
	DefaultSystemStatCollectIntervalMs uint32 = 1000
	DefaultLoadStatCollectIntervalMs   uint32 = 1000
	DefaultCpuStatCollectIntervalMs    uint32 = 1000
	DefaultMemoryStatCollectIntervalMs uint32 = 150
	DefaultWarmUpColdFactor            uint32 = 3

	DefaultDashServer          = "127.0.0.1:8080"
	DefaultHeartbeatPort       = 8089
	DefaultHeartbeatClintIp    = "127.0.0.1"
	DefaultHeartbeatPath       = "/registry/machine"
	DefaultHeartbeatRemovePath = "/registry/removeMachine"
	DefaultHeartbeatIntervalMs = 10000

	DefaultFetchRuleIntervalMs = 500

	// DefaultFindMaxVersionApiPath 默认获取系统规则相关接口
	DefaultFindMaxVersionApiPath        = "/api/findMaxVersion"
	DefaultQueryAllDegradeRuleApiPath   = "/api/queryAllDegradeRule"
	DefaultQueryAllFlowRuleApiPath      = "/api/queryAllFlowRule"
	DefaultQueryAllParamFlowRuleApiPath = "/api/queryAllParamFlowRule"
	DefaultQueryAllMockRuleApiPath      = "/api/queryAllMockRule"
	DefaultQueryAllSystemRuleApiPath    = "/api/queryAllSystemRule"
	DefaultQueryAllAuthorityApiPath     = "/api/queryAllAuthorityRule"
	DefaultQueryAllRetryApiPath         = "/api/queryAllRetryRule"
	DefaultQueryAllGrayApiPath          = "/api/queryAllGrayRule"
	DefaultQueryAllIsolationApiPath     = "/api/queryAllIsolationRule"
	DefaultQueryAllWeightRouterApiPath  = "/api/queryAllWeightRouterRule"

	DefaultSendIntervalMs     = 2000
	DefaultSendMetricsApiPath = "/api/receiveMetrics"

	DefaultSendRspIntervalMs     = 2000
	DefaultSendRequestIntervalMs = 2000
	DefaultSendRspApiPath        = "/api/receiveRsp"
	DefaultSendRequestApiPath    = "/api/receiveRequest"

	DefaultSourceFilePath       = "./rules"
	DefaultFlowRuleName         = "flowRule.json"
	DefaultAuthorityRuleName    = "authorityRule.json"
	DefaultDegradeRuleName      = "degradeRule.json"
	DefaultSystemRuleName       = "systemRule.json"
	DefaultHotspotRuleName      = "hotspotRule.json"
	DefaultMockRuleName         = "mockRule.json"
	DefaultRetryRuleName        = "retryRule.json"
	DefaultGrayRuleName         = "grayRule.json"
	DefaultIsolationRuleName    = "isolationRule.json"
	DefaultWeightRouterRuleName = "weightRouterRule.json"
	// DefaultLogLevel 默认日志级别，info
	DefaultLogLevel = 1

	DefaultExporterHttpAddr = ":19999"
	DefaultExporterHttpPath = "/sea_metrics"

	// DefaultNameSpace 默认命名空间
	DefaultNameSpace = "seaNameSpace"

	// ClientDefaultNameSpace 客户端默认命名空间
	ClientDefaultNameSpace = "seaClientNameSpace"

	// DefaultMaxAllowQps 默认最大Qps，10万
	DefaultMaxAllowQps = 100000.0

	// DefaultClientMaxAllowQps 单个客户端默认请求的最大QPS为1万
	DefaultClientMaxAllowQps = 10000.0

	// DefaultEtcdV3Prefix 默认的EtcdV3前缀
	DefaultEtcdV3Prefix = "seamiter_etcd_v3"
)
