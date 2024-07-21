package stat

import (
	"errors"
	"fmt"
	"github.com/anandshukla-sharechat/loadshedder/utils"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

const (
	cpuTicks           = 100
	cpuFields          = 8
	defaultCfsPeriodUs = 100000
)

var (
	preSystem uint64
	preTotal  uint64
	limit     float64
	request   int64
	cores     uint64
	initOnce  sync.Once
)

// if /proc not present, ignore the cpu calculation, like wsl linux
func initialize() {
	cpus, err := cpuSets()
	if err != nil {
		zap.S().Error(err)
		return
	}

	cores = uint64(len(cpus))
	zap.S().Infof("#CPU on node: %v", cores)

	requestInt, err := cpuShares()
	if err != nil {
		zap.S().Error(err)
		return
	}
	request = requestInt / 1e3
	zap.S().Infof("CPU request for cgroup: %v", request)

	limit, err = cpuLimit()
	if err != nil {
		zap.S().Error(err)
		return
	}
	zap.S().Infof("CPU limit for cgroup: %v", limit)

	preSystem, err = systemCpuUsage()
	if err != nil {
		zap.S().Error(err)
		return
	}

	preTotal, err = totalCpuUsage()
	if err != nil {
		zap.S().Error(err)
		return
	}
}

// RefreshCpu refreshes cpu usage and returns.
func RefreshCpu(metrics *prometheus.GaugeVec) uint64 {
	initOnce.Do(initialize)

	total, err := totalCpuUsage()
	if err != nil {
		return 0
	}

	system, err := systemCpuUsage()
	if err != nil {
		return 0
	}

	var usage uint64
	cpuDelta := total - preTotal
	systemDelta := system - preSystem
	if cpuDelta > 0 && systemDelta > 0 {
		usage = uint64(float64(cpuDelta*cores*1e2) / (float64(systemDelta) * float64(request)))
	}
	preSystem = system
	preTotal = total

	return usage
}

func cpuShares() (int64, error) {
	cg, err := currentCgroup()
	if err != nil {
		return 0, err
	}

	return cg.cpuShares()
}

func cpuLimit() (float64, error) {
	cq, err := cpuQuota()
	if err == nil {
		if cq != -1 {
			period, err := cpuPeriod()
			if err != nil {
				zap.S().Error(err)
				return float64(cq) / float64(defaultCfsPeriodUs), err
			}
			return float64(cq) / float64(period), nil
		}
	}
	return -1, fmt.Errorf("error while calculating cpu limit")
}

func cpuQuota() (int64, error) {
	cg, err := currentCgroup()
	if err != nil {
		return 0, err
	}

	return cg.cpuQuotaUs()
}

func cpuPeriod() (uint64, error) {
	cg, err := currentCgroup()
	if err != nil {
		return 0, err
	}

	return cg.cpuPeriodUs()
}

func cpuSets() ([]uint64, error) {
	cg, err := currentCgroup()
	if err != nil {
		return nil, err
	}

	return cg.cpus()
}

func systemCpuUsage() (uint64, error) {
	lines, err := utils.ReadTextLines("/proc/stat", utils.WithoutBlank())
	if err != nil {
		return 0, err
	}

	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			if len(fields) < cpuFields {
				return 0, fmt.Errorf("bad format of cpu stats")
			}

			var totalClockTicks uint64
			for _, i := range fields[1:cpuFields] {
				v, err := parseUint(i)
				if err != nil {
					return 0, err
				}

				totalClockTicks += v
			}

			return (totalClockTicks * uint64(time.Second)) / cpuTicks, nil
		}
	}

	return 0, errors.New("bad stats format")
}

func totalCpuUsage() (usage uint64, err error) {
	var cg cgroup
	if cg, err = currentCgroup(); err != nil {
		return
	}

	return cg.usageAllCpus()
}
