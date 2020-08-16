package main

import (
	"context"
	"github.com/arazmj/gerdu/lrucache"
	"github.com/arazmj/gerdu/proto"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	proto.RegisterGerduServer(s, &server{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestIndexHandler(t *testing.T) {
	gerdu = lrucache.NewCache(2)
	tests := []struct {
		name             string
		r                *http.Request
		w                *httptest.ResponseRecorder
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:           "Put 1:1",
			r:              httptest.NewRequest("PUT", "/cache/1", strings.NewReader("1")),
			w:              httptest.NewRecorder(),
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Put 2:2",
			r:              httptest.NewRequest("PUT", "/cache/2", strings.NewReader("2")),
			w:              httptest.NewRecorder(),
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Put 3:3",
			r:              httptest.NewRequest("PUT", "/cache/3", strings.NewReader("3")),
			w:              httptest.NewRecorder(),
			expectedStatus: http.StatusCreated,
		},
		{
			name:             "Get 2:2",
			r:                httptest.NewRequest("GET", "/cache/2", nil),
			w:                httptest.NewRecorder(),
			expectedResponse: "2",
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "Get 3:3",
			r:                httptest.NewRequest("GET", "/cache/3", nil),
			w:                httptest.NewRecorder(),
			expectedResponse: "3",
			expectedStatus:   http.StatusOK,
		},
		{
			name:           "Get 1:1",
			r:              httptest.NewRequest("GET", "/cache/1", nil),
			w:              httptest.NewRecorder(),
			expectedStatus: http.StatusNotFound,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if strings.HasPrefix(test.name, "Put") {
				router := mux.NewRouter()
				router.HandleFunc("/cache/{key}", putHandler)
				router.ServeHTTP(test.w, test.r)

				if test.w.Code != test.expectedStatus {
					t.Errorf("Failed to produce expected status code %d, got %d", test.expectedStatus, test.w.Code)
				}
			} else {
				router := mux.NewRouter()
				router.HandleFunc("/cache/{key}", getHandler)
				router.ServeHTTP(test.w, test.r)

				if test.w.Code != test.expectedStatus || test.expectedResponse != test.w.Body.String() {
					t.Errorf("Failed to produce expected result %d, %s, got %d, %s",
						test.expectedStatus, test.expectedResponse, test.w.Code, test.w.Body.String())
				}

			}
		})
	}
}

func TestServerGrpc(t *testing.T) {
	gerdu = lrucache.NewCache(1000)
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := proto.NewGerduClient(conn)
	resp, err := client.Put(ctx, &proto.PutRequest{
		Key:   "Key1",
		Value: []byte("Value1"),
	})
	if err != nil {
		t.Fatalf("gRPC Put failed: %v", err)
	}

	if resp.Created != true {
		t.Fatalf("gRPC Put could not create the key")
	}

	response, err := client.Get(ctx, &proto.GetRequest{
		Key: "Key1",
	})

	if err != nil {
		t.Fatalf("gRPC Get failed: %v", err)
	}

	value := string(response.Value)
	if value != "Value1" {
		t.Fatalf("gRPC Get the value does not match expecting Value1, but got %v", value)
	}
}
