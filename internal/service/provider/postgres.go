package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresConfig struct {
	Host              string
	Port              string
	UserName          string
	Password          string
	DB                string
	MaxConnection     string
	MaxIdleConnection string
	IsProduction      bool
}

type PGProvider struct {
	DB  *gorm.DB
	cnf PostgresConfig
}

func (receiver *PGProvider) GetRepresentation() string {
	return "PGProvider"
}

func (receiver *PGProvider) Recover(ctx context.Context) error {
	if err := receiver.IsAlive(ctx); err != nil {
		logrus.WithField("error", err.Error()).Errorln("try to reconnect...")
		pg, err := NewPostgresConnection(receiver.cnf)
		if err != nil {
			return err
		}
		receiver.DB = pg.DB
	}
	return nil
}

func (receiver *PGProvider) Run(ctx context.Context) error {
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
			logrus.Infoln(" Shouting PSQL...")
			if receiver.DB != nil {
				db, err := receiver.DB.DB()
				if err != nil {
					return err
				}
				return db.Close()
			}
			return nil
		}
	}
}

func (receiver *PGProvider) IsAlive(ctx context.Context) error {
	if receiver.DB == nil {
		return errors.New("pg is not alive")
	}
	sql, err := receiver.DB.WithContext(ctx).DB()
	if err != nil {
		return err
	}
	err = sql.PingContext(ctx)
	return err
}

func NewPostgresConnection(config PostgresConfig) (*PGProvider, error) {
	provider := PGProvider{cnf: config}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Tehran", provider.cnf.Host, provider.cnf.UserName, provider.cnf.Password, provider.cnf.DB, provider.cnf.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	provider.DB = db
	if err != nil {
		return &provider, err
	}
	sql, err := db.DB()
	if err != nil {
		return &provider, err
	}
	config.Password = "*****"
	logrus.WithField("config", fmt.Sprintf("%+v", config)).Infoln("connect to PG successfully...")
	sql.SetMaxOpenConns(cast.ToInt(provider.cnf.MaxConnection))
	sql.SetMaxIdleConns(cast.ToInt(provider.cnf.MaxIdleConnection))
	if config.IsProduction {
		provider.DB = db
	} else {
		provider.DB = db.Debug()
	}
	return &provider, nil
}
