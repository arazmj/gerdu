package LRUCache

import (
	"GoCache/Cache"
	"GoCache/DLinkList"
	"GoCache/Stats"
	"sync"
)

type LRUCache struct {
	sync.RWMutex
	stats    Stats.StatUpdater
	cache    map[string]*DLinkList.Node
	linklist *DLinkList.DLinkedList
	capacity int
}

func NewCache(capacity int, stats Stats.StatUpdater) Cache.Cache {
	return &LRUCache{
		stats:    stats,
		cache:    map[string]*DLinkList.Node{},
		linklist: DLinkList.NewLinkedList(),
		capacity: capacity,
	}
}

func (c *LRUCache) Get(key string) (value string, ok bool) {
	defer c.Unlock()
	c.Lock()
	if value, ok := c.cache[key]; ok {
		c.stats.HitOps()
		node := value
		c.linklist.RemoveNode(node)
		c.linklist.AddNode(node)
		return node.Value, true
	}
	c.stats.MissOps()
	return "", false
}

func (c *LRUCache) Put(key string, value string) (created bool) {
	defer c.Unlock()
	c.Lock()
	if v, ok := c.cache[key]; ok {
		node := v
		c.linklist.RemoveNode(node)
		c.linklist.AddNode(node)
		node.Value = value
		created = false
	} else {
		node := &DLinkList.Node{Key: key, Value: value}
		c.linklist.AddNode(node)
		c.cache[key] = node
		c.stats.AddOps()
		if len(c.cache) > c.capacity {
			c.stats.DeleteOps()
			tail := c.linklist.PopTail()
			delete(c.cache, tail.Key)
		}
		created = true
	}
	return created
}

func (c *LRUCache) HasKey(key string) bool {
	defer c.RUnlock()
	c.RLock()
	_, ok := c.cache[key]
	return ok
}
