package worker

import (
	"context"
	"errors"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/galiullindo/test-golang-14/generator/internal/metrics"
)

type Worker struct {
	ID       int
	interval time.Duration

	httpClient *http.Client
	endpoint   string
	timeout    time.Duration

	requestStatistics *metrics.RequestStatistics
}

// New ...
func New(
	ID int,
	interval time.Duration,
	httpClient *http.Client,
	endpoint string,
	timeout time.Duration,
	requestStatistics *metrics.RequestStatistics,
) *Worker {
	return &Worker{
		ID:                ID,
		interval:          interval,
		httpClient:        httpClient,
		endpoint:          endpoint,
		timeout:           timeout,
		requestStatistics: requestStatistics,
	}
}

// Run ...
func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	u, err := url.Parse(w.endpoint)
	if err != nil {
		// TODO: log: worker {w.ID}: {err}
		return err
	}

	num := strconv.Itoa(rand.IntN(201) - 100)
	q := u.Query()
	q.Set("num", num)
	u.RawQuery = q.Encode()

	for {
		select {
		case <-ticker.C:
			err := w.request(ctx, u.String())
			if errors.Is(err, context.Canceled) {
				// TODO: log: worker {w.ID}: context canceled
				err = nil
			}

			if err != nil {
				// TODO: log: worker {w.ID}: {err}
				w.requestStatistics.IncrementErrors()
			} else {
				w.requestStatistics.IncrementOK()
			}
		case <-ctx.Done():
			return nil
		}
	}

}

// request ...
func (w *Worker) request(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return err
	}

	return nil
}
