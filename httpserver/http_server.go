package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/arazmj/gerdu/cache"
	"github.com/arazmj/gerdu/raftproxy"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

func newRouter(gerdu cache.UnImplementedCache) (router *mux.Router) {
	router = mux.NewRouter()
	router.HandleFunc("/cache/{key}", func(w http.ResponseWriter, r *http.Request) {
		getHandler(w, r, gerdu)
	}).Methods(http.MethodGet)
	router.HandleFunc("/cache/{key}", func(w http.ResponseWriter, r *http.Request) {
		putHandler(w, r, gerdu)
	}).Methods(http.MethodPut)
	router.HandleFunc("/cache/{key}", func(w http.ResponseWriter, r *http.Request) {
		deleteHandler(w, r, gerdu)
	}).Methods(http.MethodDelete)
	router.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
		joinHandler(w, r, gerdu)
	}).Methods(http.MethodPost)
	router.HandleFunc("/leave", func(w http.ResponseWriter, r *http.Request) {
		leaveHandler(w, r, gerdu)
	}).Methods(http.MethodPost)
	router.Handle("/metrics", promhttp.Handler())
	return router
}

type Server struct {
	server *http.Server
}

func newServer(host string, gerdu cache.UnImplementedCache) *Server {
	return &Server{server: &http.Server{
		Addr:              host,
		Handler:           newRouter(gerdu),
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}}
}

//HTTPServe start http server in plain text
func HTTPServe(host string, gerdu cache.UnImplementedCache) *Server {
	server := newServer(host, gerdu)
	go func() {
		log.Infof("Gerdu started listening HTTP at %s\n", host)
		if err := server.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	return server
}

//HTTPServeTLS start HTTP server in secure mode
func HTTPServeTLS(host string, tlsCert, tlsKey string, gerdu cache.UnImplementedCache) *Server {
	server := newServer(host, gerdu)
	go func() {
		log.Printf("Gerdu started listening HTTPS TLS at %s\n", host)
		if err := server.server.ListenAndServeTLS(tlsCert, tlsKey); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	return server
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func putHandler(w http.ResponseWriter, r *http.Request, gerdu cache.UnImplementedCache) {
	vars := mux.Vars(r)
	key := vars["key"]
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Errorf("HTTP PUT: failed to read body: %v", err)
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	value := buf.String()
	ttl, err := ttlFromHeader(r.Header.Get("X-TTL"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	created := gerdu.PutWithTTL(key, value, ttl)
	if !created {
		log.Printf("HTTP UPDATE Key: %s Value: %s\n", key, value)
	} else {
		log.Printf("HTTP INSERT Key: %s Value: %s\n", key, value)
	}

	if created {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func ttlFromHeader(value string) (time.Duration, error) {
	if value == "" {
		return 0, nil
	}
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil || seconds < 0 {
		return 0, strconv.ErrSyntax
	}
	return time.Duration(seconds) * time.Second, nil
}

func getHandler(w http.ResponseWriter, r *http.Request, gerdu cache.UnImplementedCache) {
	vars := mux.Vars(r)
	key := vars["key"]
	if value, ok := gerdu.Get(key); ok {
		log.Printf("HTTP RETREIVED Key: %s Value: %s\n", key, value)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(value))
	} else {
		log.Printf("HTTP MISSED Key: %s \n", key)
		w.WriteHeader(http.StatusNotFound)
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request, gerdu cache.UnImplementedCache) {
	vars := mux.Vars(r)
	key := vars["key"]
	if ok := gerdu.Delete(key); ok {
		log.Printf("HTTP DELETED Key: %s\n", key)
		w.WriteHeader(http.StatusOK)
	} else {
		log.Printf("HTTP MISSED Key: %s \n", key)
		w.WriteHeader(http.StatusNotFound)
	}
}

func joinHandler(w http.ResponseWriter, r *http.Request, gerdu cache.UnImplementedCache) {
	raftCache := gerdu.(*raftproxy.RaftProxy)
	m := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(m) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	remoteAddr, ok := m["addr"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nodeID, ok := m["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := raftCache.Join(nodeID, remoteAddr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Infof("Node %s, remoteAddr %s joined", nodeID, remoteAddr)
}

func leaveHandler(w http.ResponseWriter, r *http.Request, gerdu cache.UnImplementedCache) {
	raftCache := gerdu.(*raftproxy.RaftProxy)
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Errorf("HTTP LEAVE: failed to read body: %v", err)
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	nodeId := buf.String()

	if nodeId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := raftCache.Leave(nodeId); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Infof("Node %s has left", nodeId)
}
