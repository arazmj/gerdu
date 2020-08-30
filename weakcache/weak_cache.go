// Package weakcache implements Weak Cache, eviction of entries are controlled by GC
package weakcache

import (
	"encoding/json"
	"fmt"
	"github.com/arazmj/gerdu/cache"
	"github.com/arazmj/gerdu/metrics"
	"github.com/hashicorp/raft"
	"github.com/ivanrad/go-weakref/weakref"
	"io"
	"sync"
)

// WeakCache data structure
type WeakCache struct {
	sync.Map
	cache.UnImplementedCache
}

// NewWeakCache constructor
func NewWeakCache() *WeakCache {
	return &WeakCache{}
}

// Put a new key value pair
func (c *WeakCache) Put(key string, value string) (created bool) {
	metrics.Adds.Inc()
	ref := weakref.NewWeakRef(value)
	c.Store(key, ref)
	return true
}

// Get value by key
func (c *WeakCache) Get(key string) (value string, ok bool) {
	v, ok := c.Load(key)
	if ok {
		ref := v.(*weakref.WeakRef)
		if ref.IsAlive() {
			metrics.Hits.Inc()
			return ref.GetTarget().(string), true
		}
		metrics.Deletes.Inc()
		c.Delete(key)
	}
	metrics.Miss.Inc()
	return "", false
}

//Delete deletes the key
func (c *WeakCache) Delete(key string) bool {
	metrics.Deletes.Inc()
	c.Map.Delete(key)
	return true
}

func (c *WeakCache) Snapshot() (raft.FSMSnapshot, error) {
	o := make(map[string]string)

	c.Map.Range(func(key, value interface{}) bool {
		ref := value.(*weakref.WeakRef)
		if ref.IsAlive() {
			metrics.Hits.Inc()
			o[fmt.Sprint(key)] = ref.GetTarget().(string)
		}
		return true
	})

	return &fsmSnapshot{store: o}, nil

}

func (c *WeakCache) Restore(closer io.ReadCloser) error {
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
