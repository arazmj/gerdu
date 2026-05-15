package httpserver

import (
	"context"
	"errors"
	"github.com/arazmj/gerdu/lrucache"
	"github.com/arazmj/gerdu/raftproxy"
	"github.com/gorilla/mux"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type errorReader struct {
	sent bool
}

func (r *errorReader) Read(p []byte) (int, error) {
	if r.sent {
		return 0, io.ErrUnexpectedEOF
	}
	r.sent = true
	copy(p, "partial")
	return len("partial"), nil
}

func requestWithKey(method, target, key, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	return mux.SetURLVars(req, map[string]string{"key": key})
}

func assertStatus(t *testing.T, rr *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rr.Code != want {
		t.Fatalf("expected status %d, got %d", want, rr.Code)
	}
}

func TestPutGetDeleteHandlers(t *testing.T) {
	gerdu := lrucache.NewCache(100)

	put := httptest.NewRecorder()
	putHandler(put, requestWithKey(http.MethodPut, "/cache/name", "name", "gerdu"), gerdu)
	if put.Code != http.StatusCreated && put.Code != http.StatusOK {
		t.Fatalf("expected create/update success, got %d", put.Code)
	}

	get := httptest.NewRecorder()
	getHandler(get, requestWithKey(http.MethodGet, "/cache/name", "name", ""), gerdu)
	assertStatus(t, get, http.StatusOK)
	if body := get.Body.String(); body != "gerdu" {
		t.Fatalf("expected cached value %q, got %q", "gerdu", body)
	}

	miss := httptest.NewRecorder()
	getHandler(miss, requestWithKey(http.MethodGet, "/cache/missing", "missing", ""), gerdu)
	assertStatus(t, miss, http.StatusNotFound)

	deleted := httptest.NewRecorder()
	deleteHandler(deleted, requestWithKey(http.MethodDelete, "/cache/name", "name", ""), gerdu)
	assertStatus(t, deleted, http.StatusOK)

	deleteMiss := httptest.NewRecorder()
	deleteHandler(deleteMiss, requestWithKey(http.MethodDelete, "/cache/name", "name", ""), gerdu)
	assertStatus(t, deleteMiss, http.StatusNotFound)
}

func TestPutHandlerBadBodyKeepsServerAlive(t *testing.T) {
	gerdu := lrucache.NewCache(100)

	badReq := mux.SetURLVars(httptest.NewRequest(http.MethodPut, "/cache/bad", &errorReader{}), map[string]string{"key": "bad"})
	bad := httptest.NewRecorder()
	putHandler(bad, badReq, gerdu)
	assertStatus(t, bad, http.StatusBadRequest)
	if _, ok := gerdu.Get("bad"); ok {
		t.Fatal("bad body should not be stored in cache")
	}

	good := httptest.NewRecorder()
	putHandler(good, requestWithKey(http.MethodPut, "/cache/good", "good", "ok"), gerdu)
	assertStatus(t, good, http.StatusCreated)
}

func TestPutHandlerTTLHeader(t *testing.T) {
	gerdu := lrucache.NewCache(100)
	req := requestWithKey(http.MethodPut, "/cache/ttl", "ttl", "short-lived")
	req.Header.Set("X-TTL", "1")
	put := httptest.NewRecorder()
	putHandler(put, req, gerdu)
	assertStatus(t, put, http.StatusCreated)

	immediate := httptest.NewRecorder()
	getHandler(immediate, requestWithKey(http.MethodGet, "/cache/ttl", "ttl", ""), gerdu)
	assertStatus(t, immediate, http.StatusOK)

	time.Sleep(1100 * time.Millisecond)
	expired := httptest.NewRecorder()
	getHandler(expired, requestWithKey(http.MethodGet, "/cache/ttl", "ttl", ""), gerdu)
	assertStatus(t, expired, http.StatusNotFound)
}

func TestPutHandlerInvalidTTLHeader(t *testing.T) {
	for _, ttl := range []string{"abc", "-5"} {
		t.Run(ttl, func(t *testing.T) {
			gerdu := lrucache.NewCache(100)
			req := requestWithKey(http.MethodPut, "/cache/badttl", "badttl", "value")
			req.Header.Set("X-TTL", ttl)
			rr := httptest.NewRecorder()
			putHandler(rr, req, gerdu)
			assertStatus(t, rr, http.StatusBadRequest)
			if _, ok := gerdu.Get("badttl"); ok {
				t.Fatal("invalid TTL should not store a value")
			}
		})
	}
}

func TestRouterMethodNotAllowed(t *testing.T) {
	gerdu := lrucache.NewCache(100)
	router := newRouter(gerdu)

	for _, tc := range []struct {
		name   string
		method string
		path   string
	}{
		{"post cache", http.MethodPost, "/cache/key"},
		{"get join", http.MethodGet, "/join"},
		{"get leave", http.MethodGet, "/leave"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, httptest.NewRequest(tc.method, tc.path, nil))
			assertStatus(t, rr, http.StatusMethodNotAllowed)
		})
	}
}

