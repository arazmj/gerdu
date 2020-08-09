package LFUCache

import (
	"GoCache/Stats"
	"strconv"
	"sync"
	"testing"
)

func TestThreadSafety(t *testing.T) {
	capacity := 300
	cache := NewLFUCache(capacity, Stats.NewStats())
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
func TestNewLFUCache(t *testing.T) {
	c := 100
	cache := NewLFUCache(c, nil).(*LFUCache)
	for i := 0; i < c; i++ {
		itoa := strconv.Itoa(i)
		cache.Put(itoa, itoa)
	}

	for i := 0; i < c; i++ {
		itoa := strconv.Itoa(i)
		value, ok := cache.Get(itoa)
		if !ok || value != itoa {
			t.Errorf("Bad result on %s", itoa)
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

	_, ok = cache.Get("1")
	if ok {
		t.Errorf("0 needs to be evecited by now")
	}

	for i := c; i < c+c/2; i++ {
		itoa := strconv.Itoa(i)
		cache.Put(itoa, itoa)
		cache.Get(itoa)
		cache.Get(itoa)
	}

	for i := 0; i < c+c/2; i++ {
		itoa := strconv.Itoa(i)
		_, ok := cache.Get(itoa)
		if i < c/2 {
			if ok {
				t.Errorf("Expected %d to be evicted", i)
			}
		} else {
			if !ok {
				t.Errorf("Expected %d to stay", i)
			}
		}
	}
}

func TestLFUCache_Update(t *testing.T) {
	cache := NewLFUCache(10, nil).(*LFUCache)
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
