package application

import (
	"context"
	"tradeTornado/internal/lib"
	"tradeTornado/internal/modules/order"
)

type OrderDto struct {
	CreatedAt int64
}

type OrderQueryHandler struct {
	orderRepository order.IOrderReadRepository
}

func NewOrderQueryHandler(orderRepository order.IOrderReadRepository) *OrderQueryHandler {
	return &OrderQueryHandler{orderRepository: orderRepository}
}

func (cqh *OrderQueryHandler) ListOrders(ctx context.Context, criteria lib.Criteria) ([]*OrderDto, int, error) {
	cats, total, err := cqh.orderRepository.List(ctx, criteria)
	if err != nil {
		return nil, 0, err
	}
	var dtos []*OrderDto
	for _, cat := range cats {
		dtos = append(dtos, cqh.toDtos(cat)[0])
	}
	return dtos, total, nil
}

func (cqh *OrderQueryHandler) toDtos(Order ...*order.Order) []*OrderDto {
	dtos := make([]*OrderDto, 0)
	return dtos
}