func TestJoinAndLeaveBadRequests(t *testing.T) {
	raftCache := raftproxy.NewRaftProxy(lrucache.NewCache(100), "127.0.0.1:0", "", "node1")

	cases := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request, *raftproxy.RaftProxy)
		req     *http.Request
	}{
		{
			name: "join bad json",
			handler: func(w http.ResponseWriter, r *http.Request, c *raftproxy.RaftProxy) {
				joinHandler(w, r, c)
			},
			req: httptest.NewRequest(http.MethodPost, "/join", strings.NewReader("{")),
		},
		{
			name: "join missing addr",
			handler: func(w http.ResponseWriter, r *http.Request, c *raftproxy.RaftProxy) {
				joinHandler(w, r, c)
			},
			req: httptest.NewRequest(http.MethodPost, "/join", strings.NewReader(`{"id":"node2","extra":"value"}`)),
		},
		{
			name: "join missing id",
			handler: func(w http.ResponseWriter, r *http.Request, c *raftproxy.RaftProxy) {
				joinHandler(w, r, c)
			},
			req: httptest.NewRequest(http.MethodPost, "/join", strings.NewReader(`{"addr":"127.0.0.1:1","extra":"value"}`)),
		},
		{
			name: "leave empty node id",
			handler: func(w http.ResponseWriter, r *http.Request, c *raftproxy.RaftProxy) {
				leaveHandler(w, r, c)
			},
			req: httptest.NewRequest(http.MethodPost, "/leave", strings.NewReader("")),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			tc.handler(rr, tc.req, raftCache)
			assertStatus(t, rr, http.StatusBadRequest)
		})
	}
}

func TestServerShutdown(t *testing.T) {
	gerdu := lrucache.NewCache(100)
	server := newServer("127.0.0.1:0", gerdu)
	listener, err := net.Listen("tcp", server.server.Addr)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	serveDone := make(chan error, 1)
	go func() {
		serveDone <- server.server.Serve(listener)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	select {
	case err := <-serveDone:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("unexpected serve error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop after shutdown")
	}
}

func TestHTTPServeShutdown(t *testing.T) {
	server := HTTPServe("127.0.0.1:0", lrucache.NewCache(100))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown HTTPServe server: %v", err)
	}
}

func TestIndexHandler(t *testing.T) {
	gerdu := lrucache.NewCache(4)
	tests := []struct {
		name             string
		r                *http.Request
		w                *httptest.ResponseRecorder
		expectedStatus   int
		expectedResponse string
	}{
		{name: "Put 1:1", r: httptest.NewRequest(http.MethodPut, "/cache/1", strings.NewReader("1")), w: httptest.NewRecorder(), expectedStatus: http.StatusCreated},
		{name: "Put 2:2", r: httptest.NewRequest(http.MethodPut, "/cache/2", strings.NewReader("2")), w: httptest.NewRecorder(), expectedStatus: http.StatusCreated},
		{name: "Put 3:3", r: httptest.NewRequest(http.MethodPut, "/cache/3", strings.NewReader("3")), w: httptest.NewRecorder(), expectedStatus: http.StatusCreated},
		{name: "Get 2:2", r: httptest.NewRequest(http.MethodGet, "/cache/2", nil), w: httptest.NewRecorder(), expectedResponse: "2", expectedStatus: http.StatusOK},
		{name: "Get 3:3", r: httptest.NewRequest(http.MethodGet, "/cache/3", nil), w: httptest.NewRecorder(), expectedResponse: "3", expectedStatus: http.StatusOK},
		{name: "Get 1:1", r: httptest.NewRequest(http.MethodGet, "/cache/1", nil), w: httptest.NewRecorder(), expectedStatus: http.StatusNotFound},
		{name: "Delete 3:3", r: httptest.NewRequest(http.MethodDelete, "/cache/3", nil), w: httptest.NewRecorder(), expectedStatus: http.StatusOK},
		{name: "Get 3:3", r: httptest.NewRequest(http.MethodGet, "/cache/3", nil), w: httptest.NewRecorder(), expectedStatus: http.StatusNotFound},
		{name: "Delete 3:3", r: httptest.NewRequest(http.MethodDelete, "/cache/3", nil), w: httptest.NewRecorder(), expectedStatus: http.StatusNotFound},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			router := mux.NewRouter()
			router.HandleFunc("/cache/{key}", func(w http.ResponseWriter, r *http.Request) {
				switch test.r.Method {
				case http.MethodPut:
					putHandler(w, r, gerdu)
				case http.MethodGet:
					getHandler(w, r, gerdu)
				case http.MethodDelete:
					deleteHandler(w, r, gerdu)
				}
			})
			router.ServeHTTP(test.w, test.r)
			if test.w.Code != test.expectedStatus {
				t.Errorf("Failed to produce expected status code %d, got %d", test.expectedStatus, test.w.Code)
			}
		})
	}
}
