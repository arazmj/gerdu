package Stats

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Statser interface {
	AddOps()
	MissOps()
	DeleteOps()
	HitOps()
}

type Stats struct {
	Hits prometheus.Counter
	Miss prometheus.Counter
	Adds prometheus.Counter
	Dels prometheus.Counter
}

func (s *Stats) AddOps() {
	s.Adds.Inc()
}

func (s *Stats) MissOps() {
	s.Miss.Inc()
}

func (s *Stats) DeleteOps() {
	s.Dels.Inc()
}

func (s *Stats) HitOps() {
	s.Hits.Inc()
}

func NewStats() *Stats {
	return &Stats{
		Miss: promauto.NewCounter(prometheus.CounterOpts{
			Name: "go_cache_misses_total",
			Help: "The total number of missed cache hits",
		}),
		Hits: promauto.NewCounter(prometheus.CounterOpts{
			Name: "go_cache_hits_total",
			Help: "The total number of cache hits",
		}),
		Adds: promauto.NewCounter(prometheus.CounterOpts{
			Name: "go_cache_adds_total",
			Help: "The total number of new added nodes",
		}),
		Dels: promauto.NewCounter(prometheus.CounterOpts{
			Name: "go_cache_deletes_total",
			Help: "The total number of deletes nodes",
		}),
	}
}
