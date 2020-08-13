package main

import (
	"GoCache/Cache"
	"GoCache/LFUCache"
	"GoCache/LRUCache"
	"GoCache/Stats"
	"GoCache/WeakCache"
	"bytes"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var cache Cache.Cache

func main() {
	capacity := flag.Int("capacity", 100,
		"how big the cache will be, the old values will be evicted")
	port := flag.Int("port", 8080, "the server port number")
	kind := flag.String("type", "lru", "type of cache, lru or lfu, weak")
	flag.Parse()

	stats := Stats.NewStats()
	if strings.ToLower(*kind) == "lru" {
		cache = LRUCache.NewCache(*capacity, stats)
	} else if strings.ToLower(*kind) == "lfu" {
		cache = LFUCache.NewCache(*capacity, stats)
	} else if strings.ToLower(*kind) == "weak" {
		cache = WeakCache.NewWeakCache(stats)
	} else {
		fmt.Println("Invalid value for type")
		os.Exit(1)
	}
	router := mux.NewRouter()
	router.HandleFunc("/cache/{key}", GetHandler).Methods(http.MethodGet)
	router.HandleFunc("/cache/{key}", PutHandler).Methods(http.MethodPut)
	router.Handle("/metrics", promhttp.Handler())
	log.Printf("GoCache started listening on %d port\n", *port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), router))
}

func PutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	value := buf.String()
	created := cache.Put(key, value)
	if created {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	if value, ok := cache.Get(key); ok {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(value))
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
