package wiring

import (
	"log"
	"tradeTornado/internal/modules/order"
	"tradeTornado/internal/modules/order/application"
	"tradeTornado/internal/modules/order/infrastructure"
	"tradeTornado/internal/service/provider"
)

func (c *ContainerBuilder) NewOrdereController() *infrastructure.OrderController {
	return infrastructure.NewOrderController(c.NewOrdereQueryHandler())
}

func (c *ContainerBuilder) NewOrdereQueryHandler() *application.OrderQueryHandler {
	return application.NewOrderQueryHandler(c.NewOrderReadRepository())
}

func (c *ContainerBuilder) NewOrderWriteRepository() *infrastructure.OrderRepository {
	return infrastructure.NewOrderRepository(c.NewMasterGormSession())
}

func (c *ContainerBuilder) NewOrderReadRepository() *infrastructure.OrderRepository {
	return infrastructure.NewOrderRepository(c.NewSlaveGormSession())
}

func (c *ContainerBuilder) NewOrderWriteRepositoryTx(session *provider.GormSession) *infrastructure.OrderRepository {
	return infrastructure.NewOrderRepository(session)
}

func (c *ContainerBuilder) NewOrderEventHandler() *application.OrderEventHandler {
	return application.NewOrderEventHandler(c.GetKafkaCreateOrderConsumerProvider(),
		c.GetKafkaProducerProvider(),
		c.cnf.OrderMatchedTopic,
		func() order.IOrderWriteRepository {
			return c.NewOrderWriteRepository()
		})
}

func (c *ContainerBuilder) GetKafkaCreateOrderConsumerProvider() *provider.KafkaConsumerProvider {
	if c.kafkaCreateOrderConsumerProvider == nil {
		pv, err := provider.NewKafkaConsumerProvider(c.cnf.KafkaConsumerConfig, c.GetKafkaProducerProvider(), c.cnf.OrderCreateTopic, c.cnf.OrderCreateConsumerGroup)
		if err != nil {
			log.Fatalln(err)
		}
		c.kafkaCreateOrderConsumerProvider = pv
	}
	return c.kafkaCreateOrderConsumerProvider
}

func (c *ContainerBuilder) GetKafkaProducerProvider() *provider.KafkaProducerProvider {
	if c.kafkaProducerProvider == nil {
		pv, err := provider.NewKafkaProducer(c.cnf.KafkaProducerConfig)
		if err != nil {
			log.Fatalln(err)
		}
		c.kafkaProducerProvider = pv
	}
	return c.kafkaProducerProvider
}
