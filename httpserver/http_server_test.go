package httpserver

import (
	"github.com/arazmj/gerdu/lrucache"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestPutHandlerBadBody(t *testing.T) {
	gerdu := lrucache.NewCache(1)
	router := mux.NewRouter()
	router.HandleFunc("/cache/{key}", func(w http.ResponseWriter, r *http.Request) {
		putHandler(w, r, gerdu)
	}).Methods(http.MethodPut)

	req := httptest.NewRequest(http.MethodPut, "/cache/bad", errorReader{})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Failed to produce expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
	if _, ok := gerdu.Get("bad"); ok {
		t.Error("bad body should not be stored in cache")
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
		{
			name:           "Put 1:1",
			r:              httptest.NewRequest(http.MethodPut, "/cache/1", strings.NewReader("1")),
			w:              httptest.NewRecorder(),
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Put 2:2",
			r:              httptest.NewRequest(http.MethodPut, "/cache/2", strings.NewReader("2")),
			w:              httptest.NewRecorder(),
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Put 3:3",
			r:              httptest.NewRequest(http.MethodPut, "/cache/3", strings.NewReader("3")),
			w:              httptest.NewRecorder(),
			expectedStatus: http.StatusCreated,
		},
		{
			name:             "Get 2:2",
			r:                httptest.NewRequest(http.MethodGet, "/cache/2", nil),
			w:                httptest.NewRecorder(),
			expectedResponse: "2",
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "Get 3:3",
			r:                httptest.NewRequest(http.MethodGet, "/cache/3", nil),
			w:                httptest.NewRecorder(),
			expectedResponse: "3",
			expectedStatus:   http.StatusOK,
		},
		{
			name:           "Get 1:1",
			r:              httptest.NewRequest(http.MethodGet, "/cache/1", nil),
			w:              httptest.NewRecorder(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Delete 3:3",
			r:              httptest.NewRequest(http.MethodDelete, "/cache/3", nil),
			w:              httptest.NewRecorder(),
			expectedStatus: http.StatusOK,
		},
		{
			name:             "Get 3:3",
			r:                httptest.NewRequest(http.MethodGet, "/cache/3", nil),
			w:                httptest.NewRecorder(),
			expectedResponse: "3",
			expectedStatus:   http.StatusNotFound,
		},
		{
			name:             "Delete 3:3",
			r:                httptest.NewRequest(http.MethodDelete, "/cache/3", nil),
			w:                httptest.NewRecorder(),
			expectedResponse: "3",
			expectedStatus:   http.StatusNotFound,
		},
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
