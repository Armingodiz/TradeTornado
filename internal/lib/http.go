package lib

import "github.com/gin-gonic/gin"

// TODO: use local request and response instead of gin.HandlerFunc, dependency in lib!
type IController interface {
	GetRouters() []func() (method string, url string, handler gin.HandlerFunc)
	GetRoot() string
	GetMiddlewares() []gin.HandlerFunc
}
