package etcdv3

//import (
//	"errors"
//	"github.com/coreos/etcd/clientv3"
//	"github.com/liuhailove/gmiter/core/config"
//	"github.com/liuhailove/gmiter/ext/datasource"
//	"github.com/liuhailove/gmiter/ext/datasource/util"
//	"github.com/liuhailove/gmiter/logging"
//	util2 "github.com/liuhailove/gmiter/util"
//	"strings"
//	"time"
//)
//
//var (
//	isInitialized util2.AtomicBool
//)
//
//func Initialize() {
//	if !isInitialized.CompareAndSet(false, true) {
//		return
//	}
//	if config.RuleConsistentModeType() == config.EtcdMode {
//		// 流控规则
//		flowHandler := datasource.NewFlowRulesHandler(datasource.FlowRuleJsonArrayParser)
//		// 如果地址为空，则返回
//		if strings.Trim(config.EtcdDatasourceEndpoints(), "") == "" {
//			logging.Error(errors.New("etcd address is nil"), "etcd address is nil")
//			return
//		}
//		// 配置etcd
//		etcdConfig := clientv3.Config{Endpoints: strings.Split(config.EtcdDatasourceEndpoints(), ","), DialTimeout: time.Second * 30}
//		//创建一个客户端
//		var client *clientv3.Client
//		var err error
//		client, err = clientv3.New(etcdConfig)
//		if err != nil {
//			logging.Error(err, "connect etcd failed")
//			return
//		}
//
//		dsFlowRule, err := NewDataSource(client, config.EtcdDatasourceKeyPrefix()+"/"+config.AppName()+"/"+config.FlowRuleName(), flowHandler)
//		if err != nil {
//			logging.Error(err, "create etcd flow rule failed")
//			return
//		}
//		err = dsFlowRule.Initialize()
//		if err != nil {
//			logging.Error(err, "DsFlowRule Fail to Initialize etcd datasource error", err)
//			return
//		}
//		util.RegisterFlowDataSource(dsFlowRule)
//
//		// 授权规则
//		dsAuthorityRule, err := NewDataSource(client, config.AuthorityRuleName())
//		if err != nil {
//			logging.Error(err, "create etcd authority rule failed")
//			return
//		}
//		err = dsAuthorityRule.Initialize()
//		if err != nil {
//			logging.Error(err, "DsAuthorityRule Fail to Initialize etcd datasource error", err)
//			return
//		}
//		util.RegisterAuthorityDataSource(dsAuthorityRule)
//
//		// 降级规则
//		circuitHandler := datasource.NewCircuitBreakerRulesHandler(datasource.CircuitBreakerRuleJsonArrayParser)
//		dsDegradeRule, err := NewDataSource(client, config.EtcdDatasourceKeyPrefix()+"/"+config.AppName()+"/"+config.DegradeRuleName(), circuitHandler)
//		if err != nil {
//			logging.Error(err, "create etcd circuitBreaker rule failed")
//			return
//		}
//		err = dsDegradeRule.Initialize()
//		if err != nil {
//			logging.Error(err, "DsDegradeRule Fail to Initialize etcd datasource error", err)
//			return
//		}
//		util.RegisterDegradeDataSource(dsDegradeRule)
//
//		// 降级规则
//		systemHandler := datasource.NewSystemRulesHandler(datasource.SystemRuleJsonArrayParser)
//		dsSystemRule, err := NewDataSource(client, config.EtcdDatasourceKeyPrefix()+"/"+config.AppName()+"/"+config.SystemRuleName(), systemHandler)
//		if err != nil {
//			logging.Error(err, "create etcd system rule failed")
//			return
//		}
//		err = dsSystemRule.Initialize()
//		if err != nil {
//			logging.Error(err, "DsSystemRule Fail to Initialize etcd datasource error", err)
//			return
//		}
//		util.RegisterSystemDataSource(dsSystemRule)
//
//		// 热点规则
//		hotspotHandler := datasource.NewHotSpotParamRulesHandler(datasource.HotSpotParamRuleJsonArrayParser)
//		dsHotspotRule, err := NewDataSource(client, config.EtcdDatasourceKeyPrefix()+"/"+config.AppName()+"/"+config.HotspotRuleName(), hotspotHandler)
//		if err != nil {
//			logging.Error(err, "create etcd hotspot rule failed")
//			return
//		}
//		err = dsHotspotRule.Initialize()
//		if err != nil {
//			logging.Error(err, "DsHotspotRule Fail to Initialize datasource error", err)
//			return
//		}
//		util.RegisterHotspotSource(dsHotspotRule)
//
//		// mock规则
//		mockHandler := datasource.NewMockRulesHandler(datasource.MockRuleJsonArrayParser)
//		dsMockRule, err := NewDataSource(client, config.EtcdDatasourceKeyPrefix()+"/"+config.AppName()+"/"+config.MockRuleName(), mockHandler)
//		if err != nil {
//			logging.Error(err, "create etcd mock rule failed")
//			return
//		}
//		err = dsMockRule.Initialize()
//		if err != nil {
//			logging.Error(err, "DsMockRule Fail to Initialize etcd datasource error", err)
//			return
//		}
//		util.RegisterMockDataSource(dsMockRule)
//
//		// retry规则
//		retryHandler := datasource.NewRetryRulesHandler(datasource.RetryRuleJsonArrayParser)
//		if err != nil {
//			logging.Error(err, "create etcd retry rule failed")
//			return
//		}
//		dsRetryRule, err := NewDataSource(client, config.EtcdDatasourceKeyPrefix()+"/"+config.AppName()+"/"+config.RetryRuleName(), retryHandler)
//		err = dsRetryRule.Initialize()
//		if err != nil {
//			logging.Error(err, "DsRetryRule Fail to Initialize datasource error", err)
//			return
//		}
//		util.RegisterRetryDataSource(dsRetryRule)
//
//		// gray规则
//		grayHandler := datasource.NewGrayRulesHandler(datasource.GrayRuleJsonArrayParser)
//		dsGrayRule, err := NewDataSource(client, config.EtcdDatasourceKeyPrefix()+"/"+config.AppName()+"/"+config.GrayRuleName(), grayHandler)
//		if err != nil {
//			logging.Error(err, "create etcd gray rule failed")
//			return
//		}
//		err = dsGrayRule.Initialize()
//		if err != nil {
//			logging.Error(err, "dsGrayRule Fail to Initialize datasource error", err)
//			return
//		}
//		util.RegisterGrayDataSource(dsGrayRule)
//
//		// isolation规则
//		isolationHandler := datasource.NewGrayRulesHandler(datasource.IsolationRuleJsonArrayParser)
//		dsIsolationRule, err := NewDataSource(client, config.EtcdDatasourceKeyPrefix()+"/"+config.AppName()+"/"+config.IsolationRuleName(), isolationHandler)
//		if err != nil {
//			logging.Error(err, "create etcd gray rule failed")
//			return
//		}
//		err = dsIsolationRule.Initialize()
//		if err != nil {
//			logging.Error(err, "dsIsolationRule Fail to Initialize datasource error", err)
//			return
//		}
//		util.RegisterIsolationDataSource(dsIsolationRule)
//
//		// weightRouter规则
//		weightRouterHandler := datasource.NewWeightRouterRulesHandler(datasource.WeightRouterRuleJsonArrayParser)
//		weightRouterRule, err := NewDataSource(client, config.EtcdDatasourceKeyPrefix()+"/"+config.AppName()+"/"+config.WeightRouterRuleName(), weightRouterHandler)
//		if err != nil {
//			logging.Error(err, "create etcd weight rule failed")
//			return
//		}
//		err = weightRouterRule.Initialize()
//		if err != nil {
//			logging.Error(err, "weightRouterRule Fail to Initialize etcd datasource error", err)
//			return
//		}
//		util.RegisterWeightRouterDataSource(weightRouterRule)
//	}
//}
