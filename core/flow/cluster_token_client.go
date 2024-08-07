package flow

import (
	"bytes"
	"github.com/liuhailove/gmiter/constants"
	"github.com/liuhailove/gmiter/spi"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/liuhailove/gmiter/core/base"
	"github.com/liuhailove/gmiter/logging"
	"github.com/liuhailove/gmiter/transport/common/transport/config"
	"github.com/liuhailove/gmiter/util"
)

const (
	AcquireClusterTokenPath = "/acquireClusterToken"
	HttpProtocol            = "http://"
)

var (
	TokenClient *DefaultClusterTokenClient
)

func init() {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
		MaxConnsPerHost:       200,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}
	httpClient := http.Client{
		// 读取3000ms要超时，避免影响业务
		Timeout:   time.Duration(3000) * time.Millisecond,
		Transport: transport,
	}

	TokenClient = &DefaultClusterTokenClient{
		httpClient: httpClient,
	}
}

type ClusterTokenClient interface {
}

// DefaultClusterTokenClient 默认TokenClient
type DefaultClusterTokenClient struct {
	// TokenServerIp TokenServer所在IP
	TokenServerIp string
	// TokenServerPort TokenServer所在port
	TokenServerPort int32
	// httpClient
	httpClient http.Client
	//retryClient *retryablehttp.Client
}

func (d *DefaultClusterTokenClient) RequestToken(tokenServerIp string, tokenServerPort int32, ruleId string, acquireCount uint32, prioritized int32) *base.TokResult {
	if d.notValidRequest(ruleId, acquireCount) {
		return base.BadResult
	}
	// 获取集群规则
	var tcController = getTrafficControllerFor(ruleId)
	if tcController == nil {
		return base.NoRuleExistsResult
	}
	var err error
	var ruleData []byte
	ruleData, err = jsonTraffic.Marshal(tcController.rule)
	if err != nil {
		logging.Error(err, "unmarshal failed")
		return base.BadResult
	}
	// 如果请求的IP和port和本机相等，说明自身就是master，不需要在经过网络
	if tokenServerIp == util.GetIP() && strconv.Itoa(int(tokenServerPort)) == config.GetPort() {
		var inst = spi.GetRegisterTokenServiceInst(constants.DefaultTokenServiceType)
		if inst != nil {
			return inst.GetTokenService().RequestToken(string(ruleData), acquireCount, prioritized)
		}
	}
	// 组装请求
	var requestData = make(map[string]string)
	requestData["rule"] = string(ruleData)
	requestData["acquireCount"] = strconv.Itoa(int(acquireCount))
	requestData["prioritized"] = strconv.Itoa(int(prioritized))
	var tokenResult, err2 = d.sendTokenRequest(tokenServerIp, tokenServerPort, AcquireClusterTokenPath, requestData)
	if err2 != nil {
		logging.Error(err2, "sendTokenRequest error", "tokenServerIp", tokenServerIp, "tokenServerPort", tokenServerPort, "ruleId", ruleId, "acquireCount", acquireCount)
		return base.FailResult
	}
	logging.Info("tokenResult", "tokenResult", tokenResult)
	return tokenResult
}

func (d *DefaultClusterTokenClient) sendTokenRequest(tokenServerIp string, tokenServerPort int32, requestPath string, paramsMap map[string]string) (*base.TokResult, error) {
	var httpClient = d.httpClient
	var url = HttpProtocol + tokenServerIp
	if tokenServerPort > 0 {
		url += ":" + strconv.Itoa(int(tokenServerPort))
	}
	requestPath = url + requestPath
	requestPath = d.getRequestPath(http.MethodGet, requestPath, paramsMap, "")
	var err error
	if err != nil {
		logging.Warn("[SimpleHttpClient] request error", "msg", err)
		return nil, err
	}
	req, err := http.NewRequest("GET", requestPath, nil)
	req.Header.Set("Connection", "keep-alive")
	resp, err := httpClient.Do(req)

	if err != nil {
		logging.Warn("[SimpleHttpClient] request do error", "msg", err)
		return nil, err
	}
	defer resp.Body.Close()

	// 解析Resp
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logging.Warn("[ClusterTokenClient] ReadAll err", "error msg", err)
		return base.FailResult, err
	}

	tokenResult, err := base.ThinStringToToken(string(body))
	if err != nil {
		logging.Warn("[ClusterTokenClient] Unmarshal err", "error msg", err)
		return base.FailResult, err
	}
	return tokenResult, nil
}

// GetRequestPath
// 获取请求路径
func (d *DefaultClusterTokenClient) getRequestPath(methodType string, requestPath string, paramsMap map[string]string, charset string) string {
	if methodType == http.MethodGet {
		if strings.Contains(requestPath, "?") {
			return requestPath + "&" + d.encodeRequestParams(paramsMap, charset)
		}
		return requestPath + "?" + d.encodeRequestParams(paramsMap, charset)
	}
	return requestPath
}

// encodeRequestParams
// * Encode and get the URL request parameters.
// *
// * @param paramsMap pair of parameters
// * @param charset   charset
//   - @return encoded request parameters, or empty string ("") if no parameters are provided
func (d *DefaultClusterTokenClient) encodeRequestParams(paramsMap map[string]string, charset string) string {
	if paramsMap == nil || len(paramsMap) == 0 {
		return ""
	}
	var paramsBuilder bytes.Buffer
	for k, v := range paramsMap {
		if k == "" || v == "" {
			continue
		}
		paramsBuilder.WriteString(k)
		paramsBuilder.WriteString("=")
		paramsBuilder.WriteString(v)
		paramsBuilder.WriteString("&")
	}

	if paramsBuilder.Len() > 0 {
		paramsBuilder.Truncate(paramsBuilder.Len() - 1)
	}
	return paramsBuilder.String()
}

func (d *DefaultClusterTokenClient) notValidRequest(id string, count uint32) bool {
	return id == "" || count <= 0
}
