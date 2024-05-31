package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"tradeTornado/internal/lib"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

type ServerConfigs struct {
	Port           string
	Name           string
	ReadTimeoutMS  string
	WriteTimeoutMS string
}

const ErrorsHeaderName = "errors"

type GinServer struct {
	cnf    ServerConfigs
	Router *gin.Engine
}

func NewGinServer(cnf ServerConfigs) *GinServer {
	ginServer := gin.Default()
	server := &GinServer{cnf: cnf}
	server.Router = ginServer
	return server
}

func (s *GinServer) GetRepresentation() string {
	return "GIN_API"
}

func (s *GinServer) Run(ctx context.Context) error {
	server := &http.Server{
		Handler:      s.Router,
		Addr:         fmt.Sprintf(":%s", s.cnf.Port),
		ReadTimeout:  time.Duration(cast.ToInt(s.cnf.ReadTimeoutMS)) * time.Millisecond,
		WriteTimeout: time.Duration(cast.ToInt(s.cnf.WriteTimeoutMS)) * time.Millisecond,
	}

	logrus.WithField("name", s.cnf.Name).WithField("config", fmt.Sprintf("%+v", s.cnf)).Infoln("Starting Http Server...")
	errChannel := make(chan error, 0)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			errChannel <- err
		}
	}()
	select {
	case <-ctx.Done():
		logrus.WithField("name", s.cnf.Name).Infoln("Shouting Down HttpServer...")
		return server.Shutdown(ctx)
	case err := <-errChannel:
		return err
	}
}

func (s *GinServer) AddRouter(router lib.IController) {
	gr2 := s.Router.Group(router.GetRoot())
	gr2.Use(genericErrorHandler)
	for _, handlerFunc := range router.GetMiddlewares() {
		gr2.Use(handlerFunc)
	}
	for _, handler := range router.GetRouters() {
		gr2.Handle(handler())
	}
}

func (s *GinServer) AddMiddleWare(middleware gin.HandlerFunc) {
	s.Router.Use(middleware)
}

func genericErrorHandler(context *gin.Context) {
	context.Next()
	if !context.IsAborted() {
		if lastErr := context.Errors.Last(); lastErr != nil {
			errNotification, ok := lastErr.Err.(*lib.ErrorNotification)
			if ok {
				context.AbortWithStatusJSON(http.StatusBadRequest, errNotification.Errs)
				return
			}
			context.AbortWithStatusJSON(http.StatusInternalServerError, lastErr)
			return
		}
	}
}
