//go:build !linux
// +build !linux

package stat

import "github.com/prometheus/client_golang/prometheus"

// RefreshCpu returns cpu usage, always returns 0 on systems other than linux.
func RefreshCpu(metrics *prometheus.GaugeVec) uint64 {
	return 0
}
