package weakcache

import (
	"GoCache/stats"
	"github.com/ivanrad/go-weakref/weakref"
	"sync"
)

// WeakCache data structure
type WeakCache struct {
	sync.Map
}

// NewWeakCache constructor
func NewWeakCache() *WeakCache {
	return &WeakCache{}
}

// Put a new key value pair
func (c *WeakCache) Put(key string, value string) (created bool) {
	stats.Adds.Inc()
	ref := weakref.NewWeakRef(value)
	c.Store(key, ref)
	return true
}

// Get value by key
func (c *WeakCache) Get(key string) (value string, ok bool) {
	v, ok := c.Load(key)
	if ok {
		ref := v.(*weakref.WeakRef)
		if ref.IsAlive() {
			stats.Hits.Inc()
			return ref.GetTarget().(string), true
		}
		stats.Deletes.Inc()
		c.Delete(key)
	}
	stats.Miss.Inc()
	return "", false
}
