package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsObserver struct {
	rpsCounter          *RPSCounter
	reg                 *prometheus.Registry
	cAddFuncDuration    prometheus.Summary
	rustSubFuncDuration prometheus.Summary
}

func NewMetricsObserver(ctx context.Context) *MetricsObserver {
	reg := prometheus.NewRegistry()

	cAddFuncDuration := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "c_add_function_duration_seconds",
		Help:       "Duration of C add function execution in seconds (p95, p99)",
		Objectives: map[float64]float64{0.95: 0.005, 0.99: 0.001},
	})

	rustSubFuncDuration := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "rust_sub_function_duration_seconds",
		Help:       "Duration of Rust sub function execution in seconds (p95, p99)",
		Objectives: map[float64]float64{0.95: 0.005, 0.99: 0.001},
	})

	reg.MustRegister(cAddFuncDuration)
	reg.MustRegister(rustSubFuncDuration)

	rpsCounter := NewRPSCounter(ctx)

	o := &MetricsObserver{
		reg:                 reg,
		rpsCounter:          rpsCounter,
		cAddFuncDuration:    cAddFuncDuration,
		rustSubFuncDuration: rustSubFuncDuration,
	}

	return o
}

func (o *MetricsObserver) ObserveCAdd(d float64) {
	o.cAddFuncDuration.Observe(d)
}
func (o *MetricsObserver) ObserveRustSub(d float64) {
	o.rustSubFuncDuration.Observe(d)
}

func (o *MetricsObserver) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metrics" {
			o.rpsCounter.Increment()
		}
		next.ServeHTTP(w, r)
	})
}

func (o *MetricsObserver) Handler(w http.ResponseWriter, r *http.Request) {
	promhttp.HandlerFor(o.reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)

	history := o.rpsCounter.History()

	fmt.Fprintln(
		w,
		"\n# HELP http_requests_rps_last_minute RPS per second for the last 60 seconds.",
	)
	fmt.Fprintln(w, "# TYPE http_requests_rps_last_minute gauge")

	for i, rpsValue := range history {
		secondsAgo := 59 - i
		fmt.Fprintf(
			w, "http_requests_rps_last_minute{back_seconds=\"%d\"} %d\n",
			secondsAgo,
			rpsValue,
		)
	}
}

type RPSCounter struct {
	mu      sync.Mutex
	buckets [60]uint64
	last    time.Time
}

func NewRPSCounter(ctx context.Context) *RPSCounter {
	return &RPSCounter{last: time.Now()}
}

func (c *RPSCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.сleanup(now)

	idx := now.Unix() % 60
	c.buckets[idx]++
}

func (c *RPSCounter) History() [60]uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.сleanup(now)

	var history [60]uint64
	currentIdx := now.Unix() % 60
	for i := 0; i < 60; i++ {
		idx := (currentIdx + int64(i) + 1) % 60
		history[i] = c.buckets[idx]
	}

	return history
}

func (c *RPSCounter) сleanup(now time.Time) {
	current := now.Unix()
	last := c.last.Unix()

	if current <= last {
		return
	}
	if current-last >= 60 {
		c.buckets = [60]uint64{}
		c.last = now
		return
	}

	for t := last + 1; t <= current; t++ {
		idx := t % 60
		c.buckets[idx] = 0
	}
	c.last = now
}
