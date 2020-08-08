package LRUCache

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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

type Stats struct {
	hits prometheus.Counter
	miss prometheus.Counter
	adds prometheus.Counter
	dels prometheus.Counter
}
type LRUCache struct {
	sync.RWMutex
	Stats
	cache    map[string]*Node
	head     *Node
	tail     *Node
	capacity int
}

func NewCache(capacity int) LRUCache {
	head := &Node{}
	tail := &Node{}
	head.next = tail
	tail.prev = head
	return LRUCache{
		Stats: Stats{
			miss: promauto.NewCounter(prometheus.CounterOpts{
				Name: "go_cache_misses_total",
				Help: "The total number of missed cache hits",
			}),
			hits: promauto.NewCounter(prometheus.CounterOpts{
				Name: "go_cache_hits_total",
				Help: "The total number of cache hits",
			}),
			adds: promauto.NewCounter(prometheus.CounterOpts{
				Name: "go_cache_adds_total",
				Help: "The total number of new added nodes",
			}),
			dels: promauto.NewCounter(prometheus.CounterOpts{
				Name: "go_cache_deletes_total",
				Help: "The total number of deletes nodes",
			}),
		},
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
		c.hits.Inc()
		node := value
		removeNode(node)
		c.addNode(node)
		return node.value, true
	}
	c.miss.Inc()
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
		c.adds.Inc()
		if len(c.cache) > c.capacity {
			c.dels.Inc()
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
