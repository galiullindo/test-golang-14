package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func Worker(
	ctx context.Context,
	client *http.Client,
	workerID int,
	baseUrl string,
	interval time.Duration,
	timeout time.Duration,
	s *Statistics,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	num := strconv.Itoa(rand.IntN(201) - 100)

	u, err := url.Parse(baseUrl)
	if err != nil {
		fmt.Printf("[worker %d] request failed: %s\n", workerID, err)
		return
	}

	q := u.Query()
	q.Set("num", num)
	u.RawQuery = q.Encode()

	for {
		select {
		case <-ticker.C:
			if err := request(ctx, client, u.String(), timeout); err != nil {
				fmt.Printf("[worker %d] request failed: %s\n", workerID, err)
				s.IncrementErrors()
			} else {
				s.IncrementOK()
			}
		case <-ctx.Done():
			return
		}
	}
}

func request(
	ctx context.Context,
	client *http.Client,
	u string,
	timeout time.Duration,
) error {
	requestCtx, requestCancel := context.WithTimeout(ctx, timeout)
	defer requestCancel()

	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, u, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			err = nil
		}
		return err
	}
	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return err
	}

	return nil
}
