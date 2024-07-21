package stat

import (
	"github.com/anandshukla-sharechat/loadshedder/utils"
	"github.com/prometheus/client_golang/prometheus"
	"sync/atomic"
	"time"
)

const (
	// 250ms and 0.95 as beta will count the average cpu load for past 5 seconds
	cpuRefreshInterval = time.Millisecond * 250
	allRefreshInterval = time.Second
	// moving average beta hyperparameter
	beta = 0.95
)

var (
	cpuUsage int64
)

type Usage struct {
	metrics *prometheus.GaugeVec
}

// NewUsageStat returns a SheddingStat.
func NewUsageStat(metrics *prometheus.GaugeVec) *Usage {
	usage := &Usage{
		metrics: metrics,
	}
	go usage.run()
	return usage
}

func (u *Usage) run() {
	cpuTicker := time.NewTicker(cpuRefreshInterval)
	defer cpuTicker.Stop()
	allTicker := time.NewTicker(allRefreshInterval)
	defer allTicker.Stop()

	for {
		select {
		case <-cpuTicker.C:
			utils.RunSafe(func() {
				curUsage := RefreshCpu(u.metrics)
				prevUsage := atomic.LoadInt64(&cpuUsage)
				// cpu = cpuᵗ⁻¹ * beta + cpuᵗ * (1 - beta)
				usage := int64(float64(prevUsage)*beta + float64(curUsage)*(1-beta))
				atomic.StoreInt64(&cpuUsage, usage)
			})
		case <-allTicker.C:
			u.publishUsage()
		}
	}
}

// CpuUsage returns current cpu usage.
func (u *Usage) CpuUsage() int64 {
	return atomic.LoadInt64(&cpuUsage)
}

func (u *Usage) publishUsage() {
	usage := u.CpuUsage()
	if usage < 0 {
		usage = 0
	}
	if u.metrics != nil {
		u.metrics.WithLabelValues("usage").Set(float64(usage))
	}
}

// CpuUsage returns current cpu usage.
func CpuUsage() int64 {
	return atomic.LoadInt64(&cpuUsage)
}
