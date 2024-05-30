package lib

import (
	"errors"
	"fmt"
	"sync"
)

type NotFound struct {
	entity string
}

func (n NotFound) Error() string {
	return fmt.Sprintf("%s not found", n.entity)
}

func (n NotFound) Is(err error) bool {
	var notFound NotFound
	ok := errors.As(err, &notFound)
	return ok
}

func NewNotFoundError(entity string) error {
	return NotFound{entity: entity}
}

// ErrorNotification https://www.martinfowler.com/eaaDev/Notification.html
type ErrorNotification struct {
	Errs map[string]error
	lock sync.Mutex
}

func (n *ErrorNotification) Error() string {
	var errs []error
	for _, err := range n.Errs {
		errs = append(errs, err)
	}
	return errors.Join(errs...).Error()
}

func (n *ErrorNotification) Is(err error) bool {
	_, ok := err.(*ErrorNotification)
	return ok
}

func NewErrorNotification() *ErrorNotification {
	return &ErrorNotification{lock: sync.Mutex{}, Errs: make(map[string]error)}
}

func (n *ErrorNotification) Add(key string, err error) {
	n.lock.Lock()
	n.Errs[key] = err
	n.lock.Unlock()
}

func (n *ErrorNotification) Err() error {
	if len(n.Errs) == 0 {
		return nil
	}
	return n
}

func (n *ErrorNotification) UintShouldBeEqual(id string, v1, v2 uint) {
	if v1 != v2 {
		n.Add(id, fmt.Errorf("should be equal to %d", v2))
	}
}

func (n *ErrorNotification) BoolShouldBeEqual(id string, v1, v2 bool) {
	if v1 != v2 {
		n.Add(id, fmt.Errorf("should be equal to %v", v2))
	}
}

func (n *ErrorNotification) StringNotEmpty(id string, value string) {
	if len(value) < 1 {
		n.Add(id, errors.New("should not be empty"))
	}
}

func (n *ErrorNotification) UintShouldBeGT(id string, v1, v2 uint) {
	if v1 <= v2 {
		n.Add(id, fmt.Errorf("should be grater than %d", v2))
	}
}

func (n *ErrorNotification) UintShouldBeGTE(id string, v1, v2 uint) {
	if v1 < v2 {
		n.Add(id, fmt.Errorf("should be grater than or equal to %d", v2))
	}
}

func (n *ErrorNotification) NotNil(id string, v1 any) {
	if v1 == nil {
		n.Add(id, errors.New("should not be nil"))
	}
}
