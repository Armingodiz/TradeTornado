package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	MaxIdle      int
	MaxActive    int
	IsProduction bool
}

type RedisProvider struct {
	Pool *redis.Pool
	cnf  RedisConfig
}

func (receiver *RedisProvider) GetRepresentation() string {
	return "RedisProvider"
}

func (receiver *RedisProvider) Recover(ctx context.Context) error {
	if err := receiver.IsAlive(ctx); err != nil {
		logrus.WithField("error", err.Error()).Errorln("try to reconnect...")
		redisPool := NewRedisPool(receiver.cnf)
		if err != nil {
			return err
		}
		receiver.Pool = redisPool
	}
	return nil
}

func (receiver *RedisProvider) Run(ctx context.Context) error {
	if err := receiver.Recover(ctx); err != nil {
		return err
	}
	for {
		select {
		case <-time.After(time.Minute):
			if err := receiver.Recover(ctx); err != nil {
				logrus.Errorln(err)
			}
		case <-ctx.Done():
			logrus.Infoln(" Shutting down Redis...")
			if receiver.Pool != nil {
				return receiver.Pool.Close()
			}
			return nil
		}
	}
}

func (receiver *RedisProvider) IsAlive(ctx context.Context) error {
	if receiver.Pool == nil {
		return errors.New("redis is not alive")
	}
	conn := receiver.Pool.Get()
	defer conn.Close()
	_, err := redis.String(receiver.Pool.Get().Do("PING"))
	return err
}

func NewRedisPool(config RedisConfig) *redis.Pool {
	return &redis.Pool{
		MaxIdle:   config.MaxIdle,
		MaxActive: config.MaxActive,
		Dial: func() (redis.Conn, error) {
			addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
			conn, err := redis.Dial("tcp", addr, redis.DialPassword(config.Password))
			if err != nil {
				log.Printf("ERROR: fail init redis pool: %s", err.Error())
				os.Exit(1)
			}
			return conn, err
		},
	}
}
