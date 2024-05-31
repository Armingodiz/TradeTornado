package lib

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func GetEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func Terminable() context.Context {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGINT)
	ctx, canFu := context.WithCancel(context.Background())
	go func() {
		<-c
		canFu()
	}()
	return ctx
}
