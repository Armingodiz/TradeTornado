package messaging

import (
	"context"
)

type IEvent interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	GetName() string
}

type AckFunc func(ctx context.Context, ack bool) error

type ISubscriber interface {
	Subscribe(ctx context.Context, event IEvent) error
	Representation() string
}

type IPublisher interface {
	Publish(ctx context.Context, event IEvent) error
}

type IDispatcher interface {
	Dispatch(ctx context.Context, event IEvent, ack AckFunc) error
}

type IEventBus interface { // High level function for event dispatcher and related issue about it
	IPublisher
	IDispatcher
	RegisterAtLeastOnce(subscriber ISubscriber)
	RegisterAtMostOnce(subscriber ISubscriber)
}

type IMessageReceiver interface {
	Receive(ctx context.Context, msg []byte, ackFunc AckFunc) error
}

type IMessagePublisher interface {
	Publish(ctx context.Context, event IEvent) error
}

type IMarshaller interface {
	IMessagePublisher
	IMessageReceiver
}

type IQueuePublisher interface {
	Send(ctx context.Context, msg []byte) error
}
