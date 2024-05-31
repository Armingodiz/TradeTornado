package infrastructure

import (
	"net/http"
	"tradeTornado/internal/lib"
	"tradeTornado/internal/modules/order/application"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	queryHanlder *application.OrderQueryHandler
}

func NewOrderController(qh *application.OrderQueryHandler) *OrderController {
	return &OrderController{
		queryHanlder: qh,
	}
}

func (oc *OrderController) GetRouters() []func() (method string, url string, handler gin.HandlerFunc) {
	return []func() (method string, url string, handler gin.HandlerFunc){
		oc.listOrderBook,
	}
}
func (oc *OrderController) GetRoot() string {
	return "orderbook"
}
func (oc *OrderController) GetMiddlewares() []gin.HandlerFunc {
	return nil
}

func (oc *OrderController) listOrderBook() (method string, uri string, handler gin.HandlerFunc) {
	return http.MethodGet, "", func(context *gin.Context) {
		criteria := lib.NewCriteria()
		orders, count, err := oc.queryHanlder.ListOrders(context, *criteria)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		context.JSON(http.StatusOK, gin.H{
			"total":  count,
			"orders": orders,
		})
	}
}
