package config

import (
	"github.com/liuhailove/gmiter/logging"
	"github.com/liuhailove/gmiter/util"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

var (
	globalCfg   = NewDefaultConfig()
	initLogOnce sync.Once
)

func ResetGlobalConfig(config *Entity) {
	globalCfg = config
}

func GetGlobalConfig() *Entity {
	return globalCfg
}

// InitConfigWithYaml loads general configuration from the YAML file under provided path.
func InitConfigWithYaml(filePath string) (err error) {
	// Initialize general config and logging module.
	if err = applyYamlConfigFile(filePath); err != nil {
		return err
	}
	return OverrideConfigFromEnvAndInitLog()
}

// applyYamlConfigFile loads general configuration from the given YAML file.
func applyYamlConfigFile(configPath string) error {
	// 优先级：系统环境变量 > YAML 文件 >默认配置
	if util.IsBlank(configPath) {
		// If the config file path is absent, sea will try to resolve it from the system env.
		configPath = os.Getenv(ConfFilePathEnvKey)
	}
	if util.IsBlank(configPath) {
		configPath = DefaultConfigFilename
	}
	// 我们将会尝试从配置文件中加载配置
	// 如果配置文件路径没有设置，则会使用默认配置
	return loadGlobalConfigFromYamlFile(configPath)
}

func OverrideConfigFromEnvAndInitLog() error {
	// 我们将会从环境变量重获取基础的配置项，
	// 如果环境变量中存在和配置文件相同的变量，那么配置文件的变量将会被覆盖
	err := overrideItemsFromSystemEnv()
	if err != nil {
		return err
	}
	defer logging.Info("[Config] Print effective global config", "globalConfig", *globalCfg)
	// Configured Logger is the highest priority
	if configLogger := Logger(); configLogger != nil {
		err = logging.ResetGlobalLogger(configLogger)
		if err != nil {
			return err
		}
		return nil
	}

	logDir := LogBaseDir()
	if len(logDir) == 0 {
		logDir = GetDefaultLogDir()
	}
	if err := initializeLogConfig(logDir, LogUsePid()); err != nil {
		return err
	}

	// 日志级别
	logging.ResetGlobalLoggerLevel(logging.Level(LogLevel()))

	logging.Info("[Config] App name resolved", "appName", AppName())
	return nil
}

func loadGlobalConfigFromYamlFile(filePath string) error {
	if filePath == DefaultConfigFilename {
		if _, err := os.Stat(DefaultConfigFilename); err != nil {
			//use default globalCfg.
			return nil
		}
	}
	_, err := os.Stat(filePath)
	if err != nil && !os.IsExist(err) {
		return err
	}
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(content, globalCfg)
	if err != nil {
		return err
	}
	logging.Info("[Config] Resolving sea config from file", "file", filePath)
	return checkConfValid(&(globalCfg.Conf))
}

func overrideItemsFromSystemEnv() error {
	if appName := os.Getenv(AppNameEnvKey); !util.IsBlank(appName) {
		globalCfg.Conf.App.Name = appName
	}

	if appTypeStr := os.Getenv(AppTypeEnvKey); !util.IsBlank(appTypeStr) {
		appType, err := strconv.ParseInt(appTypeStr, 10, 32)
		if err != nil {
			return err
		}
		globalCfg.Conf.App.Type = int32(appType)
	}

	if addPidStr := os.Getenv(LogNamePidEnvKey); !util.IsBlank(addPidStr) {
		addPid, err := strconv.ParseBool(addPidStr)
		if err != nil {
			return err
		}
		globalCfg.Conf.Log.UsePid = addPid
	}
	if logDir := os.Getenv(LogDirEnvKey); !util.IsBlank(logDir) {
		globalCfg.Conf.Log.Dir = logDir
	}
	return checkConfValid(&(globalCfg.Conf))
}

func initializeLogConfig(logDir string, usePid bool) (err error) {
	if logDir == "" {
		return errors.New("invalid empty log path")
	}
	initLogOnce.Do(func() {
		if err = util.CreateDirIfNotExists(logDir); err != nil {
			return
		}
		err = reconfigureRecordLogger(logDir, usePid)
	})
	return err
}

func reconfigureRecordLogger(logBaseDir string, withPid bool) error {
	filePath := filepath.Join(logBaseDir, logging.RecordLogFileName)
	if withPid {
		filePath = filePath + ".pid" + strconv.Itoa(os.Getppid())
	}
	fileLogger, err := logging.NewSimpleFileLogger(filePath)
	if err != nil {
		return err
	}
	// Note: not thread-safe!
	if err = logging.ResetGlobalLogger(fileLogger); err != nil {
		return err
	}
	logging.Info("[Config] Log base directory", "baseDir", logBaseDir)
	return nil
}

func GetDefaultLogDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, logging.DefaultDirName)
}

func AppName() string {
	return globalCfg.AppName()
}

