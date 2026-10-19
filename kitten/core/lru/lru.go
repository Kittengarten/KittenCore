/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package lru implements an LRU cache.
package lru

import (
	"sync"

	"github.com/Kittengarten/KittenCore/kitten/core/list"
)

// Cache is an LRU cache. It is safe for concurrent access.
type Cache[K comparable, V any] struct {
	sync.Mutex
	*list.List[entry[K, V]]
	// MaxEntries is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	MaxEntries int
	// OnEvicted optionally specifies a callback function to be
	// executed when an entry is purged from the cache.
	OnEvicted func(K, V)
	cache     map[K]*list.Element[entry[K, V]]
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

// Key 获取键
func (e entry[K, V]) Key() K {
	return e.key
}

// Value 获取值
func (e entry[K, V]) Value() V {
	return e.value
}

// New creates a new Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func New[K comparable, V any](maxEntries int) *Cache[K, V] {
	if maxEntries < 0 {
		panic(`lru: negative MaxEntries`)
	}
	return new(Cache[K, V]{
		MaxEntries: maxEntries,
		List:       list.New[entry[K, V]](),
		cache:      make(map[K]*list.Element[entry[K, V]]),
	})
}

// Add adds a value to the cache.
func (c *Cache[K, V]) Add(key K, value V) {
	c.Lock()
	defer c.Unlock()
	if c.cache == nil {
		c.cache = make(map[K]*list.Element[entry[K, V]])
		c.List = list.New[entry[K, V]]()
	}
	if ee, ok := c.cache[key]; ok {
		c.MoveToFront(ee)
		ee.Value.value = value
		return
	}
	c.cache[key] = c.PushFront(entry[K, V]{key, value})
	if c.MaxEntries != 0 && len(c.cache) > c.MaxEntries {
		c.removeOldest()
	}
}

// Get looks up a key's value from the cache.
func (c *Cache[K, V]) Get(key K) (value V, ok bool) {
	c.Lock()
	defer c.Unlock()
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.MoveToFront(ele)
		return ele.Value.value, true
	}
	return
}

// Remove removes the provided key from the cache.
func (c *Cache[K, V]) Remove(key K) {
	c.Lock()
	defer c.Unlock()
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

// RemoveOldest removes the oldest item from the cache.
func (c *Cache[K, V]) RemoveOldest() {
	c.Lock()
	defer c.Unlock()
	if c.cache == nil {
		return
	}
	c.removeOldest()
}

func (c *Cache[K, V]) removeOldest() {
	if ele := c.Back(); ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache[K, V]) removeElement(e *list.Element[entry[K, V]]) {
	c.List.Remove(e)
	delete(c.cache, e.Value.key)
	if c.OnEvicted != nil {
		c.Unlock()
		c.OnEvicted(e.Value.key, e.Value.value)
		c.Lock()
	}
}

// Len returns the number of items in the cache.
func (c *Cache[K, V]) Len() int {
	c.Lock()
	defer c.Unlock()
	return len(c.cache)
}

// Clear purges all stored items from the cache.
func (c *Cache[K, V]) Clear() {
	c.Lock()
	defer c.Unlock()
	if c.OnEvicted != nil {
		for _, e := range c.cache {
			c.Unlock()
			c.OnEvicted(e.Value.key, e.Value.value)
			c.Lock()
		}
	}
	c.List = nil
	c.cache = nil
}

// Range calls f sequentially for each key and value present in the cache.
func (c *Cache[K, V]) Range(f func(key K, value V) bool) {
	c.Lock()
	defer c.Unlock()
	for e := c.Front(); e != nil; e = e.Next() {
		c.Unlock()
		if !f(e.Value.key, e.Value.value) {
			c.Lock()
			return
		}
		c.Lock()
	}
}
