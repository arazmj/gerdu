package httpserver

import (
	"bytes"
	"github.com/arazmj/gerdu/cache"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
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
	router.Handle("/metrics", promhttp.Handler())
	return router
}

func HttpServe(host string, gerdu cache.UnImplementedCache) {
	router := newRouter(gerdu)
	log.Printf("Gerdu started listening HTTP on %s\n", host)
	log.Fatal(http.ListenAndServe(host, router))
}

func HttpServeTLS(host string, tlsCert, tlsKey string, gerdu cache.UnImplementedCache) {
	router := newRouter(gerdu)
	log.Printf("Gerdu started listening HTTPS TLS on %s\n", host)
	log.Fatal(http.ListenAndServeTLS(host, tlsCert, tlsKey, router))
}

func putHandler(w http.ResponseWriter, r *http.Request, gerdu cache.UnImplementedCache) {
	vars := mux.Vars(r)
	key := vars["key"]
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	value := buf.String()

	created := gerdu.Put(key, value)
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
