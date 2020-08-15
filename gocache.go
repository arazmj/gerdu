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
	"github.com/inhies/go-bytesize"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var cache Cache.Cache
var verbose = flag.Bool("verbose", false, "verbose logging")

func main() {
	capacityStr := flag.String("capacity", "64MB",
		"The size of cache, once cache reached this capacity old values will evicted.\n"+
			"Specify a numerical value followed by one of the following units (not case sensitive)"+
			"\nK or KB: Kilobytes"+
			"\nM or MB: Megabytes"+
			"\nG or GB: Gigabytes"+
			"\nT or TB: Terabytes")
	port := flag.Int("port", 8080, "the server port number")
	kind := flag.String("type", "lru", "type of cache, lru or lfu, weak")
	flag.Parse()

	capacity, _ := bytesize.Parse(*capacityStr)

	stats := Stats.NewStats()
	if strings.ToLower(*kind) == "lru" {
		cache = LRUCache.NewCache(capacity, stats)
	} else if strings.ToLower(*kind) == "lfu" {
		cache = LFUCache.NewCache(capacity, stats)
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
	if *verbose {
		if !created {
			log.Printf("UPDATED Key: %s Value: %s\n", key, value)
		} else {
			log.Printf("INSERTED Key: %s Value: %s\n", key, value)
		}
	}
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
		if *verbose {
			log.Printf("RETREIVED Key: %s Value: %s\n", key, value)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(value))
	} else {
		if *verbose {
			log.Printf("MISSED Key: %s \n", value)
		}
		w.WriteHeader(http.StatusNotFound)
	}
}
