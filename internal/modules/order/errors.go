package order

import (
	"errors"

	"tradeTornado/internal/lib"
)

var (
	OrderAlreadyProcessingFound = lib.NewErrorNotification()
	OrderAlreadyCreated         = lib.NewErrorNotification()
	NoOrderMatched              = lib.NewErrorNotification()
)

func init() {
	OrderAlreadyProcessingFound.Add("processing_order", errors.New("order is already processing"))
	OrderAlreadyCreated.Add("created_order", errors.New("order is already created"))
	NoOrderMatched.Add("no_match", errors.New("no order matched with this order"))
}
