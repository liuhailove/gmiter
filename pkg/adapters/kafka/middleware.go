package kafka

import (
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"github.com/Shopify/sarama"

	sea "git.garena.com/honggang.liu/seamiter-go/api"
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/core/config"
)

// SeaConsumerInterceptor 定义消费之拦截器
type SeaConsumerInterceptor struct {
}

// OnConsume is called when the consumed message is intercepted. Please
// avoid modifying the message until it's safe to do so, as this is _not_ a
// copy of the message.
func (c *SeaConsumerInterceptor) OnConsume(message *sarama.ConsumerMessage) {
	if !config.CloseAll() {
		resourceName := "Kafka:" + message.Topic
		logging.Info("OnKafkaConsume", "resourceName", resourceName)
		var entry, blockErr = sea.Entry(
			resourceName,
			sea.WithResourceType(base.ResTypeMQ),
			sea.WithTrafficType(base.Inbound),
		)
		if blockErr != nil {
			logging.Info("OnKafkaConsumeBlocked", "resourceName", resourceName)
			panic(blockErr)
		}
		logging.Info("OnKafkaConsumePass", "resourceName", resourceName)
		defer entry.Exit()
	}
	return
}

// NewConsumerInterceptor 生成拦截器
func NewConsumerInterceptor() *SeaConsumerInterceptor {
	c := SeaConsumerInterceptor{}
	return &c
}
