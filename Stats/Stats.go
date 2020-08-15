package Stats

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	Miss = promauto.NewCounter(prometheus.CounterOpts{
		Name: "go_cache_misses_total",
		Help: "The total number of missed cache hits",
	})
	Hits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "go_cache_hits_total",
		Help: "The total number of cache hits",
	})
	Adds = promauto.NewCounter(prometheus.CounterOpts{
		Name: "go_cache_adds_total",
		Help: "The total number of new added nodes",
	})
	Deletes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "go_cache_deletes_total",
		Help: "The total number of deletes nodes",
	})
)
