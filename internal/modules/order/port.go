package order

import (
	"context"
	"tradeTornado/internal/lib"
)

type IOrderWriteRepository interface {
	SaveWithHook(ctx context.Context, order *Order, process func(ctx context.Context, Order *Order) error) error
	CreateWithHook(ctx context.Context, order *Order, process func(ctx context.Context, Order *Order) error) error
	SelectForUpdate(ctx context.Context, side OrderSide, price, quantity int, updateFn func(ctx context.Context, Order *Order) error) error
	Save(ctx context.Context, cg *Order) error
}

type IOrderReadRepository interface {
	List(ctx context.Context, cr lib.Criteria) ([]*Order, int, error)
}

type IOrderGenericRepository interface {
	IOrderReadRepository
	IOrderWriteRepository
}

type IOrderCacheRepository interface{}

type IConsumer interface {
	Consume(ctx context.Context, process func(message string) error) error
}

type IProducer interface {
	Produce(topic, key, message string) error
}

type IEventBus interface {
	IConsumer
	IProducer
}
