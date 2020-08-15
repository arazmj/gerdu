package main

import (
	"GoCache/cache"
	"GoCache/lfucache"
	"GoCache/lrucache"
	"GoCache/weakcache"
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

var cache2 cache.ICache
var verbose = flag.Bool("verbose", false, "verbose logging")

func main() {
	capacityStr := flag.String("capacity", "64MB",
		"The size of icache, once icache reached this capacity old values will evicted.\n"+
			"Specify a numerical value followed by one of the following units (not case sensitive)"+
			"\nK or KB: Kilobytes"+
			"\nM or MB: Megabytes"+
			"\nG or GB: Gigabytes"+
			"\nT or TB: Terabytes")
	port := flag.Int("port", 8080, "the server port number")
	kind := flag.String("type", "lru", "type of icache, lru or lfu, weak")
	flag.Parse()

	capacity, _ := bytesize.Parse(*capacityStr)

	if strings.ToLower(*kind) == "lru" {
		cache2 = lrucache.NewCache(capacity)
	} else if strings.ToLower(*kind) == "lfu" {
		cache2 = lfucache.NewCache(capacity)
	} else if strings.ToLower(*kind) == "weak" {
		cache2 = weakcache.NewWeakCache()
	} else {
		fmt.Println("Invalid value for type")
		os.Exit(1)
	}
	router := mux.NewRouter()
	router.HandleFunc("/icache/{key}", getHandler).Methods(http.MethodGet)
	router.HandleFunc("/icache/{key}", putHandler).Methods(http.MethodPut)
	router.Handle("/metrics", promhttp.Handler())
	log.Printf("GoCache started listening on %d port\n", *port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), router))
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	value := buf.String()

	created := cache2.Put(key, value)
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

func getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	if value, ok := cache2.Get(key); ok {
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
