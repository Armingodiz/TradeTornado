package infrastructure

import (
	"context"
	"errors"

	"tradeTornado/internal/lib"
	"tradeTornado/internal/modules/order"
	"tradeTornado/internal/service/provider"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (c *OrderRepository) CreateWithHook(ctx context.Context, order *order.Order, process func(ctx context.Context, Order *order.Order) error) error {
	return c.session.RunTx(ctx, func() error {
		err := c.session.Gorm().WithContext(ctx).Create(order).Error
		if err != nil {
			return err
		}
		return process(ctx, order)
	})
}

func (c *OrderRepository) SaveWithHook(ctx context.Context, order *order.Order, process func(ctx context.Context, Order *order.Order) error) error {
	return c.session.RunTx(ctx, func() error {
		err := c.session.Gorm().WithContext(ctx).Save(order).Error
		if err != nil {
			return err
		}
		return process(ctx, order)
	})
}

func (c *OrderRepository) SelectForUpdate(ctx context.Context, side order.OrderSide, price, quantity int, updateFn func(context.Context, *order.Order) error) error {
	return c.session.RunTx(ctx, func() error {
		var matchedOrder *order.Order
		if err := c.session.Gorm().
			WithContext(ctx).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("matched = ? and side = ? and price = ? and quantity = ?", false, side, price, quantity).
			First(&matchedOrder).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return order.NoOrderMatched
			}
			return err
		}
		return updateFn(ctx, matchedOrder)
	})
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
