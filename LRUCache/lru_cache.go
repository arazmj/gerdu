package LRUCache

import (
	"GoCache/Cache"
	"GoCache/DLinkList"
	"GoCache/Stats"
	"github.com/inhies/go-bytesize"
	"sync"
)

type LRUCache struct {
	sync.RWMutex
	cache    map[string]*DLinkList.Node
	linklist *DLinkList.DLinkedList
	capacity bytesize.ByteSize
	size     bytesize.ByteSize
}

func NewCache(capacity bytesize.ByteSize) Cache.Cache {
	return &LRUCache{
		cache:    map[string]*DLinkList.Node{},
		linklist: DLinkList.NewLinkedList(),
		capacity: capacity,
		size:     0,
	}
}

func (c *LRUCache) Get(key string) (value string, ok bool) {
	defer c.Unlock()
	c.Lock()
	if value, ok := c.cache[key]; ok {
		Stats.Hits.Inc()
		node := value
		c.linklist.RemoveNode(node)
		c.linklist.AddNode(node)
		return node.Value, true
	}
	Stats.Miss.Inc()
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
		Stats.Adds.Inc()
		c.size += bytesize.ByteSize(len(value))
		for c.size > c.capacity {
			Stats.Deletes.Inc()
			tail := c.linklist.PopTail()
			c.size -= bytesize.ByteSize(len(tail.Value))
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
