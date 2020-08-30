// Package lrucache implements LRU (Least Recently Used) node
package lrucache

import (
	"encoding/json"
	"github.com/arazmj/gerdu/cache"
	"github.com/arazmj/gerdu/dlinklist"
	"github.com/arazmj/gerdu/metrics"
	"github.com/hashicorp/raft"
	"github.com/inhies/go-bytesize"
	"io"
	"sync"
)

//LRUCache data structure
type LRUCache struct {
	sync.RWMutex
	cache.UnImplementedCache
	node     map[string]*dlinklist.Node
	linklist *dlinklist.DLinkedList
	capacity bytesize.ByteSize
	size     bytesize.ByteSize
}

// NewCache LRUCache constructor
func NewCache(capacity bytesize.ByteSize) *LRUCache {
	l := &LRUCache{
		RWMutex:  sync.RWMutex{},
		node:     map[string]*dlinklist.Node{},
		linklist: dlinklist.NewLinkedList(),
		capacity: capacity,
		size:     0,
	}
	return l
}

// Get returns the value for the key
func (c *LRUCache) Get(key string) (value string, ok bool) {
	defer c.Unlock()
	c.Lock()
	if value, ok := c.node[key]; ok {
		metrics.Hits.Inc()
		node := value
		c.linklist.RemoveNode(node)
		c.linklist.AddNode(node)
		return node.Value, true
	}
	metrics.Miss.Inc()
	return "", false
}

// applyPut updates or insert a new entry, evicts the old entry
// if node size is larger than capacity
func (c *LRUCache) Put(key string, value string) (created bool) {
	defer c.Unlock()
	c.Lock()
	if v, ok := c.node[key]; ok {
		node := v
		c.linklist.RemoveNode(node)
		c.linklist.AddNode(node)
		node.Value = value
		created = false
	} else {
		node := &dlinklist.Node{Key: key, Value: value}
		c.linklist.AddNode(node)
		c.node[key] = node
		metrics.Adds.Inc()
		c.size += bytesize.ByteSize(len(value))
		for c.size > c.capacity {
			metrics.Deletes.Inc()
			tail := c.linklist.PopTail()
			c.size -= bytesize.ByteSize(len(tail.Value))
			delete(c.node, tail.Key)
		}
		created = true
	}
	return created
}

//applyDelete the key from the node
func (c *LRUCache) Delete(key string) (ok bool) {
	if node, ok := c.node[key]; ok {
		metrics.Deletes.Inc()
		c.linklist.RemoveNode(node)
		delete(c.node, key)
	} else {
		return false
	}
	return true
}

func (c *LRUCache) Snapshot() (raft.FSMSnapshot, error) {
	c.RLock()
	defer c.RUnlock()

	o := make(map[string]string)

	for k, v := range c.node {
		o[k] = v.Value
	}

	return &fsmSnapshot{store: o}, nil
}

func (c *LRUCache) Restore(closer io.ReadCloser) error {
	o := make(map[string]string)
	if err := json.NewDecoder(closer).Decode(&o); err != nil {
		return err
	}

	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	for k, v := range o {
		c.Put(k, v)
	}

	return nil
}

type fsmSnapshot struct {
	store map[string]string
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Encode data.
		b, err := json.Marshal(f.store)
		if err != nil {
			return err
		}

		// Write data to sink.
		if _, err := sink.Write(b); err != nil {
			return err
		}

		// Close the sink.
		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return err
}

func (f *fsmSnapshot) Release() {}
