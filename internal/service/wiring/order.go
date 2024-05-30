package wiring

import (
	"log"
	"tradeTornado/internal/modules/order/application"
	"tradeTornado/internal/modules/order/infrastructure"
	"tradeTornado/internal/service/provider"
)

func (c *ContainerBuilder) GetOrdereController() *infrastructure.OrderController {
	return infrastructure.NewOrderController(c.NewOrdereQueryHandler())
}

func (c *ContainerBuilder) NewOrdereQueryHandler() *application.OrderQueryHandler {
	return application.NewOrderQueryHandler(c.NewOrderRepository())
}

func (c *ContainerBuilder) NewOrderRepository() *infrastructure.OrderRepository {
	return infrastructure.NewOrderRepository(c.NewMasterGormSession())
}

func (c *ContainerBuilder) NewOrderRepositoryTx(session *provider.GormSession) *infrastructure.OrderRepository {
	return infrastructure.NewOrderRepository(session)
}

func (c *ContainerBuilder) NewOrderEventHandler() *application.OrderEventHandler {
	return application.NewOrderEventHandler(c.GetKafkaConsumerProvider(), c.NewOrderRepository())
}

func (c *ContainerBuilder) GetKafkaConsumerProvider() *provider.KafkaConsumerProvider {
	if c.kafkaConsumerProvider == nil {
		pv, err := provider.NewKafkaConsumerProvider(c.cnf.KafkaConsumerConfig)
		if err != nil {
			log.Fatalln(err)
		}
		c.kafkaConsumerProvider = pv
	}
	return c.kafkaConsumerProvider
}
