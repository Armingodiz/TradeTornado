package wiring

import (
	"tradeTornado/internal/service/provider"
)

func (c *ContainerBuilder) GetMetricsService() *provider.PrometheusMetricsServer {
	if c.prometheusService == nil {
		c.prometheusService = provider.NewPrometheusService(*c.cnf.MetricConfig)
	}
	return c.prometheusService
}

func (c *ContainerBuilder) initMetrics() {
}
