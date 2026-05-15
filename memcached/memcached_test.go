package memcached

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type testCache struct {
	mu     sync.Mutex
	values map[string]string
}

func newTestCache() *testCache {
	return &testCache{values: make(map[string]string)}
}

func (c *testCache) Put(key string, value string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, exists := c.values[key]
	c.values[key] = value
	return !exists
}

func (c *testCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.values[key]
	return value, ok
}

func (c *testCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.values[key]
	delete(c.values, key)
	return ok
}

func startMemcachedServer(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on random port: %v", err)
	}
	addr := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("close port probe listener: %v", err)
	}

	go Serve(addr, newTestCache())
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return addr
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("memcached server did not start on %s", addr)
	return ""
}

func rawMemcachedLine(t *testing.T, addr, command string) string {
	t.Helper()
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("dial memcached: %v", err)
	}
	defer conn.Close()
	if _, err := fmt.Fprintf(conn, "%s\r\n", command); err != nil {
		t.Fatalf("write %q: %v", command, err)
	}
	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatalf("read response for %q: %v", command, err)
	}
	return line
}

func TestMemcachedProtocolOperations(t *testing.T) {
	addr := startMemcachedServer(t)
	client := memcache.New(addr)

	if err := client.Set(&memcache.Item{Key: "foo", Value: []byte("bar")}); err != nil {
		t.Fatalf("set foo: %v", err)
	}
	item, err := client.Get("foo")
	if err != nil {
		t.Fatalf("get foo: %v", err)
	}
	if string(item.Value) != "bar" {
		t.Fatalf("get foo value = %q, want bar", item.Value)
	}

	if _, err := client.Get("missing"); err != memcache.ErrCacheMiss {
		t.Fatalf("missing key err = %v, want ErrCacheMiss", err)
	}

	if err := client.Delete("foo"); err != nil {
		t.Fatalf("delete foo: %v", err)
	}
	if _, err := client.Get("foo"); err != memcache.ErrCacheMiss {
		t.Fatalf("deleted key err = %v, want ErrCacheMiss", err)
	}

	if err := client.Set(&memcache.Item{Key: "flush-me", Value: []byte("value")}); err != nil {
		t.Fatalf("set flush-me: %v", err)
	}
	if err := client.FlushAll(); err != nil {
		t.Fatalf("flush_all: %v", err)
	}
	if _, err := client.Get("flush-me"); err != memcache.ErrCacheMiss {
		t.Fatalf("flushed key err = %v, want ErrCacheMiss", err)
	}

	if err := client.Set(&memcache.Item{Key: "ttl", Value: []byte("short-lived"), Expiration: 1}); err != nil {
		t.Fatalf("set ttl: %v", err)
	}
	time.Sleep(2500 * time.Millisecond)
	if _, err := client.Get("ttl"); err != memcache.ErrCacheMiss {
		t.Fatalf("expired key err = %v, want ErrCacheMiss", err)
	}

	version := strings.TrimSpace(rawMemcachedLine(t, addr, "version"))
	if version != "Gerdu VERSION 0.1" {
		t.Fatalf("version response = %q", version)
	}

	unknown := rawMemcachedLine(t, addr, "unknown")
	if !strings.Contains(unknown, "CLIENT_ERROR") || !strings.Contains(unknown, "unknown command") {
		t.Fatalf("unknown command response = %q", unknown)
	}
}
