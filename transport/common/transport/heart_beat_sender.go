package transport

// HeartBeatSender  The heartbeat sender which is responsible for sending heartbeat to remote dashboard
//  periodically per {@code interval}.
type HeartBeatSender interface {

	// SendHeartbeat
	// Send heartbeat to sea Dashboard. Each invocation of this method will send
	// heartbeat once. sea core is responsible for invoking this method
	// at every {@link #intervalMs()} interval.
	//
	// @return whether heartbeat is successfully send.
	// @throws Exception if error occurs
	SendHeartbeat() (bool, error)

	// IntervalMs
	// Default interval in milliseconds of the sender. It would take effect only when
	// the heartbeat interval is not configured in sea config property.
	//
	// @return default interval of the sender in milliseconds
	//
	IntervalMs() uint64

	// SendRemove 发送移除消息，此方法需要在容器销毁前执行
	SendRemove() (bool, error)
}