func AppType() int32 {
	return globalCfg.AppType()
}

func Logger() logging.Logger {
	return globalCfg.Logger()
}

func LogBaseDir() string {
	return globalCfg.LogBaseDir()
}

func LogLevel() uint8 {
	return globalCfg.LogLevel()
}

// LogUsePid returns whether the log file name contains the PID suffix.
func LogUsePid() bool {
	return globalCfg.LogUsePid()
}

func MetricExportHTTPAddr() string {
	return globalCfg.MetricExportHTTPAddr()
}

func MetricExportHTTPPath() string {
	return globalCfg.MetricExportHTTPPath()
}

func MetricLogFlushIntervalSec() uint32 {
	return globalCfg.MetricLogFlushIntervalSec()
}

func MetricLogSingleFileMaxSize() uint64 {
	return globalCfg.MetricLogSingleFileMaxSize()
}

func MetricLogMaxFileAmount() uint32 {
	return globalCfg.MetricLogMaxFileAmount()
}

func SystemStatCollectIntervalMs() uint32 {
	return globalCfg.SystemStatCollectIntervalMs()
}

func LoadStatCollectIntervalMs() uint32 {
	return globalCfg.LoadStatCollectIntervalMs()
}

func CpuStatCollectIntervalMs() uint32 {
	return globalCfg.CpuStatCollectIntervalMs()
}

func MemoryStatCollectIntervalMs() uint32 {
	return globalCfg.MemoryStatCollectIntervalMs()
}

func UseCacheTime() bool {
	return globalCfg.UseCacheTime()
}

func GlobalStatisticIntervalMsTotal() uint32 {
	return globalCfg.GlobalStatisticIntervalMsTotal()
}

func GlobalStatisticSampleCountTotal() uint32 {
	return globalCfg.GlobalStatisticSampleCountTotal()
}

func GlobalStatisticBucketLengthInMs() uint32 {
	return globalCfg.GlobalStatisticIntervalMsTotal() / GlobalStatisticSampleCountTotal()
}

func MetricStatisticIntervalMs() uint32 {
	return globalCfg.MetricStatisticIntervalMs()
}
func MetricStatisticSampleCount() uint32 {
	return globalCfg.MetricStatisticSampleCount()
}

func ConsoleServer() string {
	return globalCfg.Conf.Dashboard.Server
}

func ConsolePort() uint32 {
	return globalCfg.Conf.Dashboard.Port
}

func HeartbeatClintIp() string {
	if AutoHeartbeatClientIp() {
		return util.GetIP()
	}
	return globalCfg.Conf.Dashboard.HeartbeatClientIp
}

func AutoHeartbeatClientIp() bool {
	return globalCfg.Conf.Dashboard.AutoHeartbeatClientIp
}

func HeartbeatApiPath() string {
	return globalCfg.Conf.Dashboard.HeartbeatApiPath
}

func HeartbeatRemoveApiDefaultPath() string {
	return globalCfg.Conf.Dashboard.HeartbeatRemoveApiPath
}

func HeartBeatIntervalMs() uint64 {
	return globalCfg.Conf.Dashboard.HeartBeatIntervalMs
}

func HeartBeatDynamicRouterFlag() string {
	return globalCfg.Conf.Dashboard.HeartBeatDynamicRouterFlag
}

func FetchRuleIntervalMs() uint64 {
	return globalCfg.Conf.Dashboard.FetchRuleIntervalMs
}

func Version() string {
	return globalCfg.Version
}

func FindMaxVersionApiPath() string {
	return globalCfg.Conf.Dashboard.FindMaxVersionApiPath
}

func QueryAllDegradeRuleApiPath() string {
	return globalCfg.Conf.Dashboard.QueryAllDegradeRuleApiPath
}

func QueryAllFlowRuleApiPath() string {
	return globalCfg.Conf.Dashboard.QueryAllFlowRuleApiPath
}

func QueryAllParamFlowRuleApiPath() string {
	return globalCfg.Conf.Dashboard.QueryAllParamFlowRuleApiPath
}

func QueryAllMockRuleApiPath() string {
	return globalCfg.Conf.Dashboard.QueryAllMockRuleApiPath
}

func QueryAllSystemRuleApiPath() string {
	return globalCfg.Conf.Dashboard.QueryAllSystemRuleApiPath
}

func QueryAllAuthorityRuleApiPath() string {
	return globalCfg.Conf.Dashboard.QueryAllAuthorityRuleApiPath
}

func QueryAllRetryRuleApiPath() string {
	return globalCfg.Conf.Dashboard.QueryAllRetryRuleApiPath
}

func QueryAllGrayRuleApiPath() string {
	return globalCfg.Conf.Dashboard.QueryAllGrayRuleApiPath
}

func QueryAllIsolationRuleApiPath() string {
	return globalCfg.Conf.Dashboard.QueryAllIsolationRuleApiPath
}

