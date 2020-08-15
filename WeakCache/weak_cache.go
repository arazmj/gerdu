package WeakCache

import (
	"GoCache/Stats"
	"github.com/ivanrad/go-weakref/weakref"
	"sync"
)

type WeakCache struct {
	sync.Map
}

func NewWeakCache() *WeakCache {
	return &WeakCache{}
}

func (c *WeakCache) Put(key string, value string) (created bool) {
	Stats.Adds.Inc()
	ref := weakref.NewWeakRef(value)
	c.Store(key, ref)
	return true
}

func (c *WeakCache) Get(key string) (value string, ok bool) {
	v, ok := c.Load(key)
	if ok {
		ref := v.(*weakref.WeakRef)
		if ref.IsAlive() {
			Stats.Hits.Inc()
			return ref.GetTarget().(string), true
		} else {
			Stats.Deletes.Inc()
			c.Delete(key)
		}
	}
	Stats.Miss.Inc()
	return "", false
}
