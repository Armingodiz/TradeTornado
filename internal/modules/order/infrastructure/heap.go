package infrastructure

import (
	"context"
	"errors"
	"tradeTornado/internal/lib"
	"tradeTornado/internal/modules/order"
	"tradeTornado/internal/service/provider"
)

// TODO: if we need to persist we can use redis set forexample, I wasn't sure what we need from orderBook
type orderHeapData struct {
	*order.Order
}

func (ohd *orderHeapData) GetRank() int {
	return ohd.Price
}

type OrderBook struct {
	order.IOrderGenericRepository
	// TODO: add mutex lock for concurrency concerns
	minHeap lib.IMinHeap
	maxHeap lib.IMaxHeap
	session *provider.GormSession
}

func NewOrderBook(session *provider.GormSession, orderRepo order.IOrderGenericRepository) *OrderBook {
	return &OrderBook{
		minHeap:                 lib.NewMinHeap(),
		maxHeap:                 lib.NewMaxHeap(),
		session:                 session,
		IOrderGenericRepository: orderRepo,
	}
}

func (c *OrderBook) CreateWithHook(ctx context.Context, or *order.Order, process func(ctx context.Context, Order *order.Order) error) error {
	return c.session.RunTx(ctx, func() error {
		err := c.IOrderGenericRepository.CreateWithHook(ctx, or, process)
		if err != nil {
			return err
		}
		if or.Side == order.BuyOrderSide {
			c.maxHeap.Insert(&orderHeapData{Order: or})
		} else {
			c.minHeap.Insert(&orderHeapData{Order: or})
		}
		return nil
	})
}

func (c *OrderBook) GetMax(ctx context.Context) (*order.Order, error) {
	max := c.maxHeap.Max()
	maxOrder, ok := max.(*orderHeapData)
	if !ok {
		return nil, errors.New("invalid heap data")
	}
	return maxOrder.Order, nil
}

func (c *OrderBook) GetMin(ctx context.Context) (*order.Order, error) {
	min := c.minHeap.Min()
	minOrder, ok := min.(*orderHeapData)
	if !ok {
		return nil, errors.New("invalid heap data")
	}
	return minOrder.Order, nil
}
