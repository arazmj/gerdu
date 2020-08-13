package WeakCache

import (
	"GoCache/Stats"
	"github.com/ivanrad/go-weakref/weakref"
	"sync"
)

type WeakCache struct {
	sync.Map
	stats *Stats.Stats
}

func NewWeakCache(stats *Stats.Stats) *WeakCache {
	return &WeakCache{stats: stats}
}

func (c *WeakCache) Put(key string, value string) (created bool) {
	c.stats.AddOps()
	ref := weakref.NewWeakRef(value)
	c.Store(key, ref)
	return true
}

func (c *WeakCache) Get(key string) (value string, ok bool) {
	v, ok := c.Load(key)
	if ok {
		ref := v.(*weakref.WeakRef)
		if ref.IsAlive() {
			c.stats.HitOps()
			return ref.GetTarget().(string), true
		} else {
			c.stats.DeleteOps()
			c.Delete(key)
		}
	}
	c.stats.MissOps()
	return "", false
}
