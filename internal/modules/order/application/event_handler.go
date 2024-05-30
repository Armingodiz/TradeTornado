package application

import (
	"context"
	"encoding/json"
	"fmt"
	"tradeTornado/internal/modules/order"
)

type OrderEventHandler struct {
	consumerProvider order.IOrderConsumer
	orderRepository  order.IOrderWriteRepository
}

type orderEvent struct {
	OrderID  uint   `json:"orderID"`
	Price    int64  `json:"price"`
	Quantity int32  `json:"quantity"`
	Side     string `json:"side"`
}

func NewOrderEventHandler(consumerProvider order.IOrderConsumer, orderRepository order.IOrderGenericRepository) *OrderEventHandler {
	return &OrderEventHandler{consumerProvider: consumerProvider, orderRepository: orderRepository}
}

func (o *OrderEventHandler) Run(ctx context.Context) error {
	fmt.Println("### --> running")
	return o.consumerProvider.Consume(ctx, func(message string) error {
		fmt.Println("##### OrderEventHandler Received message: ", message)
		var oe orderEvent
		err := json.Unmarshal([]byte(message), &oe)
		if err != nil {
			return err
		}
		return nil
	})
}

func (o *OrderEventHandler) GetRepresentation() string {
	return "OrderEventHandler"
}

func (o *OrderEventHandler) createOrder(ctx context.Context, oe orderEvent) (uint, error) {
	cat, err := order.NewOrder(oe.OrderID, oe.Side, oe.Price, oe.Quantity)
	if err != nil {
		return 0, err
	}
	err = o.orderRepository.Save(ctx, cat)
	if err != nil {
		return 0, err
	}
	return cat.ID, nil
}
