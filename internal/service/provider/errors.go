package provider

import (
	"errors"

	"tradeTornado/internal/lib"
)

var (
	SessionNotInTransactionError     = lib.NewErrorNotification()
	SessionAlreadyInTransactionError = lib.NewErrorNotification()
)

func init() {
	SessionNotInTransactionError.Add("processing_order", errors.New("session not in transaction"))
	SessionAlreadyInTransactionError.Add("created_order", errors.New("session already in transaction"))
}
