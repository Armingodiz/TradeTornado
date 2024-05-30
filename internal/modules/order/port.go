package order

import (
	"context"
	"tradeTornado/internal/lib"
)

type IOrderWriteRepository interface {
	SaveWithHook(ctx context.Context, order *Order, process func(ctx context.Context, Order *Order) error) error
	Save(ctx context.Context, order *Order) error
	SelectForUpdate(ctx context.Context, price, quantity int, updateFn func(ctx context.Context, Order *Order) error) error
}

type IOrderReadRepository interface {
	List(ctx context.Context, cr lib.Criteria) ([]*Order, int, error)
}

type IOrderGenericRepository interface {
	IOrderReadRepository
	IOrderWriteRepository
}

type IOrderCacheRepository interface{}

type IOrderConsumer interface {
	Consume(ctx context.Context, process func(message string) error) error
}
