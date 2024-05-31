package order

import (
	"time"
	"tradeTornado/internal/lib"
)

type Order struct {
	ID        uint `gorm:"primarykey;column:id"`
	CreatedAt time.Time
	Matched   bool      `criteria:"matched" gorm:"column:matched;index:idx_matched_side_price_quantity,priority:1"`
	Side      OrderSide `criteria:"side" gorm:"column:side;index:idx_matched_side_price_quantity,priority:2"`
	Price     int64     `criteria:"price" gorm:"column:price;index:idx_matched_side_price_quantity,priority:3"`
	Quantity  int32     `criteria:"quantity" gorm:"column:quantity;index:idx_matched_side_price_quantity,priority:4"`
}

type OrderSide string

const (
	SellOrderSide OrderSide = "sell"
	BuyOrderSide  OrderSide = "buy"
)

func (os OrderSide) GetMatchSide() OrderSide {
	if os == SellOrderSide {
		return BuyOrderSide
	} else {
		return SellOrderSide
	}
}

func NewOrder(id uint, side string, price int64, quantity int32) (*Order, error) {
	order := &Order{
		Price:     price,
		Quantity:  quantity,
		ID:        id,
		Side:      OrderSide(side),
		CreatedAt: time.Now(),
	}
	return order, order.validate()
}

func (order *Order) Match() {
	order.Matched = true
}

func (order *Order) validate() error {
	validation := lib.NewErrorNotification()

	validation.UintShouldBeGT("price", uint(order.Price), 0)
	validation.UintShouldBeGT("quantity", uint(order.Quantity), 0)

	return validation.Err()
}
