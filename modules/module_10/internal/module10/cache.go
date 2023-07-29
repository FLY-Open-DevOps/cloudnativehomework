package module10

import (
	"errors"
	"sync"
)

var errCacheMiss = errors.New("cache miss")

type Cache struct {
	mu   sync.RWMutex
	data map[int]int
}

func NewCache() *Cache {
	return &Cache{data: make(map[int]int)}
}

func (c *Cache) Get(key int) (v int, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[key]
	if !ok {
		err = errCacheMiss
	}
	return
}

func (c *Cache) Set(k, v int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[k] = v
}
