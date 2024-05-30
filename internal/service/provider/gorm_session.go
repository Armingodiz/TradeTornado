package provider

import (
	"context"
	"errors"
	"sync"

	"tradeTornado/internal/lib/messaging"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var SessionNotInTransactionError = errors.New("session not in transaction")
var SessionAlreadyInTransactionError = errors.New("session already in transaction")

func NewGormSession(db *gorm.DB, bus messaging.IEventBus) *GormSession {
	return &GormSession{
		database: db,
		lock:     sync.Mutex{},
		events:   []messaging.IEvent{},
		eventBus: bus,
	}
}

type GormSession struct {
	database *gorm.DB
	tx       *gorm.DB
	events   []messaging.IEvent
	lock     sync.Mutex
	eventBus messaging.IEventBus
}

func (s *GormSession) publishEvents(ctx context.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(s.events) > 0 {
		for _, event := range s.events {
			if err := s.eventBus.Publish(ctx, event); err != nil {
				return err
			}
		}
		s.events = []messaging.IEvent{}
	}
	return nil
}
func (s *GormSession) AddTXEvent(ctx context.Context, events ...messaging.IEvent) error {
	logrus.Debugln("AddTXEvent ", len(events))
	for _, event := range events {
		bt, _ := event.Marshal()
		logrus.Debugln("AddTXEvent", event.GetName(), string(bt))
	}
	s.lock.Lock()
	s.events = append(s.events, events...)
	s.lock.Unlock()
	if !s.InTransaction() {
		logrus.Debugln("###########Publishing events")
		return s.publishEvents(ctx)
	}
	return nil
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

	if err := s.publishEvents(ctx); err != nil {
		return err
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
