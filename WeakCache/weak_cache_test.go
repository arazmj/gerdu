package WeakCache

import (
	"runtime"
	"strconv"
	"testing"
)

func TestWeakCache(t *testing.T) {
	cache := NewWeakCache()

	c := 500
	for i := 0; i < c; i++ {
		itoa := strconv.Itoa(i)
		cache.Put(itoa, itoa)
	}

	for i := 0; i < c; i++ {
		itoa := strconv.Itoa(i)
		_, ok := cache.Get(itoa)
		if !ok {
			t.Errorf("Expected %d before GC", i)
		}
	}

	// ouch
	for i := 1; i < 10; i++ {
		runtime.Gosched()
		runtime.GC()
	}

	for i := 0; i < c; i++ {
		itoa := strconv.Itoa(i)
		_, ok := cache.Get(itoa)
		if ok {
			t.Errorf("Not expected %d after GC", i)
		}
	}

}
