package infrastructure

import (
	"tradeTornado/internal/modules/order/application"
)

type OrderController struct {
	queryHanlder *application.OrderQueryHandler
}

func NewOrderController(qh *application.OrderQueryHandler) *OrderController {
	return &OrderController{
		queryHanlder: qh,
	}
}
