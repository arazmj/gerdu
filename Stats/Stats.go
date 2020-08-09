package Stats

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type StatUpdater interface {
	AddOps()
	MissOps()
	DeleteOps()
	HitOps()
}

type Stats struct {
	hits prometheus.Counter
	miss prometheus.Counter
	adds prometheus.Counter
	dels prometheus.Counter
}

func (s *Stats) AddOps() {
	s.adds.Inc()
}

func (s *Stats) MissOps() {
	s.miss.Inc()
}

func (s *Stats) DeleteOps() {
	s.dels.Inc()
}

func (s *Stats) HitOps() {
	s.hits.Inc()
}

func NewStats() *Stats {
	return &Stats{
		miss: promauto.NewCounter(prometheus.CounterOpts{
			Name: "go_cache_misses_total",
			Help: "The total number of missed cache hits",
		}),
		hits: promauto.NewCounter(prometheus.CounterOpts{
			Name: "go_cache_hits_total",
			Help: "The total number of cache hits",
		}),
		adds: promauto.NewCounter(prometheus.CounterOpts{
			Name: "go_cache_adds_total",
			Help: "The total number of new added nodes",
		}),
		dels: promauto.NewCounter(prometheus.CounterOpts{
			Name: "go_cache_deletes_total",
			Help: "The total number of deletes nodes",
		}),
	}
}
