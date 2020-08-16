// Package lrucache implements LRU (Least Recently Used) cache
package lrucache

import (
	"github.com/arazmj/gerdu/dlinklist"
	"github.com/arazmj/gerdu/metrics"
	"github.com/inhies/go-bytesize"
	"sync"
)

//LruCache data structure
type LruCache struct {
	sync.RWMutex
	cache    map[string]*dlinklist.Node
	linklist *dlinklist.DLinkedList
	capacity bytesize.ByteSize
	size     bytesize.ByteSize
}

// NewCache LruCache constructor
func NewCache(capacity bytesize.ByteSize) *LruCache {
	return &LruCache{
		cache:    map[string]*dlinklist.Node{},
		linklist: dlinklist.NewLinkedList(),
		capacity: capacity,
		size:     0,
	}
}

// Get returns the value for the key
func (c *LruCache) Get(key string) (value string, ok bool) {
	defer c.Unlock()
	c.Lock()
	if value, ok := c.cache[key]; ok {
		metrics.Hits.Inc()
		node := value
		c.linklist.RemoveNode(node)
		c.linklist.AddNode(node)
		return node.Value, true
	}
	metrics.Miss.Inc()
	return "", false
}

// Put updates or insert a new entry, evicts the old entry
// if cache size is larger than capacity
func (c *LruCache) Put(key string, value string) (created bool) {
	defer c.Unlock()
	c.Lock()
	if v, ok := c.cache[key]; ok {
		node := v
		c.linklist.RemoveNode(node)
		c.linklist.AddNode(node)
		node.Value = value
		created = false
	} else {
		node := &dlinklist.Node{Key: key, Value: value}
		c.linklist.AddNode(node)
		c.cache[key] = node
		metrics.Adds.Inc()
		c.size += bytesize.ByteSize(len(value))
		for c.size > c.capacity {
			metrics.Deletes.Inc()
			tail := c.linklist.PopTail()
			c.size -= bytesize.ByteSize(len(tail.Value))
			delete(c.cache, tail.Key)
		}
		created = true
	}
	return created
}

// HasKey indicates the key exists or not
func (c *LruCache) HasKey(key string) bool {
	defer c.RUnlock()
	c.RLock()
	_, ok := c.cache[key]
	return ok
}

//Delete the key from the cache
func (c *LruCache) Delete(key string) (ok bool) {
	if node, ok := c.cache[key]; ok {
		metrics.Deletes.Inc()
		c.linklist.RemoveNode(node)
		delete(c.cache, key)
	} else {
		return false
	}
	return true
}
