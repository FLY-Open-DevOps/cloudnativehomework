package module10

import (
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/exp/slog"
)

var (
	r *rand.Rand
)

func init() {
	r = rand.New(rand.NewSource(time.Now().Unix()))
}

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

// caculate value of fibonacci(n)
func fibo(n int) int {
	if n < 2 {
		return n
	}
	return fibo(n-1) + fibo(n-2)
}

// create a custom delay from 0 ~ 2 seconds
func Delay() {
	time.Sleep(time.Millisecond * time.Duration(r.Intn(2000)))
}
