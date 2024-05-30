package order

import (
	"errors"

	"tradeTornado/internal/lib"
)

var (
	OrderNotFound       = lib.NewErrorNotification()
	ParentOrderNotFound = lib.NewErrorNotification()
	OrderDuplicateName  = lib.NewErrorNotification()
	OrderDuplicateRefId = lib.NewErrorNotification()
	NotFound            = lib.NewNotFoundError("Order")
)

func init() {
	OrderNotFound.Add("Order_not_found", errors.New("Order not found"))
	ParentOrderNotFound.Add("parent_Order_not_found", errors.New("parent Order not found"))
	OrderDuplicateName.Add("Order_name", errors.New("Order with this name already exist"))
	OrderDuplicateRefId.Add("Order_ref_id", errors.New("Order with this ref id already exist"))
}
