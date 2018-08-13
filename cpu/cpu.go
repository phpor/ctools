package cpu

import (
	"os/exec"
	"strconv"
	"strings"


	"bytes"
	"fmt"
	"github.com/phpor/ctools/utils"
	"io/ioutil"
	"time"

)



var lastCpuState *CpuStat
var cpu_tick = uint64(100)
const nanoSecondsPerSecond = 1e9


type CpuStat struct {
	Total uint64
	Usage uint64
	Usage_user uint64
	Usage_system uint64

}


func init() {
	getconf, err := exec.LookPath("getconf")
	if err != nil {
		return
	}
	cmd := exec.Command(getconf, "CLK_TCK")
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return
	}

	i, err := strconv.ParseUint(strings.TrimSpace(out.String()), 10, 64)
	if err == nil {
		cpu_tick = uint64(i)
	}

}


func getCgroupCpuUsage() (*CpuStat, error){

	cpustat := &CpuStat{}
	cpustat.Total, _ = getSystemCPUUsage()

	var err error

	cgroup_cpuacct_dir := utils.GetCgroupDir("cpuacct")

	if cpustat.Usage,err = utils.ReadUint64(cgroup_cpuacct_dir + "/cpuacct.usage");err != nil {
		return nil, err
	}

	utils.ForEachFile(cgroup_cpuacct_dir + "/cpuacct.stat", func(line string)(bool, error) {
		arr := strings.Split(line, " ")
		if arr[0] == "user" {
			cpustat.Usage_user,_ = strconv.ParseUint(arr[1], 10, 64)
		} else
		if arr[0] == "system" {
			cpustat.Usage_system,_ = strconv.ParseUint(arr[1], 10, 64)
		}
		return true, nil
	})
	return cpustat, nil

}


func getPerCPUTotalUsage() (float64, error) {
	totalCpu ,err := getSystemCPUUsage()
	if err != nil {
		return 0, err
	}
	cpuNum, err := CountAllCPU()
	if err != nil {
		return 0, err
	}

	return float64(totalCpu)/cpuNum, err
}

// getSystemCPUUsage collect the total cpu resource of all cpu core
func getSystemCPUUsage() (uint64, error) {
	totalCpu := uint64(0)
	err := utils.ForEachFile("/proc/stat", func(line string)(bool, error){
		parts := strings.Fields(line)
		switch parts[0] {
		case "cpu":
			if len(parts) < 8 {
				return false, fmt.Errorf("invalid number of cpu fields")
			}
			var totalClockTicks uint64
			for _, i := range parts[1:8] {
				v, err := strconv.ParseUint(i, 10, 64)
				if err != nil {
					return false, fmt.Errorf("Unable to convert value %s to int: %s", i, err)
				}
				totalClockTicks += v
			}
			totalCpu = (totalClockTicks * nanoSecondsPerSecond) / cpu_tick
		}
		return true, nil
	})
	return totalCpu, err
}

func CountAllCPU() (float64, error){
	var cpuNum float64
	cgroup_cpuacct_dir := utils.GetCgroupDir("cpuacct")
	if usage_percpu, err := ioutil.ReadFile(cgroup_cpuacct_dir + "/cpuacct.usage_percpu"); err != nil {
		return 0, err
	} else {
		cpuNum = float64(len(strings.Split(strings.Trim(string(usage_percpu), "\n "), " ")))
	}
	return cpuNum, nil
}

func CountLimitedCPU() (float64, error) {

	cgroup_cpu_dir := utils.GetCgroupDir("cpu")
	var err error
	var quota_us float64
	var period_us uint64
	if _quota_us, err := ioutil.ReadFile(cgroup_cpu_dir + "/cpu.cfs_quota_us"); err != nil {
		return 0, err
	} else {
		quota_us,err = strconv.ParseFloat(strings.Trim(string(_quota_us), "\n"), 64)
	}

	if period_us,err = utils.ReadUint64(cgroup_cpu_dir + "/cpu.cfs_period_us");err != nil {
		return 0, err
	}


	if quota_us == -1 {
		return CountAllCPU()
	}

	return quota_us / float64(period_us), nil
}


func GetCpuUsage() float64 {
	cpuNum, _ := CountLimitedCPU()
	cpuAll, _ := CountAllCPU()
	preCpuState, err := getCgroupCpuUsage()
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Millisecond * 100)
	nextCpuState, _ := getCgroupCpuUsage()
	utils.Debug(preCpuState)
	utils.Debug(nextCpuState)
	totalUsage := nextCpuState.Usage - preCpuState.Usage
	total := nextCpuState.Total - preCpuState.Total
	return float64(totalUsage)/((float64(total)/cpuAll) * cpuNum)
}

func GetCpuUsageNoDelay() float64 {
	if lastCpuState == nil {
		var err error
		lastCpuState, err = getCgroupCpuUsage()
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Millisecond * 500)

	}
	cpuNum, _ := CountLimitedCPU()
	cpuAll, _ := CountAllCPU()

	preCpuState := lastCpuState
	nextCpuState, _ := getCgroupCpuUsage()
	lastCpuState = nextCpuState
	utils.Debug(preCpuState)
	utils.Debug(nextCpuState)
	totalUsage := nextCpuState.Usage - preCpuState.Usage
	total := nextCpuState.Total - preCpuState.Total
	return float64(totalUsage)/((float64(total)/cpuAll) * cpuNum)
}
