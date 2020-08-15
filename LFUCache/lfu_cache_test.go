package LFUCache

import (
	"github.com/inhies/go-bytesize"
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
	cache := NewCache(size).(*LFUCache)
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
	cache := NewCache(10).(*LFUCache)
	cache.Put("20", "20")
	if cache.freq[1].Size != 1 {
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
	cache := NewCache(10).(*LFUCache)

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
