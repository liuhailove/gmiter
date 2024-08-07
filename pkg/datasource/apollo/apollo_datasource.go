package apollo

import (
	"git.garena.com/honggang.liu/seamiter-go/ext/datasource"
	"git.garena.com/honggang.liu/seamiter-go/ext/datasource/util"
	apollo_config "github.com/apolloconfig/agollo/v4/env/config"

	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	util2 "git.garena.com/honggang.liu/seamiter-go/util"
	"github.com/pkg/errors"
	"strings"
)

var (
	isInitialized util2.AtomicBool
)

func Initialize() {
	if !isInitialized.CompareAndSet(false, true) {
		return
	}
	if config.RuleConsistentModeType() == config.ApolloMode {
		// 如果应用为空则返回
		if strings.Trim(config.ApolloDatasourceAppId(), "") == "" {
			logging.Error(errors.New("apollo datasource appId cannot be empty"), "apollo datasource appId cannot be empty")
			return
		}
		// 如果地址为空，则返回
		if strings.Trim(config.ApolloDatasourceIP(), "") == "" {
			logging.Error(errors.New("apollo datasource ip cannot be empty"), "apollo datasource ip cannot be empty")
			return
		}

		c := &apollo_config.AppConfig{
			AppID:          config.ApolloDatasourceNamespaceName(),
			Cluster:        config.ApolloDatasourceCluster(),
			IP:             config.ApolloDatasourceIP(),
			NamespaceName:  config.ApolloDatasourceNamespaceName(),
			IsBackupConfig: config.ApolloDatasourceIsBackupConfig(),
			Secret:         config.ApolloDatasourceSecret(),
		}
		// 流控规则
		flowHandler := datasource.NewFlowRulesHandler(datasource.FlowRuleJsonArrayParser)

		dsFlowRule, err := NewDatasource(c, config.AppName()+"-"+config.FlowRuleName(), WithPropertyHandlers(flowHandler))
		if err != nil {
			logging.Error(err, "create apollo flow rule failed")
			return
		}
		err = dsFlowRule.Initialize()
		if err != nil {
			logging.Error(err, "DsFlowRule Fail to Initialize apollo datasource error", err)
			return
		}
		util.RegisterFlowDataSource(dsFlowRule)

		// 授权规则
		dsAuthorityRule, err := NewDatasource(c, config.AppName()+"-"+config.AuthorityRuleName())
		if err != nil {
			logging.Error(err, "create apollo authority rule failed")
			return
		}
		err = dsAuthorityRule.Initialize()
		if err != nil {
			logging.Error(err, "DsAuthorityRule Fail to Initialize aploo datasource error", err)
			return
		}
		util.RegisterAuthorityDataSource(dsAuthorityRule)

		// 降级规则
		circuitHandler := datasource.NewCircuitBreakerRulesHandler(datasource.CircuitBreakerRuleJsonArrayParser)
		dsDegradeRule, err := NewDatasource(c, config.AppName()+"-"+config.DegradeRuleName(), WithPropertyHandlers(circuitHandler))
		if err != nil {
			logging.Error(err, "create apollo circuitBreaker rule failed")
			return
		}
		err = dsDegradeRule.Initialize()
		if err != nil {
			logging.Error(err, "DsDegradeRule Fail to Initialize apollo datasource error", err)
			return
		}
		util.RegisterDegradeDataSource(dsDegradeRule)

		// 降级规则
		systemHandler := datasource.NewSystemRulesHandler(datasource.SystemRuleJsonArrayParser)
		dsSystemRule, err := NewDatasource(c, config.AppName()+"/"+config.SystemRuleName(), WithPropertyHandlers(systemHandler))
		if err != nil {
			logging.Error(err, "create apollo system rule failed")
			return
		}
		err = dsSystemRule.Initialize()
		if err != nil {
			logging.Error(err, "DsSystemRule Fail to Initialize apollo datasource error", err)
			return
		}
		util.RegisterSystemDataSource(dsSystemRule)

		// 热点规则
		hotspotHandler := datasource.NewHotSpotParamRulesHandler(datasource.HotSpotParamRuleJsonArrayParser)
		dsHotspotRule, err := NewDatasource(c, config.AppName()+"-"+config.HotspotRuleName(), WithPropertyHandlers(hotspotHandler))
		if err != nil {
			logging.Error(err, "create apollo hotspot rule failed")
			return
		}
		err = dsHotspotRule.Initialize()
		if err != nil {
			logging.Error(err, "DsHotspotRule Fail to Initialize apollo datasource error", err)
			return
		}
		util.RegisterHotspotSource(dsHotspotRule)

		// mock规则
		mockHandler := datasource.NewMockRulesHandler(datasource.MockRuleJsonArrayParser)
		dsMockRule, err := NewDatasource(c, config.AppName()+"-"+config.MockRuleName(), WithPropertyHandlers(mockHandler))
		if err != nil {
			logging.Error(err, "create apollo mock rule failed")
			return
		}
		err = dsMockRule.Initialize()
		if err != nil {
			logging.Error(err, "DsMockRule Fail to Initialize apollo datasource error", err)
			return
		}
		util.RegisterMockDataSource(dsMockRule)

		// retry规则
		retryHandler := datasource.NewRetryRulesHandler(datasource.RetryRuleJsonArrayParser)
		if err != nil {
			logging.Error(err, "create apollo retry rule failed")
			return
		}
		dsRetryRule, err := NewDatasource(c, config.AppName()+"-"+config.RetryRuleName(), WithPropertyHandlers(retryHandler))
		err = dsRetryRule.Initialize()
		if err != nil {
			logging.Error(err, "DsRetryRule Fail to Initialize apollo datasource error", err)
			return
		}
		util.RegisterRetryDataSource(dsRetryRule)

		// gray规则
		grayHandler := datasource.NewGrayRulesHandler(datasource.GrayRuleJsonArrayParser)
		dsGrayRule, err := NewDatasource(c, config.AppName()+"-"+config.GrayRuleName(), WithPropertyHandlers(grayHandler))
		if err != nil {
			logging.Error(err, "create apollo gray rule failed")
			return
		}
		err = dsGrayRule.Initialize()
		if err != nil {
			logging.Error(err, "dsGrayRule Fail to Initialize apollo datasource error", err)
			return
		}
		util.RegisterGrayDataSource(dsGrayRule)

		// isolation规则
		isolationHandler := datasource.NewGrayRulesHandler(datasource.IsolationRuleJsonArrayParser)
		dsIsolationRule, err := NewDatasource(c, config.AppName()+"-"+config.IsolationRuleName(), WithPropertyHandlers(isolationHandler))
		if err != nil {
			logging.Error(err, "create apollo gray rule failed")
			return
		}
		err = dsIsolationRule.Initialize()
		if err != nil {
			logging.Error(err, "dsIsolationRule Fail to Initialize apollo datasource error", err)
			return
		}
		util.RegisterIsolationDataSource(dsIsolationRule)

		// weightRouter规则
		weightRouterHandler := datasource.NewWeightRouterRulesHandler(datasource.WeightRouterRuleJsonArrayParser)
		weightRouterRule, err := NewDatasource(c, config.AppName()+"-"+config.WeightRouterRuleName(), WithPropertyHandlers(weightRouterHandler))
		if err != nil {
			logging.Error(err, "create apollo weight rule failed")
			return
		}
		err = weightRouterRule.Initialize()
		if err != nil {
			logging.Error(err, "weightRouterRule Fail to Initialize apollo datasource error", err)
			return
		}
		util.RegisterWeightRouterDataSource(weightRouterRule)
	}
}
