package wiring

import (
	"context"

	configs "tradeTornado/config"

	"github.com/sirupsen/logrus"

	"tradeTornado/internal/lib/messaging"
	"tradeTornado/internal/service"
	"tradeTornado/internal/service/provider"
)

type ContainerBuilder struct {
	cnf                     configs.Configs
	systemDb                *provider.PGProvider
	oldStalinDb             *provider.PGProvider
	threadPool              *service.ExecutorRegistry
	eventBus                *messaging.EventBus
	migrationRegistry       *service.MigrationRegistry
	prometheusService       *provider.PrometheusMetricsServer
	feedbackUserGromSession *provider.GormSession
	kafkaConsumerProvider   *provider.KafkaConsumerProvider
}

func NewContainer(cnf configs.Configs) *ContainerBuilder {
	tmp := &ContainerBuilder{cnf: cnf}
	return tmp
}

func (c *ContainerBuilder) GetThreadPool() *service.ExecutorRegistry {
	if c.threadPool == nil {
		c.threadPool = service.NewExecutorRegistry()
	}
	return c.threadPool
}

func (c *ContainerBuilder) RunMigration(ctx context.Context, name string) error {
	return c.getMigrationRegistry().Run(ctx, name)
}

func (c *ContainerBuilder) Run(ctx context.Context) error {
	if !c.cnf.IsProduction {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	c.initEventBus()
	c.initThreadPool()
	c.initMetrics()
	c.GetThreadPool().Run(ctx)
	return nil
}

func (c *ContainerBuilder) initThreadPool() {
	pool := c.GetThreadPool()
	pool.AddExecutor(c.GetMasterDB())
	pool.AddExecutor(c.GetSlaveDB())
	pool.AddExecutor(c.GetKafkaConsumerProvider())
	pool.AddExecutor(c.NewOrderEventHandler())
	pool.AddExecutor(c.GetMetricsService())
}

func (c *ContainerBuilder) initMigrationRegistry() {
	// todo : pass a GormSession!
	// this will not work by TX
	c.getMigrationRegistry().RegisterMigration("orders", c.NewOrderRepository())
}

func (c *ContainerBuilder) getEventBus() *messaging.EventBus {
	if c.eventBus == nil {
		c.eventBus = messaging.NewEventBus()
	}
	return c.eventBus
}

func (c *ContainerBuilder) getMigrationRegistry() *service.MigrationRegistry {
	if c.migrationRegistry == nil {
		c.migrationRegistry = service.NewMigrationRegistry(c.NewMasterGormSession())
		c.initMigrationRegistry()
	}
	return c.migrationRegistry
}

func (c *ContainerBuilder) NewMasterGormSession() *provider.GormSession {
	return provider.NewGormSession(c.GetMasterDB().DB, c.getEventBus())
}

func (c *ContainerBuilder) NewSlaveGormSession() *provider.GormSession {
	return provider.NewGormSession(c.GetSlaveDB().DB, c.getEventBus())
}

func (c *ContainerBuilder) GetMasterDB() *provider.PGProvider {
	if c.systemDb == nil {
		pg, err := provider.NewPostgresConnection(c.cnf.MasterDatabase)
		for err != nil {
			logrus.Panic(err)
		}
		c.systemDb = pg
	}
	return c.systemDb
}

func (c *ContainerBuilder) GetSlaveDB() *provider.PGProvider {
	if c.systemDb == nil {
		pg, err := provider.NewPostgresConnection(c.cnf.SlaveDatabase)
		for err != nil {
			logrus.Panic(err)
		}
		c.systemDb = pg
	}
	return c.systemDb
}

func (c *ContainerBuilder) initEventBus() {
	// bus := c.getEventBus()
	// marshaller := messaging.NewJsonMarshaller(bus, c.getNatsQueue())
	// c.getNatsQueue().SetPublisher(marshaller)
	// bus.SetPublisher(marshaller)
	// marshaller.RegisterEvent(func() messaging.IEvent { return &feedback.FeedbackChangedEvent{} })
	// marshaller.RegisterEvent(func() messaging.IEvent { return &feedback.FillFeedbackChangesetEvent{} })
	// marshaller.RegisterEvent(func() messaging.IEvent { return &exportation.FeedbackExportProcessCreatedEvent{} })
	// marshaller.RegisterEvent(func() messaging.IEvent { return &exportation.FeedbackExportProcessChangesetCreatedEvent{} })
	// marshaller.RegisterEvent(func() messaging.IEvent { return &exportation.FeedbackExportProcessChangedEvent{} })
	// bus.RegisterAtLeastOnce(c.NewFeedbackEventHandler())
	// bus.RegisterAtLeastOnce(c.NewFeedbackExportEventHandler())
	// bus.RegisterAtLeastOnce(c.NewFeedbackViewEventHandler())
}
