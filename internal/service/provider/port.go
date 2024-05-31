package provider

import "context"

type IConsumer interface {
	Consume(ctx context.Context, process func(message string) error) error
}

type IProducer interface {
	Produce(ctx context.Context, topic, message string) error
	ProduceWithKey(ctx context.Context, topic, key, message string) error
}

type IEventBus interface {
	IConsumer
	IProducer
}
