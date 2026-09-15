package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/galiullindo/test-golang-14/calculator_server/internal/handler"
	"github.com/galiullindo/test-golang-14/calculator_server/internal/middleware"
	"github.com/galiullindo/test-golang-14/calculator_server/internal/pkg"
	"github.com/galiullindo/test-golang-14/calculator_server/internal/repository"
	"github.com/galiullindo/test-golang-14/calculator_server/internal/service"
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

	cLib, rustLib, err := pkg.LoadLibraries(args.CLib, args.RustLib)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load native libraries: %s\n", err)
		fmt.Fprintf(os.Stderr, "Did you run build.sh first?\n")
		os.Exit(1)
	}
	defer cLib.Close()
	defer rustLib.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := repository.NewCalcRepository()
	metrics := middleware.NewMetricsObserver(ctx)
	serv := service.NewService(ctx, args.Interval, repo, metrics, cLib, rustLib)
	handl := handler.NewHandler(serv)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /metrics", metrics.Handler)
	handl.RegisterRoutes(mux)

	server := http.Server{
		Addr:         net.JoinHostPort(args.Host, strconv.Itoa(args.Port)),
		Handler:      metrics.Middleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf(
		"Calculator server listening on %s\n",
		net.JoinHostPort(args.Host, strconv.Itoa(args.Port)),
	)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			fmt.Fprintf(os.Stderr, "Filed to start server: %s\n", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	<-sigChan

	cancel()

	fmt.Printf("\nSIGINT received, shutting down...\n")
	serv.PrintTotals("final")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		fmt.Fprintf(os.Stderr, "Filed to shutdown server: %s\n", err)
	}
}