func QueryAllWeightRouterRuleApiPath() string {
	return globalCfg.Conf.Dashboard.QueryAllWeightRouterRuleApiPath
}
func SendMetricIntervalMs() uint64 {
	return globalCfg.Conf.Dashboard.SendMetricIntervalMs
}

func SendMetricApiPath() string {
	return globalCfg.Conf.Dashboard.SendMetricApiPath
}

func SendRspApiPathIntervalMs() uint64 {
	return globalCfg.Conf.Dashboard.SendRspApiPathIntervalMs
}

func SendRequestApiPathIntervalMs() uint64 {
	return globalCfg.Conf.Dashboard.SendRequestApiPathIntervalMs
}

func SendRspApiPath() string {
	return globalCfg.Conf.Dashboard.SendRspApiPath
}

func SendRequestApiPath() string {
	return globalCfg.Conf.Dashboard.SendRequestApiPath
}

func ProxyUrl() string {
	return globalCfg.Conf.Dashboard.ProxyUrl
}

func OpenConnectDashboard() bool {
	return globalCfg.Conf.Dashboard.OpenConnectDashboard
}

func CloseAll() bool {
	return globalCfg.Conf.CloseAll
}

func RuleConsistentModeType() RulePersistencetMode {
	return globalCfg.Conf.RulePersistentMode
}

func SourceFilePath() string {
	return globalCfg.Conf.FileDatasourceConfig.SourceFilePath
}

func FlowRuleName() string {
	return globalCfg.Conf.FileDatasourceConfig.FlowRuleName
}

func AuthorityRuleName() string {
	return globalCfg.Conf.FileDatasourceConfig.AuthorityRuleName
}

func DegradeRuleName() string {
	return globalCfg.Conf.FileDatasourceConfig.DegradeRuleName
}

func SystemRuleName() string {
	return globalCfg.Conf.FileDatasourceConfig.SystemRuleName
}

func HotspotRuleName() string {
	return globalCfg.Conf.FileDatasourceConfig.HotspotRuleName
}

func MockRuleName() string {
	return globalCfg.Conf.FileDatasourceConfig.MockRuleName
}

func RetryRuleName() string {
	return globalCfg.Conf.FileDatasourceConfig.RetryRuleName
}

func GrayRuleName() string {
	return globalCfg.Conf.FileDatasourceConfig.GrayRuleName
}

func IsolationRuleName() string {
	return globalCfg.Conf.FileDatasourceConfig.IsolationRuleName
}

func WeightRouterRuleName() string {
	return globalCfg.Conf.FileDatasourceConfig.WeightRouterRuleName
}
func ImmediatelyFetch() bool {
	return globalCfg.Conf.Dashboard.ImmediatelyFetch
}

func Namespace() string {
	return globalCfg.Conf.ClusterConfig.Namespace
}

func ClientNamespace() string {
	return globalCfg.Conf.ClusterConfig.ClientNamespace
}

func MaxAllowQps() float64 {
	return globalCfg.Conf.ClusterConfig.MaxAllowQps
}

func ClientMaxAllowQps() float64 {
	return globalCfg.Conf.ClusterConfig.ClientMaxAllowQps
}

func EtcdDatasourceEndpoints() string {
	return globalCfg.Conf.EtcdV3DatasourceConfig.Endpoints
}

func EtcdDatasourceKeyPrefix() string {
	return globalCfg.Conf.EtcdV3DatasourceConfig.KeyPrefix
}

func ApolloDatasourceAppId() string {
	return globalCfg.Conf.ApolloDatasourceConfig.AppID
}

func ApolloDatasourceCluster() string {
	return globalCfg.Conf.ApolloDatasourceConfig.Cluster
}

func ApolloDatasourceIP() string {
	return globalCfg.Conf.ApolloDatasourceConfig.IP
}

func ApolloDatasourceNamespaceName() string {
	return globalCfg.Conf.ApolloDatasourceConfig.NamespaceName
}

func ApolloDatasourceIsBackupConfig() bool {
	return globalCfg.Conf.ApolloDatasourceConfig.IsBackupConfig
}

func ApolloDatasourceSecret() string {
	return globalCfg.Conf.ApolloDatasourceConfig.Secret
}

// RedisClusterHost redis host
func RedisClusterHost() string {
	return globalCfg.Conf.RedisClusterConfig.Host
}

// RedisClusterPort redis port
func RedisClusterPort() int64 {
	return globalCfg.Conf.RedisClusterConfig.Port
}

// RedisClusterPassword redis集群密码
func RedisClusterPassword() string {
	return globalCfg.Conf.RedisClusterConfig.Password
}

// RedisClusterDatabase redis集群DB索引
func RedisClusterDatabase() int {
	return globalCfg.Conf.RedisClusterConfig.Database
}

// RedisIsCluster redis是否为集群
func RedisIsCluster() bool {
	return globalCfg.Conf.RedisClusterConfig.IsCluster
}
