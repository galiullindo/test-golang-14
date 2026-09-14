package handlers

import (
	"fmt"
	"net/http"

	"github.com/galiullindo/test-golang-14/calculator_server/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsHandler struct {
	rpsCounter *metrics.RPSCounter
	reg        prometheus.Registry
}

func NewMetricsHandler(rpsCounte *metrics.RPSCounter) *MetricsHandler {
	return &MetricsHandler{rpsCounter: rpsCounte}

}

func (h *MetricsHandler) Get(w http.ResponseWriter, r *http.Request) {
	promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)

	history := h.rpsCounter.History()

	fmt.Fprintln(
		w, "\n# HELP http_requests_rps_last_minute RPS per second for the last 60 seconds.",
	)
	fmt.Fprintln(w, "# TYPE http_requests_rps_last_minute gauge")

	for i, rpsValue := range history {
		secondsAgo := 59 - i
		fmt.Fprintf(
			w, "http_requests_rps_last_minute{back_seconds=\"%d\"} %d\n", secondsAgo, rpsValue,
		)
	}
}

func (h *MetricsHandler) MakeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metrics" {
			h.rpsCounter.Increment()
		}
		next.ServeHTTP(w, r)
	})
}
