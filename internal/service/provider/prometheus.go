package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type PrometheusConfig struct {
	Port    string
	Disable bool
	Name    string
}

type PrometheusMetricsServer struct {
	*http.Server
	configs PrometheusConfig
}

func NewPrometheusService(configs PrometheusConfig) *PrometheusMetricsServer {
	server := &http.Server{
		Handler: promhttp.Handler(),
		Addr:    fmt.Sprintf(":%s", configs.Port),
	}
	return &PrometheusMetricsServer{Server: server, configs: configs}
}

func (ps *PrometheusMetricsServer) GetRepresentation() string {
	return "PrometheusMetricsServer"
}

func (ps *PrometheusMetricsServer) Run(ctx context.Context) error {
	if ps.configs.Disable {
		return errors.New("PrometheusMetricsServer is disable")
	}
	logrus.WithField("name", ps.configs.Name).WithField("config", fmt.Sprintf("%+v", ps.configs)).Infoln("Starting Prometheus Server...")
	errChannel := make(chan error)
	go func() {
		if err := ps.Server.ListenAndServe(); err != nil {
			errChannel <- err
		}
	}()
	select {
	case <-ctx.Done():
		logrus.WithField("name", ps.configs.Name).Infoln("Shouting Down Prometheus Server...")
		return ps.Server.Shutdown(ctx)
	case err := <-errChannel:
		return err
	}
}

// AddMetricCollectorGroup OverEngineering ? why just not call prometheus.MustRegister in packages metric.go
func (ps *PrometheusMetricsServer) AddMetricCollectorGroup(collectorGroup ...prometheus.Collector) {
	prometheus.MustRegister(collectorGroup...)
}
