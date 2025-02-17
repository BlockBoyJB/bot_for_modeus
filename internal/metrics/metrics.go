package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

const (
	namespace = "bot_for_modeus"
	subsystem = "bot"
)

var (
	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "request_duration",
		Buckets:   []float64{.01, .1, .5, 1, 3, 5, 10},
	}, []string{"type"})

	requestTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "request_total",
	}, []string{"type"})

	errorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "errors_total",
	}, []string{"type"})
)

func RequestDuration(t string, d time.Duration) {
	requestDuration.WithLabelValues(t).Observe(d.Seconds())
}

func RequestTotal(t string) {
	requestTotal.WithLabelValues(t).Inc()
}

func ErrorsTotal(t string) {
	errorsTotal.WithLabelValues(t).Inc()
}

func Listen(addr string) error {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, mux)
}
