package LRUCache

import "sync"

type Node struct {
	next 	*Node
	prev 	*Node
	key 	string
	value 	string
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
	cache 	sync.Map
	head 	*Node
	tail 	*Node
	capacity int
}


func NewCache(capacity int) LRUCache {
	head := &Node{}
	tail := &Node{}
	head.next = tail
	tail.prev = head
	return LRUCache {
		cache: sync.Map{},
		head: head,
		tail: tail,
		capacity: capacity,
	}
}


func (c *LRUCache) Get(key string) (value string, ok bool) {
	if value, ok := c.cache.Load(key); ok {
		node := value.(*Node)
		removeNode(node)
		c.addNode(node)
		return node.value, true
	}
	return "", false
}

func (c *LRUCache) Put(key string, value string)  {
	if v, ok := c.cache.Load(key); ok {
		node := v.(*Node)
		removeNode(node)
		c.addNode(node)
		node.value = value
	} else {
		node := &Node{ key: key, value: value}
		c.addNode(node)
		c.cache.Store(key, node)
		if len(c.cache) > c.capacity {
			tail := c.popTail()
			c.cache.Delete(tail.key)
		}
	}
}

func len(sm sync.Map) (length int) {
	sm.Range(func(_, _ interface{}) bool {
		length++
		return true
	})
	return length
}


