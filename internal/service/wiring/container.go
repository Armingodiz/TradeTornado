package wiring

import (
	"context"

	configs "tradeTornado/config"

	"github.com/sirupsen/logrus"

	"tradeTornado/internal/service"
	"tradeTornado/internal/service/provider"
)

type ContainerBuilder struct {
	cnf                              configs.Configs
	masterDb                         *provider.PGProvider
	slaveDb                          *provider.PGProvider
	threadPool                       *service.ExecutorRegistry
	migrationRegistry                *service.MigrationRegistry
	prometheusService                *provider.PrometheusMetricsServer
	kafkaCreateOrderConsumerProvider *provider.KafkaConsumerProvider
	kafkaProducerProvider            *provider.KafkaProducerProvider
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
	c.initThreadPool()
	c.initMetrics()
	c.GetThreadPool().Run(ctx)
	return nil
}

func (c *ContainerBuilder) initThreadPool() {
	pool := c.GetThreadPool()
	pool.AddExecutor(c.GetMasterDB())
	pool.AddExecutor(c.GetSlaveDB())
	pool.AddExecutor(c.GetKafkaCreateOrderConsumerProvider())
	pool.AddExecutor(c.NewOrderEventHandler())
	pool.AddExecutor(c.GetMetricsService())
}

func (c *ContainerBuilder) initMigrationRegistry() {
	session := c.NewMasterGormSession()
	c.getMigrationRegistry().RegisterMigration("orders", c.NewOrderWriteRepositoryTx(session))
}

func (c *ContainerBuilder) getMigrationRegistry() *service.MigrationRegistry {
	if c.migrationRegistry == nil {
		c.migrationRegistry = service.NewMigrationRegistry(c.NewMasterGormSession())
		c.initMigrationRegistry()
	}
	return c.migrationRegistry
}

func (c *ContainerBuilder) NewMasterGormSession() *provider.GormSession {
	return provider.NewGormSession(c.GetMasterDB().DB)
}

func (c *ContainerBuilder) NewSlaveGormSession() *provider.GormSession {
	return provider.NewGormSession(c.GetSlaveDB().DB)
}

func (c *ContainerBuilder) GetMasterDB() *provider.PGProvider {
	if c.masterDb == nil {
		pg, err := provider.NewPostgresConnection(c.cnf.MasterDatabase)
		for err != nil {
			logrus.Panic(err)
		}
		c.masterDb = pg
	}
	return c.masterDb
}

func (c *ContainerBuilder) GetSlaveDB() *provider.PGProvider {
	if c.slaveDb == nil {
		pg, err := provider.NewPostgresConnection(c.cnf.SlaveDatabase)
		for err != nil {
			logrus.Panic(err)
		}
		c.slaveDb = pg
	}
	return c.slaveDb
}
