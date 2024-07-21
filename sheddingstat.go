package loadshedder

import (
	"github.com/anandshukla-sharechat/loadshedder/stat"
	"github.com/prometheus/client_golang/prometheus"
	"sync/atomic"
	"time"
)

const (
	totalLabel = "total"
	dropLabel  = "dropped"
)

type (
	// A SheddingStat is used to store the statistics for load shedding.
	SheddingStat struct {
		total           int64
		pass            int64
		drop            int64
		sheddingMetrics *prometheus.CounterVec
		cpuMetric       *prometheus.GaugeVec
	}

	snapshot struct {
		Total int64
		Pass  int64
		Drop  int64
	}
)

// NewSheddingStat returns a SheddingStat.
func NewSheddingStat(loadSheddingMetrics *prometheus.CounterVec, cpuMetrics *prometheus.GaugeVec) *SheddingStat {
	st := &SheddingStat{
		sheddingMetrics: loadSheddingMetrics,
		cpuMetric:       cpuMetrics,
	}
	go st.run()
	go stat.NewUsageStat(st.cpuMetric)
	return st
}

// IncrementTotal increments the total requests.
func (s *SheddingStat) IncrementTotal() {
	atomic.AddInt64(&s.total, 1)
}

// IncrementPass increments the passed requests.
func (s *SheddingStat) IncrementPass() {
	atomic.AddInt64(&s.pass, 1)
}

// IncrementDrop increments the dropped requests.
func (s *SheddingStat) IncrementDrop() {
	atomic.AddInt64(&s.drop, 1)
}

func (s *SheddingStat) loop(c <-chan time.Time) {
	for range c {
		loadStat := s.reset()
		if s.sheddingMetrics != nil {
			s.sheddingMetrics.WithLabelValues(totalLabel).Add(float64(loadStat.Total))
			s.sheddingMetrics.WithLabelValues(dropLabel).Add(float64(loadStat.Drop))
		}
	}
}

func (s *SheddingStat) reset() snapshot {
	return snapshot{
		Total: atomic.SwapInt64(&s.total, 0),
		Pass:  atomic.SwapInt64(&s.pass, 0),
		Drop:  atomic.SwapInt64(&s.drop, 0),
	}
}

func (s *SheddingStat) run() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	s.loop(ticker.C)
}
