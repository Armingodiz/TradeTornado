package application

import (
	"context"
	"tradeTornado/internal/lib"
	"tradeTornado/internal/modules/order"
)

type OrderDto struct {
	Matched   bool
	Side      string
	Price     int64
	Quantity  int32
	CreatedAt int64
}

type OrderQueryHandler struct {
	orderRepository order.IOrderReadRepository
}

func NewOrderQueryHandler(orderRepository order.IOrderReadRepository) *OrderQueryHandler {
	return &OrderQueryHandler{orderRepository: orderRepository}
}

func (cqh *OrderQueryHandler) ListOrders(ctx context.Context, criteria lib.Criteria) ([]*OrderDto, int, error) {
	orders, total, err := cqh.orderRepository.List(ctx, criteria)
	if err != nil {
		return nil, 0, err
	}
	return cqh.toDtos(orders...), total, nil
}

func (cqh *OrderQueryHandler) toDtos(orders ...*order.Order) []*OrderDto {
	dtos := make([]*OrderDto, 0)
	for _, ord := range orders {
		dtos = append(dtos, &OrderDto{
			Price:     ord.Price,
			Quantity:  ord.Quantity,
			Matched:   ord.Matched,
			Side:      string(ord.Side),
			CreatedAt: ord.CreatedAt.Unix(),
		})
	}
	return dtos
}
