package cache

import (
	"context"
	"time"

	dataloader "github.com/graph-gophers/dataloader/v7"

	lru "github.com/hashicorp/golang-lru/v2/expirable"
)

// LRU implements the dataloader.Cache interface
type LRU[K comparable, V any] struct {
	innerLru *lru.LRU[K, dataloader.Thunk[V]]
}

// NewLRU creates a new LRU cache
// size is the size of the cache. If size is 0, the cache has no limit
func NewLRU[K comparable, V any](size int, ttl time.Duration) *LRU[K, V] {
	l := lru.NewLRU[K, dataloader.Thunk[V]](size, nil, ttl)
	return &LRU[K, V]{innerLru: l}
}

// Get gets an item from the cache
func (c *LRU[K, V]) Get(_ context.Context, key K) (dataloader.Thunk[V], bool) {
	return c.innerLru.Get(key)
}

// Set sets an item in the LRU
func (c *LRU[K, V]) Set(_ context.Context, key K, value dataloader.Thunk[V]) {
	c.innerLru.Add(key, value)
}

// Delete deletes an item in the cache
func (c *LRU[K, V]) Delete(_ context.Context, key K) bool {
	return c.innerLru.Remove(key)
}

// Clear clears the cache
func (c *LRU[K, V]) Clear() {
	c.innerLru.Purge()
}
