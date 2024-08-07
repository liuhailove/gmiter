package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/liuhailove/gmiter/constants"
	"github.com/liuhailove/gmiter/logging"
	"github.com/liuhailove/gmiter/spi"
	"github.com/liuhailove/gmiter/transport/common/command"
	"github.com/liuhailove/gmiter/transport/common/transport/config"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	PortUninitialized      = -1
	DefaultServerSoTimeout = 3000
	DefaultPort            = 8719
	ServerErrorMessage     = "Command server error"
)

var (
	handlerMap = make(map[string]command.Handler)
)

type SimpleHttpCommandCenter struct {
	server *http.Server
}

func (s SimpleHttpCommandCenter) BeforeStart() error {
	// Register handlers
	handlerMap = command.ProviderInst().NamedHandlers()
	return nil
}

func (s SimpleHttpCommandCenter) Start() error {
	var err error
	// 创建路由器
	mux := http.NewServeMux()
	// 设置路由规则
	mux.HandleFunc("/acquireClusterToken", s.acquireClusterToken)
	mux.HandleFunc("/", s.handle)
	//mux.HandleFunc("/jsonTree", s.fetchJsonTreeHandler)
	var ln net.Listener
	if ln, err = net.Listen("tcp4", ":0"); err != nil {
		logging.Warn("net.Listen err,%s", err)
		return err
	}
	host, port, err := getRealHost(ln)
	// 创建服务器
	s.server = &http.Server{
		Addr:         host,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 30,
		Handler:      mux,
		IdleTimeout:  180 * time.Second,
	}
	// 监听端口并提供服务
	go listenAndServer(ln, s.server)
	config.SetRuntimePort(port)
	return nil
}

func listenAndServer(ln net.Listener, server *http.Server) {
	err := server.Serve(ln)
	if err != nil {
		logging.Warn("listenAndServer error", "error", err)
		return
	}
}

func (s SimpleHttpCommandCenter) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	if err := s.server.Shutdown(ctx); err != nil {
		logging.Warn("server closed", "error", err)
	}
	defer cancel()
	config.SetRuntimePort(PortUninitialized)
	return nil
}

// getRealHost 获取应用host
func getRealHost(ln net.Listener) (host string, port int, err error) {
	adds, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	var localIPV4 string
	var nonLocalIPV4 string
	for _, addr := range adds {
		if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil {
			if ipNet.IP.IsLoopback() {
				localIPV4 = ipNet.IP.String()
			} else {
				nonLocalIPV4 = ipNet.IP.String()
			}
		}
	}
	if nonLocalIPV4 != "" {
		port = ln.Addr().(*net.TCPAddr).Port
		host = fmt.Sprintf("%s:%d", nonLocalIPV4, port)
	} else {
		port = ln.Addr().(*net.TCPAddr).Port
		host = fmt.Sprintf("%s:%d", localIPV4, port)
	}
	return
}

func (s SimpleHttpCommandCenter) acquireClusterToken(writer http.ResponseWriter, request *http.Request) {
	var uri = request.RequestURI
	var rule = ""
	var acquireCount int64 = 0
	var prioritized int64 = 0
	var parameterMap = s.getParameterMap(uri)
	if parameterMap["rule"] != nil {
		rule = parameterMap["rule"][0]
	}
	if parameterMap["acquireCount"] != nil && parameterMap["acquireCount"][0] != "" {
		acquireCount, _ = strconv.ParseInt(parameterMap["acquireCount"][0], 10, 64)
	}
	if parameterMap["prioritized"] != nil && parameterMap["prioritized"][0] != "" {
		prioritized, _ = strconv.ParseInt(parameterMap["prioritized"][0], 10, 64)
	}
	var tokenResult = spi.GetRegisterTokenServiceInst(constants.DefaultTokenServiceType).GetTokenService().RequestToken(rule, uint32(acquireCount), int32(prioritized))
	writer.WriteHeader(http.StatusOK)
	// Here we directly use `toString` to encode the result to plain text.
	writer.Write([]byte(tokenResult.TokenToThinString()))
}

// handle 请求处理
func (s SimpleHttpCommandCenter) handle(writer http.ResponseWriter, request *http.Request) {
	var start = time.Now().UnixNano()
	var uri = request.RequestURI
	var commandRequest = command.NewRequest()
	var path = s.getPath(uri)
	var parameterMap = s.getParameterMap(uri)
	for key, value := range parameterMap {
		if value != nil && len(value) >= 1 {
			commandRequest.AddParam(key, value[0])
		}
	}
	var h = handlerMap[path]
	if h == nil {
		logging.Warn("[SimpleHttpCommandCenter] h not exist", "handler", h)
		return
	}
	var response = h.Handle(*commandRequest)
	if response != nil {
		s.handleResponse(response, writer)
	} else {
		response = command.OfFailure(errors.New("response is nil"))
	}
	var cost = time.Now().UnixNano() - start
	logging.Debug("[seaApiHandler] ", "Deal request", path, "cost(ms)", time.Duration(cost)/time.Millisecond)
}

// handleResponse 响应处理
func (s SimpleHttpCommandCenter) handleResponse(response *command.Response, writer http.ResponseWriter) {
	if response.IsSuccess() {
		writer.WriteHeader(http.StatusOK)
		if response.GetResult() == nil {
			writer.Write(nil)
			return
		}
		// Here we directly use `toString` to encode the result to plain text.
		writer.Write([]byte(response.GetResult().(string)))
	} else {
		writer.WriteHeader(http.StatusBadRequest)
		var msg = ServerErrorMessage
		if response.GetException() != nil {
			msg = response.GetException().Error()
		}
		writer.Write([]byte(msg))
	}
}
func (s SimpleHttpCommandCenter) getPath(uri string) string {
	var path = strings.Split(uri, "?")[0]
	if strings.HasPrefix(path, "/") {
		return path[1:]
	}
	return path
}
func (s SimpleHttpCommandCenter) getParameterMap(uri string) map[string][]string {
	var parameterMap = make(map[string][]string)
	var paramIdx = strings.LastIndex(uri, "?")
	if paramIdx == -1 {
		return parameterMap
	}
	var paramSubStr = uri[paramIdx+1:]
	values, err := url.ParseQuery(paramSubStr)
	if err != nil {
		logging.Warn("[SimpleHttpCommandCenter] uri parse error", "uri", uri, "err", err)
		return parameterMap
	}
	for k, v := range values {
		parameterMap[k] = v
	}
	return parameterMap
}
