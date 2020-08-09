package LFUCache

import (
	"GoCache/Cache"
	"GoCache/DLinkList"
	"GoCache/Stats"
	"sync"
)

type LFUCache struct {
	sync.Mutex
	stats    *Stats.Stats
	size     int
	capacity int
	node     map[string]*DLinkList.Node
	freq     map[int]*DLinkList.DLinkedList
	minFreq  int
}

func NewCache(capacity int, stats *Stats.Stats) Cache.Cache {
	return &LFUCache{
		size:     0,
		capacity: capacity,
		stats:    stats,
		node:     map[string]*DLinkList.Node{},
		freq:     map[int]*DLinkList.DLinkedList{},
		minFreq:  0,
	}
}

/*
	This is a helper function that used in the following two cases:

	1. when Get(key)` is called; and
	2. when Put(key, value)` is called and the key exists.

	The common point of these two cases is that:

	1. no new node comes in, and
	2. the node is visited one more times -> node.freq changed ->
	thus the place of this node will change

	The logic of this function is:

	1. pop the node from the old DLinkedList (with freq `f`)
	2. append the node to new DLinkedList (with freq `f+1`)
	3. if old DLinkedList has size 0 and minFreq is `f`,
	update minFreq to `f+1`

	All of the above operations took O(1) time.
*/

func (c *LFUCache) update(node *DLinkList.Node) {
	freq := node.Freq

	c.freq[freq].RemoveNode(node)
	if v, _ := c.freq[freq]; c.minFreq == freq && v.Size == 0 {
		delete(c.freq, freq)
		c.minFreq += 1
	}

	node.Freq++
	freq = node.Freq
	if _, ok := c.freq[freq]; !ok {
		c.freq[freq] = DLinkList.NewLinkedList()
	}
	c.freq[freq].AddNode(node)
}

/*
	Through checking node[key], we can get the node in O(1) time.
	Just performs update, then we can return the value of node.
*/
func (c *LFUCache) Get(key string) (value string, ok bool) {
	defer c.Unlock()
	c.Lock()

	if _, ok := c.node[key]; !ok {
		c.stats.MissOps()
		return "", false
	}

	c.stats.HitOps()
	node := c.node[key]
	c.update(node)
	return node.Value, true
}

/*
	If `key` already exists in self._node, we do the same operations as `get`, except
	updating the node.val to new value.	Otherwise
	1. if the cache reaches its capacity, pop the least frequently used item. (*)
	2. add new node to self._node
	3. add new node to the DLinkedList with frequency 1
	4. reset minFreq to 1

	(*) How to pop the least frequently used item? Two facts:

	1. we maintain the minFreq, the minimum possible frequency in cache.
	2. All cache with the same frequency are stored as a DLinkedList, with
	recently used order (Always append at head)
	3. The tail of the DLinkedList with minFreq is the least
	recently used one, pop it.
*/
func (c *LFUCache) Put(key, value string) {
	defer c.Unlock()
	c.Lock()
	if c.capacity == 0 {
		return
	}
	if _, ok := c.node[key]; ok {
		c.stats.HitOps()
		node := c.node[key]
		c.update(node)
		node.Value = value
	} else {
		if c.size == c.capacity {
			c.stats.DeleteOps()
			node := c.freq[c.minFreq].PopTail()
			delete(c.node, node.Key)
			c.size--
		}
		c.stats.AddOps()
		node := &DLinkList.Node{
			Key:   key,
			Value: value,
			Freq:  1,
		}
		c.node[key] = node
		if _, ok := c.freq[1]; !ok {
			c.freq[1] = DLinkList.NewLinkedList()
		}
		c.freq[1].AddNode(node)
		c.minFreq = 1
		c.size++
	}
}
