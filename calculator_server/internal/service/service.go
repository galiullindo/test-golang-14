package service

import (
	"context"
	"fmt"
	"time"

	"github.com/galiullindo/test-golang-14/calculator_server/internal/repository"
)

type CLib interface {
	Add(a, b int64) int64
	Close()
}
type RustLib interface {
	Sub(a, b int64) int64
	Close()
}
type MetricsObserver interface {
	ObserveCAdd(d float64)
	ObserveRustSub(d float64)
}

type Service struct {
	repo    *repository.CalcRepository
	cLib    CLib
	rustLib RustLib
	metrics MetricsObserver
}

func NewService(
	ctx context.Context,
	interval time.Duration,
	repo *repository.CalcRepository,
	metrics MetricsObserver,
	cLib CLib,
	rustLib RustLib,
) *Service {
	s := &Service{
		repo:    repo,
		metrics: metrics,
		cLib:    cLib,
		rustLib: rustLib,
	}

	go s.printer(ctx, interval)

	return s
}

func (s *Service) ProcessCalc(number int) {
	s.repo.Lock()
	defer s.repo.Unlock()

	sumValue, subValue := s.repo.GetValuesWithoutLock()

	startC := time.Now()
	newSum := s.cLib.Add(sumValue, int64(number))
	s.metrics.ObserveCAdd(time.Since(startC).Seconds())

	startRust := time.Now()
	newSub := s.rustLib.Sub(subValue, int64(number))
	s.metrics.ObserveRustSub(time.Since(startRust).Seconds())

	s.repo.SetValuesWithoutLock(newSum, newSub)
}

func (s *Service) PrintTotals(label string) {
	sumValue, subValue := s.repo.GetValues()
	fmt.Printf("[%s] sum=%d sub=%d\n", label, sumValue, subValue)
}

func (s *Service) printer(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.PrintTotals("periodic")
		case <-ctx.Done():
			return
		}
	}
}
