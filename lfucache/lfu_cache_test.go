package lfucache

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/inhies/go-bytesize"
	"io"
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	rand.Seed(int64(n))
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func TestThreadSafety(t *testing.T) {
	capacity, _ := bytesize.Parse("3KB")
	cache := NewCache(capacity)
	var wg sync.WaitGroup
	c := 300
	wg.Add(c)

	for i := 0; i < c; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < c; j++ {
				key := RandStringRunes((i + 1) + j)
				cache.Put(key, key)
				newValue, ok := cache.Get(key)
				if ok && key != newValue {
					t.Errorf("The value is not the same for the key %s \n%s",
						key, newValue)
				}
			}
		}(i)
	}

	wg.Wait()
}
func TestNewLFUCache(t *testing.T) {
	c := 100
	size := bytesize.ByteSize(10 + 2*10*9)
	cache := NewCache(size)
	for i := 0; i < c; i++ {
		itoa := strconv.Itoa(i)
		cache.Put(itoa, itoa)
	}

	for i := 0; i < c; i++ {
		itoa := strconv.Itoa(i)
		value, ok := cache.Get(itoa)
		if !ok {
			t.Errorf("Shouldn't evict %s", itoa)
		} else if value != itoa {
			t.Errorf("Does not match %s", itoa)
		}
	}

	if cache.minFreq != 2 {
		t.Errorf("Iterated through all elements the minFreq needs to be 2")
	}

	if len(cache.freq) > 1 {
		t.Errorf("Freq length needs to be 1")
	}

	newValue := strconv.Itoa(c)
	cache.Put(newValue, newValue)

	_, ok := cache.Get("0")
	if ok {
		t.Errorf("0 needs to be evecited by now")
	}
	_, ok = cache.Get(newValue)
	if !ok {
		t.Errorf("%s is not available", newValue)
	}

	newValue = strconv.Itoa(c + 1)
	cache.Put(newValue, newValue)

	_, ok = cache.Get(newValue)
	if !ok {
		t.Errorf("%s is not available", newValue)
	}

	_, ok = cache.Get("0")
	if ok {
		t.Errorf("0 needs to be evecited by now")
	}

}

func TestLFUCache_Update(t *testing.T) {
	cache := NewCache(10)
	cache.Put("20", "20")
	if cache.freq[1].Size() != 1 {
		t.Errorf("Expected size of 1")
	}
	cache.Get("20")

	if _, ok := cache.freq[1]; ok {
		t.Errorf("Expected frequency bucket 1 to be deleted")
	}

	if _, ok := cache.freq[2]; !ok {
		t.Errorf("Expected frequency bucket 2 to exist")
	}

	for i := 0; i < 10; i++ {
		itoa := strconv.Itoa(i)
		cache.Put(itoa, itoa)
	}

	_, ok := cache.Get("0")
	if ok {
		t.Errorf("Expected 0 to be already removed")
	}
}

func TestNewCache2(t *testing.T) {
	cache := NewCache(10)

	cache.Put("1", "1")
	cache.Put("2", "1")
	cache.Put("3", "1")
	cache.Put("4", "1")
	cache.Put("5", "1")
	cache.Put("7", "1")
	cache.Put("8", "1")
	cache.Put("9", "12345")

	_, ok := cache.Get("1")
	if ok {
		t.Errorf("1 needs to be evicted.")
	}

	_, ok = cache.Get("2")
	if ok {
		t.Errorf("2 needs to be efecited.")
	}

	_, ok = cache.Get("3")

	if !ok {
		t.Errorf("3 should not be eficted")
	}

	cache.Put("10", "1")

	_, ok = cache.Get("3")

	if !ok {
		t.Errorf("3 should not be eficted")
	}

	_, ok = cache.Get("4")

	if ok {
		t.Errorf("4 should be eficted")
	}
}

func TestLfuCache_Delete(t *testing.T) {
	cache := NewCache(10)
	cache.Put("1", "1")
	_, getOk1 := cache.Get("1")
	cache.Delete("1")
	_, getOk2 := cache.Get("1")
	if !getOk1 || getOk2 {
		t.Fatal("Expected the ket to be deleted")
	}
}

