package provider

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func NewGormSession(db *gorm.DB) *GormSession {
	return &GormSession{
		database: db,
		lock:     sync.Mutex{},
	}
}

type GormSession struct {
	database *gorm.DB
	tx       *gorm.DB
	lock     sync.Mutex
}

func (s *GormSession) RunTx(ctx context.Context, closure func() error) error {
	localTx := !s.InTransaction()
	if localTx {
		if err := s.Begin(ctx); err != nil {
			return err
		}
	}
	success := false

	defer func() {
		if !success && localTx {
			if err := s.Rollback(); err != nil {
				logrus.Error(err)
			}
		}
	}()

	err := closure()
	if err == nil {
		if localTx {
			err := s.Commit(ctx)
			if err != nil {
				return err
			}
			success = true
		}
	}

	return err
}

func (s *GormSession) Gorm() *gorm.DB {
	if s.tx != nil {
		return s.tx
	}
	return s.database
}
func (s *GormSession) Begin(ctx context.Context) error {
	if s.tx != nil {
		return SessionAlreadyInTransactionError
	}
	s.tx = s.database.WithContext(ctx).Begin()

	return s.tx.Error
}

func (s *GormSession) Commit(ctx context.Context) error {
	if s.tx == nil {
		return SessionNotInTransactionError
	}
	if err := s.tx.WithContext(ctx).Commit().Error; err != nil {
		return err
	}
	logrus.Debugln("Item has been committed")
	s.tx = nil
	return nil
}

func (s *GormSession) Rollback() error {
	if s.tx == nil {
		return nil
	}

	if err := s.tx.Rollback().Error; err != nil {
		return err
	}

	s.tx = nil
	return nil
}

func (s *GormSession) InTransaction() bool {
	return s.tx != nil
}
