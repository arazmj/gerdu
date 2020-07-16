package LRUCache

import (
	"math/rand"
	"strconv"
	"testing"
)

func TestLRUCache(t *testing.T) {
	cache := NewCache(2)
	cache.Put("1","1")
	cache.Put("2", "2")
	cache.Put("3","3")
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

func BenchmarkLRUCache(b *testing.B) {
	cache := NewCache(100)
	for i:=0; i < b.N; i++ {
		key := strconv.Itoa(rand.Int())
		value := strconv.Itoa(rand.Int())
		cache.Put(key, value)
	}

	for i:=0; i < b.N; i++ {
		key := strconv.Itoa(rand.Int())
		cache.Get(key)
	}
}

