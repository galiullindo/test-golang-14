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

	"github.com/galiullindo/test-golang-14/generator/internal"
	"github.com/spf13/pflag"
)

type arguments struct {
	URL      string
	Threads  int
	Interval time.Duration
	Timeout  time.Duration
}

func main() {
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Calculator HTTP server\n")
		pflag.PrintDefaults()
	}

	var args arguments
	pflag.StringVar(&args.URL, "url", "http://localhost:8080/calc", "calculator endpoint")
	pflag.IntVarP(&args.Threads, "threads", "n", 10, "number of worker threads")
	pflag.DurationVar(
		&args.Interval,
		"interval",
		100*time.Microsecond,
		"pause between requests per thread (e.g. 100ms, 0.5s, 1s; 0s = as fast as possible)",
	)
	pflag.DurationVar(
		&args.Timeout,
		"timeout",
		5*time.Second,
		"HTTP request timeout (e.g. 100ms, 0.5s, 1s, 5s)",
	)
	pflag.Parse()

	stat := &internal.Statistics{}
	client := &http.Client{}
	wg := &sync.WaitGroup{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(args.Threads)
	for i := range args.Threads {
		go func() {
			defer wg.Done()
			internal.Worker(ctx, client, i, args.URL, args.Interval, args.Timeout, stat)
		}()
	}

	fmt.Printf("Generator started: %d threads -> %s", args.Threads, args.URL)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	<-sigChan

	fmt.Println("\nSIGINT received, stopping generator...")

	cancel()
	wg.Wait()

	fmt.Printf("Total requests: ok=%d errors=%d", stat.OK(), stat.Errors())
}
