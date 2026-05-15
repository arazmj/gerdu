package redis

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type testCache struct {
	mu     sync.Mutex
	items  map[string]string
	timers map[string]*time.Timer
}

func newTestCache() *testCache {
	return &testCache{
		items:  make(map[string]string),
		timers: make(map[string]*time.Timer),
	}
}

func (c *testCache) Put(key string, value string) (created bool) {
	return c.PutWithTTL(key, value, 0)
}

func (c *testCache) PutWithTTL(key string, value string, ttl time.Duration) (created bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if timer, ok := c.timers[key]; ok {
		timer.Stop()
		delete(c.timers, key)
	}
	_, exists := c.items[key]
	c.items[key] = value
	if ttl > 0 {
		c.timers[key] = time.AfterFunc(ttl, func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			delete(c.items, key)
			delete(c.timers, key)
		})
	}
	return !exists
}

func (c *testCache) Get(key string) (value string, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok = c.items[key]
	return value, ok
}

func (c *testCache) Delete(key string) (ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if timer, ok := c.timers[key]; ok {
		timer.Stop()
		delete(c.timers, key)
	}
	_, ok = c.items[key]
	delete(c.items, key)
	return ok
}

type respClient struct {
	conn net.Conn
	r    *bufio.Reader
}

func newRESPClient(t *testing.T, addr string) *respClient {
	t.Helper()
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("dial redis server: %v", err)
	}
	return &respClient{conn: conn, r: bufio.NewReader(conn)}
}

func (c *respClient) close() {
	_ = c.conn.Close()
}

func (c *respClient) do(t *testing.T, args ...string) interface{} {
	t.Helper()
	if _, err := fmt.Fprintf(c.conn, "*%d\r\n", len(args)); err != nil {
		t.Fatalf("write command header: %v", err)
	}
	for _, arg := range args {
		if _, err := fmt.Fprintf(c.conn, "$%d\r\n%s\r\n", len(arg), arg); err != nil {
			t.Fatalf("write command arg: %v", err)
		}
	}
	resp, err := c.readValue()
	if err != nil {
		t.Fatalf("read response for %v: %v", args, err)
	}
	return resp
}

func (c *respClient) readValue() (interface{}, error) {
	line, err := c.r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
	if line == "" {
		return nil, fmt.Errorf("empty RESP line")
	}

	switch line[0] {
	case '+':
		return line[1:], nil
	case '-':
		return respError(line[1:]), nil
	case ':':
		return strconv.Atoi(line[1:])
	case '$':
		length, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, err
		}
		if length == -1 {
			return nil, nil
		}
		buf := make([]byte, length+2)
		if _, err := io.ReadFull(c.r, buf); err != nil {
			return nil, err
		}
		return string(buf[:length]), nil
	case '*':
		count, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, err
		}
		values := make([]interface{}, 0, count)
		for i := 0; i < count; i++ {
			value, err := c.readValue()
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		return values, nil
	default:
		return nil, fmt.Errorf("unknown RESP prefix %q", line[0])
	}
}

type respError string

func startRedisServer(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve redis port: %v", err)
	}
	addr := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("close reserved redis port: %v", err)
	}

	go Serve(addr, newTestCache())

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return addr
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("redis server did not start on %s", addr)
	return ""
}

func TestRedisProtocolCommands(t *testing.T) {
	client := newRESPClient(t, startRedisServer(t))
	defer client.close()

	if got := client.do(t, "PING"); got != "PONG" {
		t.Fatalf("PING = %v, want PONG", got)
	}
	if got := client.do(t, "COMMAND"); len(got.([]interface{})) != 0 {
		t.Fatalf("COMMAND = %v, want empty array", got)
	}
	if got := client.do(t, "GET", "missing"); got != nil {
		t.Fatalf("GET missing = %v, want nil", got)
	}
	if got := client.do(t, "SET", "alpha", "one"); got != "OK" {
		t.Fatalf("SET = %v, want OK", got)
	}
	if got := client.do(t, "GET", "alpha"); got != "one" {
		t.Fatalf("GET alpha = %v, want one", got)
	}
	if got := client.do(t, "DEL", "alpha"); got != 1 {
		t.Fatalf("DEL alpha = %v, want 1", got)
	}
	if got := client.do(t, "DEL", "alpha"); got != 0 {
		t.Fatalf("DEL missing alpha = %v, want 0", got)
	}
	if got := client.do(t, "NOPE"); !strings.Contains(string(got.(respError)), "unknown command") {
		t.Fatalf("unknown command = %v, want unknown command error", got)
	}
}

func TestRedisTTLCommands(t *testing.T) {
	client := newRESPClient(t, startRedisServer(t))
	defer client.close()

	if got := client.do(t, "SETEX", "short", "1", "value"); got != "OK" {
		t.Fatalf("SETEX = %v, want OK", got)
	}
	if got := client.do(t, "GET", "short"); got != "value" {
		t.Fatalf("GET short before expiry = %v, want value", got)
	}
	time.Sleep(1100 * time.Millisecond)
	if got := client.do(t, "GET", "short"); got != nil {
		t.Fatalf("GET short after expiry = %v, want nil", got)
	}

	if got := client.do(t, "SET", "with-ex", "value", "EX", "1"); got != "OK" {
		t.Fatalf("SET EX = %v, want OK", got)
	}
	if got := client.do(t, "GET", "with-ex"); got != "value" {
		t.Fatalf("GET with-ex before expiry = %v, want value", got)
	}
	time.Sleep(1100 * time.Millisecond)
	if got := client.do(t, "GET", "with-ex"); got != nil {
		t.Fatalf("GET with-ex after expiry = %v, want nil", got)
	}
}

func TestRedisCommandErrors(t *testing.T) {
	client := newRESPClient(t, startRedisServer(t))
	defer client.close()

	cases := []struct {
		name string
		args []string
		want string
	}{
		{name: "set wrong arity", args: []string{"SET", "only-key"}, want: "wrong number of arguments"},
		{name: "set bad option", args: []string{"SET", "key", "value", "PX", "1"}, want: "syntax error"},
		{name: "set invalid ttl", args: []string{"SET", "key", "value", "EX", "0"}, want: "invalid expire time"},
		{name: "setex wrong arity", args: []string{"SETEX", "key", "1"}, want: "wrong number of arguments"},
		{name: "setex invalid ttl", args: []string{"SETEX", "key", "0", "value"}, want: "invalid expire time"},
		{name: "get wrong arity", args: []string{"GET"}, want: "wrong number of arguments"},
		{name: "del wrong arity", args: []string{"DEL"}, want: "wrong number of arguments"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := client.do(t, tc.args...).(respError)
			if !ok || !strings.Contains(string(got), tc.want) {
				t.Fatalf("%v = %v, want error containing %q", tc.args, got, tc.want)
			}
		})
	}
}

func TestParseTTLSeconds(t *testing.T) {
	if ttl, ok := parseTTLSeconds([]byte("2")); !ok || ttl != 2*time.Second {
		t.Fatalf("parseTTLSeconds valid = %v %v, want 2s true", ttl, ok)
	}
	for _, raw := range [][]byte{[]byte("0"), []byte("-1"), []byte("nan")} {
		if ttl, ok := parseTTLSeconds(raw); ok || ttl != 0 {
			t.Fatalf("parseTTLSeconds(%q) = %v %v, want 0 false", raw, ttl, ok)
		}
	}
}
