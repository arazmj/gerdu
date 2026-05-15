package lrucache

import (
	"github.com/inhies/go-bytesize"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestLRUCache(t *testing.T) {
	cache := NewCache(2)
	cache.Put("1", "1")
	cache.Put("2", "2")
	cache.Put("3", "3")
	if value, ok := cache.Get("1"); ok {
		t.Errorf("Expected value 1 to be evicted but got %s %t", value, ok)
	}
	if value, ok := cache.Get("2"); value != "2" && !ok {
		t.Errorf("Expected value 2 but got %s %t", value, ok)
	}
	if value, ok := cache.Get("3"); value != "3" && !ok {
		t.Errorf("Expected value 3 but got %s %t", value, ok)
	}
}

func TestThreadSafety(t *testing.T) {
	capacity, _ := bytesize.Parse("100B")
	cache := NewCache(capacity)
	var wg sync.WaitGroup
	c := 200
	wg.Add(c)

	for i := 0; i < c; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < c; j++ {
				key := strconv.Itoa((i + 1) * j)
				cache.Put(key, key)
				value, ok := cache.Get(key)
				if ok && value != key {
					t.Errorf("The value is not the same %s", value)
				}
			}
		}(i)
	}

	wg.Wait()
}

func BenchmarkLRUCache(b *testing.B) {
	cache := NewCache(100)
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(rand.Int())
		value := strconv.Itoa(rand.Int())
		cache.Put(key, value)
	}

	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(rand.Int())
		cache.Get(key)
	}
}

func TestLruCache_Delete(t *testing.T) {
	cache := NewCache(10)
	cache.Put("1", "1")
	_, getOk1 := cache.Get("1")
	cache.Delete("1")
	_, getOk2 := cache.Get("1")
	if !getOk1 || getOk2 {
		t.Fatal("Expected the ket to be deleted")
	}
}

func TestDeleteDecrementsSize(t *testing.T) {
	cache := NewCache(4)
	deletedValue := "aa"
	remainingValue := "bb"
	newValue := "cc"

	cache.Put("deleted", deletedValue)
	sizeAfterFirstPut := cache.size
	cache.Put("remaining", remainingValue)
	sizeAfterSecondPut := cache.size

	if !cache.Delete("deleted") {
		t.Fatal("expected delete to report success")
	}

	expectedSize := sizeAfterSecondPut - bytesize.ByteSize(len(deletedValue))
	if cache.size != expectedSize {
		t.Fatalf("expected size to drop from %s to %s after deleting %q, got %s", sizeAfterSecondPut, expectedSize, deletedValue, cache.size)
	}
	if expectedSize != sizeAfterFirstPut {
		t.Fatalf("expected remaining size %s to match first put size %s", expectedSize, sizeAfterFirstPut)
	}

	cache.Put("new", newValue)
	if value, ok := cache.Get("remaining"); !ok || value != remainingValue {
		t.Fatalf("expected remaining entry to survive equivalent-size put, got %q %t", value, ok)
	}
}

func TestDeleteConcurrentSafe(t *testing.T) {
	capacity, _ := bytesize.Parse("1KB")
	cache := NewCache(capacity)
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				key := strconv.Itoa((i + 1) * j)
				cache.Put(key, key)
				cache.Delete(key)
			}
		}(i)
	}

	wg.Wait()
}

func TestLRUCachePutWithTTLExpiresOnGet(t *testing.T) {
	cache := NewCache(10)
	cache.PutWithTTL("1", "1", 20*time.Millisecond)
	if value, ok := cache.Get("1"); !ok || value != "1" {
		t.Fatalf("Expected value before TTL expiry, got %q %t", value, ok)
	}

	time.Sleep(40 * time.Millisecond)
	if value, ok := cache.Get("1"); ok {
		t.Fatalf("Expected value to expire, got %q", value)
	}
}

func TestLRUCachePutWithZeroTTLDoesNotExpire(t *testing.T) {
	cache := NewCache(10)
	cache.PutWithTTL("1", "1", 0)
	time.Sleep(20 * time.Millisecond)
	if value, ok := cache.Get("1"); !ok || value != "1" {
		t.Fatalf("Expected value without TTL to remain, got %q %t", value, ok)
	}
}

func TestLRUCacheEvictExpiredRemovesEntries(t *testing.T) {
	cache := NewCache(10)
	cache.PutWithTTL("1", "1", 20*time.Millisecond)
	time.Sleep(40 * time.Millisecond)
	cache.evictExpired(time.Now())
	if _, ok := cache.node["1"]; ok {
		t.Fatal("Expected sweeper eviction to remove expired entry")
	}
}
