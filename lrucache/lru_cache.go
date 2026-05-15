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
	"time"
)

const sweepInterval = 30 * time.Second

// LRUCache data structure
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
	go l.sweepExpiredEntries()
	return l
}

func isExpired(node *dlinklist.Node, now time.Time) bool {
	return !node.ExpiresAt.IsZero() && now.After(node.ExpiresAt)
}

func expirationFromTTL(ttl time.Duration) time.Time {
	if ttl <= 0 {
		return time.Time{}
	}
	return time.Now().Add(ttl)
}

func (c *LRUCache) removeNode(node *dlinklist.Node) {
	metrics.Deletes.Inc()
	c.linklist.RemoveNode(node)
	delete(c.node, node.Key)
	c.size -= bytesize.ByteSize(len(node.Value))
}

// Get returns the value for the key
func (c *LRUCache) Get(key string) (value string, ok bool) {
	defer c.Unlock()
	c.Lock()
	if node, ok := c.node[key]; ok {
		if isExpired(node, time.Now()) {
			c.removeNode(node)
			metrics.Miss.Inc()
			return "", false
		}
		metrics.Hits.Inc()
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
	return c.PutWithTTL(key, value, 0)
}

// PutWithTTL updates or inserts a new entry with an optional TTL.
func (c *LRUCache) PutWithTTL(key string, value string, ttl time.Duration) (created bool) {
	defer c.Unlock()
	c.Lock()
	expiresAt := expirationFromTTL(ttl)
	if node, ok := c.node[key]; ok {
		c.size -= bytesize.ByteSize(len(node.Value))
		c.size += bytesize.ByteSize(len(value))
		c.linklist.RemoveNode(node)
		c.linklist.AddNode(node)
		node.Value = value
		node.ExpiresAt = expiresAt
		created = false
	} else {
		node := &dlinklist.Node{Key: key, Value: value, ExpiresAt: expiresAt}
		c.linklist.AddNode(node)
		c.node[key] = node
		metrics.Adds.Inc()
		c.size += bytesize.ByteSize(len(value))
		created = true
	}
	for c.size > c.capacity {
		tail := c.linklist.PopTail()
		metrics.Deletes.Inc()
		c.size -= bytesize.ByteSize(len(tail.Value))
		delete(c.node, tail.Key)
	}
	return created
}

// applyDelete the key from the node
func (c *LRUCache) Delete(key string) (ok bool) {
	c.Lock()
	defer c.Unlock()

	if node, ok := c.node[key]; ok {
		c.removeNode(node)
		return true
	}
	return false
}

func (c *LRUCache) sweepExpiredEntries() {
	ticker := time.NewTicker(sweepInterval)
	for range ticker.C {
		c.evictExpired(time.Now())
	}
}

func (c *LRUCache) evictExpired(now time.Time) {
	c.Lock()
	defer c.Unlock()

	for _, node := range c.node {
		if isExpired(node, now) {
			c.removeNode(node)
		}
	}
}

func (c *LRUCache) Snapshot() (raft.FSMSnapshot, error) {
	c.RLock()
	defer c.RUnlock()

	o := make(map[string]string)
	now := time.Now()

	for k, v := range c.node {
		if !isExpired(v, now) {
			o[k] = v.Value
		}
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
