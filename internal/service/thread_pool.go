package service

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
)

type IExecutor interface {
	Run(ctx context.Context) error
	GetRepresentation() string
}

type ExecutorRegistry struct {
	executors []IExecutor
}

func NewExecutorRegistry() *ExecutorRegistry {
	return &ExecutorRegistry{executors: []IExecutor{}}
}

func (r *ExecutorRegistry) AddExecutor(e IExecutor) {
	r.executors = append(r.executors, e)
}

func (r *ExecutorRegistry) Run(ctx context.Context) {
	locker := sync.WaitGroup{}
	for _, exe := range r.executors {
		locker.Add(1)
		go func(execute IExecutor) {
			if err := execute.Run(ctx); err != nil {
				logrus.WithField("Executor", execute.GetRepresentation()).Errorln(err.Error())
			}
			locker.Done()
		}(exe)
	}
	locker.Wait()
}
