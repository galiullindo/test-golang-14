package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/galiullindo/test-golang-14/calculator_server/internal/libraries"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	mu       sync.Mutex
	sumValue int64
	subValue int64
)

var (
	reg = prometheus.NewRegistry()

	cAddFuncDuration = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "c_add_function_duration_seconds",
		Help:       "Duration of C add function execution in seconds (p95, p99)",
		Objectives: map[float64]float64{0.95: 0.005, 0.99: 0.001}, // квантили и допустимая погрешность
	})

	rustSubFuncDuration = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "rust_sub_function_duration_seconds",
		Help:       "Duration of Rust sub function execution in seconds (p95, p99)",
		Objectives: map[float64]float64{0.95: 0.005, 0.99: 0.001},
	})
)

func init() {
	reg.MustRegister(cAddFuncDuration)
	reg.MustRegister(rustSubFuncDuration)
}

type CalcHandler struct {
	cLibrary    *libraries.CLibrary
	rustLibrary *libraries.RustLibrary
}

func NewCalcHandler(
	cLibrary *libraries.CLibrary,
	rustLibrary *libraries.RustLibrary,
) *CalcHandler {
	return &CalcHandler{cLibrary: cLibrary, rustLibrary: rustLibrary}
}

func (h *CalcHandler) Post(w http.ResponseWriter, r *http.Request) {
	strNum := r.URL.Query().Get("num")
	if strNum == "" {
		respond(w, http.StatusBadRequest, []byte("missing 'num' query parameter"))
		return
	}

	num, err := strconv.ParseInt(strNum, 10, 64)
	if err != nil {
		respond(w, http.StatusBadRequest, []byte("'num' must be an integer"))
		return
	}

	mu.Lock()
	defer mu.Unlock()

	startC := time.Now()
	sumValue = h.cLibrary.Add(sumValue, num)
	cAddFuncDuration.Observe(time.Since(startC).Seconds())

	startRust := time.Now()
	subValue = h.rustLibrary.Sub(subValue, num)
	rustSubFuncDuration.Observe(time.Since(startRust).Seconds())

	respond(w, http.StatusOK, []byte("ok"))
}

func respond(w http.ResponseWriter, code int, body []byte) {
	w.WriteHeader(code)
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	if len(body) != 0 {
		w.Write(body)
	}
}

func PrintTotals(label string) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Printf("[%s] sum=%d sub=%d\n", label, sumValue, subValue)
}

func PeriodicPrinter(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			PrintTotals("periodic")
		case <-ctx.Done():
			return
		}
	}
}
