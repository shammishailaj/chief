package system

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	log "github.com/sirupsen/logrus"
)

const (
	CPU_CORES_CPUINFO_ERROR  = -2
	LAVG_TREND_CPUINFO_ERROR = -1
	LAVG_TREND_ERROR         = 1
	LAVG_TREND_INCREASING    = 2
	LAVG_TREND_NORMAL        = 3
	LAVG_LOAD_LEVEL_YELLOW   = 70
)

func LoadAvg() (*load.AvgStat, error) {
	return load.Avg()
}

func CPUInfo() ([]cpu.InfoStat, error) {
	return cpu.Info()
}

func CPUCores() int32 {
	var cores int32 = 0
	cpuInfo, cpuInfoErr := CPUInfo()
	if cpuInfoErr != nil {
		log.Errorf("Error getting CPU info. %s", cpuInfoErr.Error())
		return CPU_CORES_CPUINFO_ERROR
	}

	for _, info := range cpuInfo {
		cores += info.Cores
	}
	return cores
}

func LoadAvgCheck() int {
	lavg, lavgErr := LoadAvg()

	if lavgErr != nil {
		log.Errorf("Error getting load average. %s", lavgErr.Error())
		return LAVG_TREND_ERROR
	}
	if lavg.Load1 > lavg.Load5 {
		cpuCores := CPUCores()
		log.Infof("LoadAvgCheck():: Found %d CPU Cores", cpuCores)
		if cpuCores != CPU_CORES_CPUINFO_ERROR {
			loadValueYellow := (LAVG_LOAD_LEVEL_YELLOW * float64(cpuCores)) / 100
			if lavg.Load1 >= loadValueYellow {
				return LAVG_TREND_INCREASING
			} else {
				return LAVG_TREND_NORMAL
			}
		} else {
			return LAVG_TREND_CPUINFO_ERROR
		}
	} else {
		return LAVG_TREND_NORMAL
	}
}

func LoadAvgCheckCPUCores(cpuCores int32) int {
	lavg, lavgErr := LoadAvg()

	if lavgErr != nil {
		log.Errorf("Error getting load average. %s", lavgErr.Error())
		return LAVG_TREND_ERROR
	}
	if lavg.Load1 > lavg.Load5 {
		if cpuCores != CPU_CORES_CPUINFO_ERROR {
			loadValueYellow := (LAVG_LOAD_LEVEL_YELLOW * float64(cpuCores)) / 100
			if lavg.Load1 >= loadValueYellow {
				return LAVG_TREND_INCREASING
			} else {
				return LAVG_TREND_NORMAL
			}
		} else {
			return LAVG_TREND_CPUINFO_ERROR
		}
	} else {
		return LAVG_TREND_NORMAL
	}
}
