package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
	"tradeTornado/internal/modules/order"
	"tradeTornado/internal/service/provider"

	"github.com/sirupsen/logrus"
)

type OrderEventHandler struct {
	createOrderConsumer provider.IConsumer
	matchOrderProducer  provider.IProducer
	matchOrderTopic     string
	orderRepositoryGen  func() order.IOrderWriteRepository
	processingOrders    sync.Map
}

type orderCreateEvent struct {
	OrderID  uint   `json:"orderID"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
	Side     string `json:"side"`
}

type orderMatchEvent struct {
	OrderID        uint      `json:"orderID"`
	MatchedOrderID uint      `json:"matchedOrderID"`
	CreatedAt      time.Time `json:"createdAt"`
}

func NewOrderEventHandler(createOrderConsumer provider.IConsumer, matchOrderProducer provider.IProducer, mot string, orderRepositoryGet func() order.IOrderWriteRepository) *OrderEventHandler {
	return &OrderEventHandler{createOrderConsumer: createOrderConsumer, matchOrderProducer: matchOrderProducer, matchOrderTopic: mot, orderRepositoryGen: orderRepositoryGet}
}

func (o *OrderEventHandler) Run(ctx context.Context) error {
	fmt.Println("### --> running")
	return o.createOrderConsumer.Consume(ctx, func(message string) error {
		var oe orderCreateEvent
		err := json.Unmarshal([]byte(message), &oe)
		if err != nil {
			return err
		}
		_, loaded := o.processingOrders.LoadOrStore(oe.OrderID, oe.OrderID)
		if loaded {
			logrus.Warningln(order.OrderAlreadyProcessingFound)
			return nil
		}
		defer o.processingOrders.Delete(oe.OrderID)
		om, err := order.NewOrder(oe.OrderID, oe.Side, oe.Price, oe.Quantity)
		if err != nil {
			logrus.Errorln(err)
			// Invalid orders are erased from queue
			return nil
		}
		orderRepo := o.orderRepositoryGen()
		err = orderRepo.CreateWithHook(ctx, om, func(ctx context.Context, createdOrder *order.Order) error {
			if err := o.matchOrder(ctx, orderRepo, createdOrder); err != nil {
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

func (o *OrderEventHandler) matchOrder(ctx context.Context, orderRepo order.IOrderWriteRepository, createdOrder *order.Order) error {
	// TODO: database may become bottleneck, check for better approach
	return orderRepo.SelectForUpdate(ctx, createdOrder.Side.GetMatchSide(), int(createdOrder.Price), int(createdOrder.Quantity), func(ctx context.Context, matchedOrder *order.Order) error {
		createdOrder.Match()
		err := orderRepo.Save(ctx, createdOrder)
		if err != nil {
			return err
		}
		matchedOrder.Match()
		err = orderRepo.Save(ctx, matchedOrder)
		if err != nil {
			return err
		}
		matchEvent := orderMatchEvent{OrderID: createdOrder.ID, MatchedOrderID: matchedOrder.ID, CreatedAt: time.Now()}
		bts, err := json.Marshal(matchEvent)
		if err != nil {
			return err
		}
		fmt.Println("--------------- MATCH EVENT:", string(bts))
		return o.matchOrderProducer.Produce(ctx, o.matchOrderTopic, string(bts))
	})
}

func (o *OrderEventHandler) GetRepresentation() string {
	return "OrderEventHandler"
}
