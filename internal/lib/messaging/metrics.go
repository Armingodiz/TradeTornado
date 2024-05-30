package messaging

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricEventBusErrorCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "event_bus_error_count",
		Help: "Total number of event bus errors",
	}, []string{"provider", "event_name"})
	metricEventBusSuccessCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "event_bus_success_count",
		Help: "Total number of event bus success",
	}, []string{"provider"})
	metricEventBusEventCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "event_bus_event_current_count",
		Help: "Total number of event bus events",
	}, []string{"provider"})
)

func GetMetrics() []prometheus.Collector {
	return []prometheus.Collector{
		metricEventBusErrorCount,
		metricEventBusSuccessCount,
		metricEventBusEventCount,
	}
}
