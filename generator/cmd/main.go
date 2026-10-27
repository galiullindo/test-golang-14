package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"sync"
	"syscall"
	"time"

	"github.com/galiullindo/test-golang-14/generator/internal/metrics"
	"github.com/galiullindo/test-golang-14/generator/internal/worker"
	"github.com/spf13/pflag"
)

func main() {
	args := parseArguments()
	requestsStatistics := metrics.NewRequestStatistics()
	httpClient := &http.Client{}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	wg := &sync.WaitGroup{}
	wg.Add(args.Threads)
	for ID := range args.Threads {
		go func() {
			defer wg.Done()

			w := worker.New(
				ID,
				args.Interval,
				httpClient,
				args.Endpoint,
				args.Timeout,
				requestsStatistics,
			)

			if err := w.Run(ctx); err != nil {
				//TODO: log: worker {ID}: {err}
			}
		}()
	}

	// TODO: log: generator started: {arguments.Threads} threads -> {arguments.URL}
	<-ctx.Done()
	// TODO: log: \nSIGINT received, stopping generator...
	wg.Wait()
	// TODO: log: total requests: ok={requestsStatistics.OK()} errors={requestsStatistics.Errors()}
}

type arguments struct {
	Endpoint string
	Threads  int
	Interval time.Duration
	Timeout  time.Duration
}

func parseArguments() arguments {
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Calculator HTTP server\n")
		pflag.PrintDefaults()
	}

	var a arguments

	// --endpoint
	pflag.StringVar(
		&a.Endpoint,
		"endpoint",
		"http://localhost:8080/calc",
		"calculator endpoint",
	)

	// --threads
	pflag.IntVar(
		&a.Threads,
		"threads",
		10,
		"number of worker threads",
	)

	// --interval
	pflag.DurationVar(
		&a.Interval,
		"interval",
		100*time.Millisecond,
		"pause between requests per thread (e.g. 100ms, 0.5s, 1s; 0s = as fast as possible)",
	)

	// --timeout
	pflag.DurationVar(
		&a.Timeout,
		"timeout",
		5*time.Second,
		"HTTP request timeout (e.g. 100ms, 0.5s, 1s, 5s)",
	)

	pflag.Parse()

	return a
}
