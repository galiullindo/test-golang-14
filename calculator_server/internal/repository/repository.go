package repository

import "sync"

type CalcRepository struct {
	mu       sync.RWMutex
	sumValue int64
	subValue int64
}

func NewCalcRepository() *CalcRepository {
	return &CalcRepository{}
}

func (r *CalcRepository) Lock() {
	r.mu.Lock()
}
func (r *CalcRepository) Unlock() {
	r.mu.Unlock()
}

func (r *CalcRepository) GetValuesWithoutLock() (int64, int64) {
	return r.sumValue, r.subValue
}
func (r *CalcRepository) GetValues() (int64, int64) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.GetValuesWithoutLock()
}

func (r *CalcRepository) SetValuesWithoutLock(sum, sub int64) {
	r.sumValue = sum
	r.subValue = sub
}
func (r *CalcRepository) SetValues(sum, sub int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.SetValuesWithoutLock(sum, sub)
}
