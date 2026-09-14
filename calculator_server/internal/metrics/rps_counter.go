package metrics

import (
	"context"
	"sync"
	"time"
)

type RPSCounter struct {
	mu      sync.Mutex
	buckets [60]uint64 // 60 buckets per second
	last    time.Time
}

func NewRPSCounter(ctx context.Context) *RPSCounter {
	c := &RPSCounter{last: time.Now()}
	// go c.cleaner(ctx)
	return c
}

func (c *RPSCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.unlockedCleanup(now)

	idx := now.Unix() % 60
	c.buckets[idx]++
}
func (c *RPSCounter) History() [60]uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.unlockedCleanup(now)

	var history [60]uint64
	currentIdx := now.Unix() % 60
	for i := 0; i < 60; i++ {
		idx := (currentIdx + int64(i) + 1) % 60
		history[i] = c.buckets[idx]
	}

	return history
}

func (c *RPSCounter) cleaner(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			c.cleanup(now)
		}
	}
}
func (c *RPSCounter) cleanup(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.unlockedCleanup(now)
}
func (c *RPSCounter) unlockedCleanup(now time.Time) {
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