func TestLFUCache_MissesCapacityAndEvictionOrder(t *testing.T) {
	zero := NewCache(0)
	if zero.Put("a", "1") {
		t.Fatal("zero-capacity cache should not create entries")
	}
	if value, ok := zero.Get("a"); ok || value != "" {
		t.Fatalf("expected miss from zero-capacity cache, got %q %t", value, ok)
	}
	if zero.Delete("missing") {
		t.Fatal("delete of missing key should return false")
	}

	cache := NewCache(2)
	cache.Put("a", "aa")
	if value, ok := cache.Get("a"); !ok || value != "aa" {
		t.Fatalf("expected exact-capacity value to remain, got %q %t", value, ok)
	}

	cache = NewCache(2)
	cache.Put("a", "1")
	cache.Put("b", "2")
	cache.Put("c", "3")
	if _, ok := cache.Get("a"); ok {
		t.Fatal("expected oldest key among equal frequencies to be evicted")
	}
	if _, ok := cache.Get("b"); !ok {
		t.Fatal("expected newer key b to remain after tie eviction")
	}
	if _, ok := cache.Get("c"); !ok {
		t.Fatal("expected inserted key c to remain")
	}

	cache = NewCache(2)
	cache.Put("a", "1")
	cache.Put("b", "2")
	cache.Get("a")
	cache.Put("c", "3")
	if _, ok := cache.Get("b"); ok {
		t.Fatal("expected least frequently used key b to be evicted")
	}
	if value, ok := cache.Get("a"); !ok || value != "1" {
		t.Fatalf("expected frequently used key a to remain, got %q %t", value, ok)
	}
	if cache.Put("a", "A") {
		t.Fatal("updating an existing key should not create")
	}
	if value, ok := cache.Get("a"); !ok || value != "A" {
		t.Fatalf("expected updated value, got %q %t", value, ok)
	}
}

func TestLFUCache_SnapshotLifecycle(t *testing.T) {
	cache := NewCache(20)
	cache.Put("a", "1")
	cache.Put("b", "2")
	cache.Get("a")

	snapshot, err := cache.Snapshot()
	if err != nil {
		t.Fatalf("snapshot failed: %v", err)
	}
	sink := &recordingSnapshotSink{}
	if err := snapshot.Persist(sink); err != nil {
		t.Fatalf("persist failed: %v", err)
	}
	if !sink.closed || sink.canceled {
		t.Fatalf("expected clean close without cancel, closed=%t canceled=%t", sink.closed, sink.canceled)
	}
	snapshot.Release()

	var payload map[string]string
	if err := json.Unmarshal(sink.Bytes(), &payload); err != nil {
		t.Fatalf("snapshot payload is not json: %v", err)
	}
	if payload["a"] != "1" || payload["b"] != "2" {
		t.Fatalf("unexpected snapshot payload: %#v", payload)
	}

	restored := NewCache(20)
	if err := restored.Restore(io.NopCloser(bytes.NewReader(sink.Bytes()))); err != nil {
		t.Fatalf("restore failed: %v", err)
	}
	if value, ok := restored.Get("a"); !ok || value != "1" {
		t.Fatalf("expected restored value, got %q %t", value, ok)
	}
	if err := restored.Restore(io.NopCloser(bytes.NewBufferString("not-json"))); err == nil {
		t.Fatal("expected invalid restore payload to fail")
	}

	failing := &recordingSnapshotSink{failClose: true}
	if err := snapshot.Persist(failing); err == nil || !failing.canceled {
		t.Fatalf("expected failing close to cancel, err=%v canceled=%t", err, failing.canceled)
	}
}

type recordingSnapshotSink struct {
	bytes.Buffer
	closed    bool
	canceled  bool
	failWrite bool
	failClose bool
}

func (s *recordingSnapshotSink) ID() string { return "test" }

func (s *recordingSnapshotSink) Close() error {
	s.closed = true
	if s.failClose {
		return errors.New("close failed")
	}
	return nil
}

func (s *recordingSnapshotSink) Cancel() error {
	s.canceled = true
	return nil
}

func (s *recordingSnapshotSink) Write(p []byte) (int, error) {
	if s.failWrite {
		return 0, errors.New("write failed")
	}
	return s.Buffer.Write(p)
}
