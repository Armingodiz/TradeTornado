package infrastructure

import (
	"context"

	"tradeTornado/internal/lib"
	"tradeTornado/internal/modules/order"
	"tradeTornado/internal/service/provider"
)

type OrderRepository struct {
	session *provider.GormSession
}

func NewOrderRepository(session *provider.GormSession) *OrderRepository {
	return &OrderRepository{
		session: session,
	}
}

func (c *OrderRepository) Save(ctx context.Context, cg *order.Order) error {
	return c.session.Gorm().WithContext(ctx).Save(cg).Error
}

func (c *OrderRepository) SaveWithHook(ctx context.Context, order *order.Order, process func(ctx context.Context, Order *order.Order) error) error {
	return c.session.RunTx(ctx, func() error {
		err := c.Save(ctx, order)
		if err != nil {
			return err
		}
		return process(ctx, order)
	})
}

// TODO: implement this to use befor redis
func (c *OrderRepository) SelectForUpdate(ctx context.Context, price, quantity int, updateFn func(ctx context.Context, order *order.Order) error) error {
	return nil
}

func (c *OrderRepository) List(ctx context.Context, cr lib.Criteria) ([]*order.Order, int, error) {
	var orders []*order.Order
	query, err := lib.GenericApplyGormCriteria(c.session.Gorm().WithContext(ctx), order.Order{}, &cr)
	if err != nil {
		return nil, 0, err
	}
	err = query.Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}
	cr.Pagination = nil
	var total int64
	countQ, err := lib.GenericApplyGormCriteria(c.session.Gorm().WithContext(ctx), order.Order{}, &cr)
	if err != nil {
		return nil, 0, err
	}
	if err := countQ.Model(order.Order{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	return orders, int(total), nil
}

func (c *OrderRepository) Migrate(ctx context.Context) error {
	return c.session.Gorm().WithContext(ctx).AutoMigrate(&order.Order{})
}
