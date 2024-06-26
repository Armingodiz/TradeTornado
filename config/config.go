package configs

import (
	"tradeTornado/internal/lib"
	"tradeTornado/internal/service/provider"

	"github.com/spf13/cast"
)

type Configs struct {
	AppName                  string
	MasterDatabase           provider.PostgresConfig
	SlaveDatabase            provider.PostgresConfig
	MetricConfig             *provider.PrometheusConfig
	IsProduction             bool
	KafkaConsumerConfig      provider.KafkaConsumerConfig
	KafkaProducerConfig      provider.KafkaProducerConfig
	OrderCreateTopic         string
	OrderMatchedTopic        string
	OrderCreateConsumerGroup string
	ServerConfigs            provider.ServerConfigs
}

func ConfigFromEnv() Configs {
	return Configs{
		AppName:      lib.GetEnv("APP_NAME", "tradeTornado"),
		IsProduction: cast.ToBool(lib.GetEnv("IS_PRODUCTION", "FALSE")),
		MasterDatabase: provider.PostgresConfig{
			Host:              lib.GetEnv("POSTGRES_MASTER_HOST", "localhost"),
			Port:              lib.GetEnv("POSTGRES_MASTER_PORT", "5432"),
			UserName:          lib.GetEnv("POSTGRES_MASTER_USER", "admin"),
			Password:          lib.GetEnv("POSTGRES_MASTER_PASS", "adminpassword"),
			DB:                lib.GetEnv("POSTGRES_MASTER_DB", "tradeTornado"),
			MaxConnection:     lib.GetEnv("POSTGRES_MASTER_MAX_CONNECTION", "30"),
			MaxIdleConnection: lib.GetEnv("POSTGRES_MASTER_MAX_IDLE", "4"),
			IsProduction:      cast.ToBool(lib.GetEnv("IS_PRODUCTION", "FALSE")),
		},
		SlaveDatabase: provider.PostgresConfig{
			Host:              lib.GetEnv("POSTGRES_SLAVE_HOST", "localhost"),
			Port:              lib.GetEnv("POSTGRES_SLAVE_PORT", "5433"),
			UserName:          lib.GetEnv("POSTGRES_SLAVE_USER", "admin"),
			Password:          lib.GetEnv("POSTGRES_SLAVE_PASS", "adminpassword"),
			DB:                lib.GetEnv("POSTGRES_SLAVE_DB", "tradeTornado"),
			MaxConnection:     lib.GetEnv("POSTGRES_SLAVE_MAX_CONNECTION", "30"),
			MaxIdleConnection: lib.GetEnv("POSTGRES_SLAVE_MAX_IDLE", "4"),
			IsProduction:      cast.ToBool(lib.GetEnv("IS_PRODUCTION", "FALSE")),
		},
		MetricConfig: &provider.PrometheusConfig{
			Port:    lib.GetEnv("METRIC_PORT", "9095"),
			Disable: cast.ToBool(lib.GetEnv("MONITOR_DISABLE", "false")),
		},
		KafkaConsumerConfig: provider.KafkaConsumerConfig{
			Brokers:   lib.GetEnv("KAFKA_BROKERS", "localhost:29092"),
			BatchSize: cast.ToInt(lib.GetEnv("KAFKA_CONSUMER_BATCH_SIZE", "100")),
		},
		KafkaProducerConfig: provider.KafkaProducerConfig{
			Brokers: lib.GetEnv("KAFKA_BROKERS", "localhost:29092"),
		},
		OrderCreateTopic:         lib.GetEnv("KAFKA_ORDER_CREATE_TOPIC", "order-events"),
		OrderMatchedTopic:        lib.GetEnv("KAFKA_ORDER_MATCH_TOPIC", "order-matches"),
		OrderCreateConsumerGroup: lib.GetEnv("KAFKA_ORDER_CREATE_CONSUMER_GROUP", "matcher"),
		ServerConfigs: provider.ServerConfigs{
			Port:           lib.GetEnv("API_PORT", "8080"),
			Name:           lib.GetEnv("API_NAME", "order-matcher"),
			ReadTimeoutMS:  lib.GetEnv("API_READ_TIMEOUT", "50000"),
			WriteTimeoutMS: lib.GetEnv("API_WRITE_TIMEOUT", "10000"),
		},
	}
}
