package micro

var (
	_grpcPort  int64  = 0
	_severName string = ""
)

// SetGrpcPort 设置GRPC接口
func SetGrpcPort(grpcPort int64) {
	_grpcPort = grpcPort
}

// GetGrpcPort 获取Grpc接口
func GetGrpcPort() int64 {
	return _grpcPort
}

// SetServerName 设置服务名称
func SetServerName(serverName string) {
	_severName = serverName
}

// GetServerName 获取服务名称
func GetServerName() string {
	return _severName
}
