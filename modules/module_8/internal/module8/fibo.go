package module8

import (
	"fmt"

	"golang.org/x/exp/slog"
)

type FiboCaculator struct {
	cache *Cache
	// max sequence can be caculate in fibonacci array
	maxseq int
	// if true, then cache result in each call
	cacheResult bool
	logger      *slog.Logger
}

func NewFiboCaculator(maxseq int, cacheResult bool, logger *slog.Logger) *FiboCaculator {
	return &FiboCaculator{
		NewCache(),
		maxseq,
		cacheResult,
		logger,
	}
}

func (f *FiboCaculator) Caculate(idx int) (n int, err error) {
	if idx > f.maxseq {
		err = fmt.Errorf("index %d is out of range [0, %d]", idx, f.maxseq)
		return
	}
	if f.cacheResult {
		n, err := f.cache.Get(idx)
		if err == nil {
			return n, nil
		}
		f.logger.Debug("cache miss", "key", idx)
	}
	n = fibo(idx)
	f.cache.Set(idx, n)
	f.logger.Debug("cache fibo", "key", idx, "value", n)
	return
}

func fibo(n int) int {
	if n < 2 {
		return n
	}
	return fibo(n-1) + fibo(n-2)
}
