package LRUCache

import (
	"GoCache/Stats"
	"sync"
)

type Node struct {
	next  *Node
	prev  *Node
	key   string
	value string
}

func (c *LRUCache) addNode(node *Node) {
	next := c.head.next
	c.head.next = node
	next.prev = node
	node.next = next
	node.prev = c.head
}

func removeNode(node *Node) {
	prev := node.prev
	next := node.next
	prev.next = next
	next.prev = prev
}

func (c *LRUCache) popTail() *Node {
	prev := c.tail.prev
	removeNode(prev)
	return prev
}

type LRUCache struct {
	sync.RWMutex
	stats    Stats.Statser
	cache    map[string]*Node
	head     *Node
	tail     *Node
	capacity int
}

func NewCache(capacity int, stats Stats.Statser) LRUCache {
	head := &Node{}
	tail := &Node{}
	head.next = tail
	tail.prev = head
	return LRUCache{
		stats:    stats,
		cache:    map[string]*Node{},
		head:     head,
		tail:     tail,
		capacity: capacity,
	}
}

func (c *LRUCache) Get(key string) (value string, ok bool) {
	defer c.Unlock()
	c.Lock()
	if value, ok := c.cache[key]; ok {
		c.stats.HitOps()
		node := value
		removeNode(node)
		c.addNode(node)
		return node.value, true
	}
	c.stats.MissOps()
	return "", false
}

func (c *LRUCache) Put(key string, value string) {
	defer c.Unlock()
	c.Lock()
	if v, ok := c.cache[key]; ok {
		node := v
		removeNode(node)
		c.addNode(node)
		node.value = value
	} else {
		node := &Node{key: key, value: value}
		c.addNode(node)
		c.cache[key] = node
		c.stats.AddOps()
		if len(c.cache) > c.capacity {
			c.stats.DeleteOps()
			tail := c.popTail()
			delete(c.cache, tail.key)
		}
	}
}

func (c *LRUCache) HasKey(key string) bool {
	defer c.RUnlock()
	c.RLock()
	_, ok := c.cache[key]
	return ok
}
