package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/galiullindo/test-golang-14/calculator_server/internal/handlers"
	"github.com/galiullindo/test-golang-14/calculator_server/internal/libraries"
	"github.com/galiullindo/test-golang-14/calculator_server/internal/metrics"
	"github.com/spf13/pflag"
)

type arguments struct {
	Host     string
	Port     int
	CLib     string
	RustLib  string
	Interval time.Duration
}

func main() {
	filename, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get source directory: %s\n", err)
	}
	scriptDir := filepath.Dir(filename)

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Calculator HTTP server\n")
		pflag.PrintDefaults()
	}

	var args arguments
	pflag.StringVar(&args.Host, "host", "0.0.0.0", "")
	pflag.IntVar(&args.Port, "port", 8080, "")
	pflag.StringVar(
		&args.CLib,
		"c-lib",
		filepath.Join(scriptDir, "libcalculator.so"),
		"path to the compiled C shared library",
	)
	pflag.StringVar(
		&args.RustLib,
		"rust-lib",
		filepath.Join(scriptDir, "libcalculator_rust.so"),
		"path to the compiled Rust shared library",
	)
	pflag.DurationVar(
		&args.Interval,
		"interval",
		5*time.Second,
		"interval between periodic sum/sub reports (e.g. 100ms, 0.5s, 1s, 5s)",
	)
	pflag.Parse()

	cLibrary, rustLibrary, err := libraries.Load(args.CLib, args.RustLib)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load native libraries: %s\n", err)
		fmt.Fprintf(os.Stderr, "Did you run build.sh first?\n")
		os.Exit(1)
	}
	defer cLibrary.Close()
	defer rustLibrary.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	calcHandler := handlers.NewCalcHandler(cLibrary, rustLibrary)

	rpsCounter := metrics.NewRPSCounter(ctx)
	metricsHandler := handlers.NewMetricsHandler(rpsCounter)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /calc", calcHandler.Post)
	mux.HandleFunc("GET /metrics", metricsHandler.Get)

	server := &http.Server{
		Addr:    net.JoinHostPort(args.Host, strconv.Itoa(args.Port)),
		Handler: metricsHandler.MakeMiddleware(mux),
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		handlers.PeriodicPrinter(ctx, args.Interval)
	}()

	fmt.Printf("Calculator server listening on %s:%d\n", args.Host, args.Port)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Filed to start server: %s\n", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	<-sigChan

	fmt.Println("\nSIGINT received, shutting down...")
	handlers.PrintTotals("final")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "Filed to shutdown server: %s\n", err)
	}

	wg.Wait()
}
