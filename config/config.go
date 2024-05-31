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
	RedisConfig              provider.RedisConfig
	KafkaConsumerConfig      provider.KafkaConsumerConfig
	KafkaProducerConfig      provider.KafkaProducerConfig
	OrderCreateTopic         string
	OrderMatchedTopic        string
	OrderCreateConsumerGroup string
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
		RedisConfig: provider.RedisConfig{
			Host:      lib.GetEnv("REDIS_HOST", "localhost"),
			Port:      lib.GetEnv("REDIS_PORT", "6379"),
			Password:  lib.GetEnv("REDIS_PASSWORD", ""),
			MaxIdle:   cast.ToInt(lib.GetEnv("REDIS_MAX_IDLE", "0")),
			MaxActive: cast.ToInt(lib.GetEnv("REDIS_MAX_ACTIVE", "0")),
		},
		KafkaConsumerConfig: provider.KafkaConsumerConfig{
			Brokers:   lib.GetEnv("KAFKA_BROKERS", "localhost:29092"),
			BatchSize: 100,
		},
		KafkaProducerConfig: provider.KafkaProducerConfig{
			Brokers: lib.GetEnv("KAFKA_BROKERS", "localhost:29092"),
		},
		OrderCreateTopic:         "order-events",
		OrderMatchedTopic:        "order-matches",
		OrderCreateConsumerGroup: "matcher",
	}
}
