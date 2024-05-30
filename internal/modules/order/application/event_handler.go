package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
	"tradeTornado/internal/modules/order"

	"github.com/sirupsen/logrus"
)

type OrderEventHandler struct {
	eventBus         order.IEventBus
	orderRepository  order.IOrderWriteRepository
	processingOrders sync.Map
}

type orderCreateEvent struct {
	OrderID  uint   `json:"orderID"`
	Price    int64  `json:"price"`
	Quantity int32  `json:"quantity"`
	Side     string `json:"side"`
}

type orderMatchEvent struct {
	OrderID        uint      `json:"orderID"`
	MatchedOrderID uint      `json:"price"`
	CreatedAt      time.Time `json:"createdAt"`
}

func NewOrderEventHandler(eventBus order.IEventBus, orderRepository order.IOrderGenericRepository) *OrderEventHandler {
	return &OrderEventHandler{eventBus: eventBus, orderRepository: orderRepository}
}

func (o *OrderEventHandler) Run(ctx context.Context) error {
	fmt.Println("### --> running")
	return o.eventBus.Consume(ctx, func(message string) error {
		fmt.Println("##### OrderEventHandler Received message: ", message)
		var oe orderCreateEvent
		err := json.Unmarshal([]byte(message), &oe)
		if err != nil {
			return err
		}
		_, loaded := o.processingOrders.LoadOrStore(oe.OrderID, oe)
		if loaded {
			logrus.Warningln(order.OrderAlreadyProcessingFound)
			return nil
		}
		om, err := order.NewOrder(oe.OrderID, oe.Side, oe.Price, oe.Quantity)
		if err != nil {
			logrus.Errorln(err)
			// Invalid orders are erased from queue
			return nil
		}
		err = o.orderRepository.CreateWithHook(ctx, om, func(ctx context.Context, createdOrder *order.Order) error {
			if err := o.matchOrder(ctx, createdOrder); err != nil {
				if !errors.Is(err, order.NoOrderMatched) {
					return err
				}
			}
			return nil
		})
		if err != nil {
			if errors.Is(err, order.OrderAlreadyCreated) {
				logrus.Warningln(order.OrderAlreadyCreated)
				return nil
			} else {
				return err
			}
		}
		return nil
	})
}

func (o *OrderEventHandler) matchOrder(ctx context.Context, createdOrder *order.Order) error {
	// TODO: database may become bottleneck, check for better approach
	return o.orderRepository.SelectForUpdate(ctx, createdOrder.Side, int(createdOrder.Price), int(createdOrder.Quantity), func(ctx context.Context, matchedOrder *order.Order) error {
		createdOrder.Match()
		err := o.orderRepository.Save(ctx, createdOrder)
		if err != nil {
			return err
		}
		matchedOrder.Match()
		err = o.orderRepository.Save(ctx, matchedOrder)
		if err != nil {
			return err
		}
		matchEvent := orderMatchEvent{OrderID: createdOrder.ID, MatchedOrderID: matchedOrder.ID, CreatedAt: time.Now()}
		bts, err := json.Marshal(matchEvent)
		if err != nil {
			return err
		}
		// TODO: fix topic and message
		return o.eventBus.Produce("", "", string(bts))
	})
}

func (o *OrderEventHandler) GetRepresentation() string {
	return "OrderEventHandler"
}
