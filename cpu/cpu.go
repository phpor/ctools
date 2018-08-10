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
	"github.com/davecgh/go-spew/spew"
	"os"
)

const  UNLIMITTED  = 9223372036854771712


var lastCpuState *CpuStat
var cpu_tick = uint64(100)
const nanoSecondsPerSecond = 1e9


type CpuStat struct {
	Total uint64
	Usage uint64
	Usage_user uint64
	Usage_system uint64
	Cpu_num float64
	Restricted_cpu_num float64

}

func debug(val ...interface{}) {
 	if os.Getenv("debug_ctools") == "on" {
 		spew.Dump(val...)
	}
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

	cgroup_cpu_dir := utils.GetCgroupDir("cpu")
	cpustat := &CpuStat{}
	cpustat.Total, _ = getSystemCPUUsage()

	var err error
	var quota_us float64
	var period_us uint64
	if _quota_us, err := ioutil.ReadFile(cgroup_cpu_dir + "/cpu.cfs_quota_us"); err != nil {
		return nil, err
	} else {
		quota_us,err = strconv.ParseFloat(strings.Trim(string(_quota_us), "\n"), 64)
	}

	if period_us,err = utils.ReadUint64(cgroup_cpu_dir + "/cpu.cfs_period_us");err != nil {
		return nil, err
	}

	cgroup_cpuacct_dir := utils.GetCgroupDir("cpuacct")


	if cpustat.Usage,err = utils.ReadUint64(cgroup_cpuacct_dir + "/cpuacct.usage");err != nil {
		return nil, err
	}

	var cpu_num float64
	if usage_percpu, err := ioutil.ReadFile(cgroup_cpuacct_dir + "/cpuacct.usage_percpu"); err != nil {
		return nil, err
	} else {
		cpu_num = float64(len(strings.Split(strings.Trim(string(usage_percpu), "\n "), " ")))
	}

	cpustat.Cpu_num = cpu_num
	if quota_us == -1 {
		cpustat.Restricted_cpu_num = cpu_num
	} else {
		cpustat.Restricted_cpu_num = quota_us / float64(period_us)

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



func GetCpuUsage() float64 {
	preCpuState, err := getCgroupCpuUsage()
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Millisecond * 100)
	nextCpuState, _ := getCgroupCpuUsage()
	debug(preCpuState)
	debug(nextCpuState)
	totalUsage := nextCpuState.Usage - preCpuState.Usage
	total := nextCpuState.Total - preCpuState.Total
	debug(fmt.Sprintf("%d/((%d/%f)*%f)",totalUsage, total, nextCpuState.Cpu_num, nextCpuState.Restricted_cpu_num))
	return float64(totalUsage)/((float64(total)/nextCpuState.Cpu_num)*float64(nextCpuState.Restricted_cpu_num))
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
	preCpuState := lastCpuState
	nextCpuState, _ := getCgroupCpuUsage()
	lastCpuState = nextCpuState
	debug(preCpuState)
	debug(nextCpuState)
	totalUsage := nextCpuState.Usage - preCpuState.Usage
	total := nextCpuState.Total - preCpuState.Total
	debug(fmt.Sprintf("%d/((%d/%f)*%f)",totalUsage, total, nextCpuState.Cpu_num, nextCpuState.Restricted_cpu_num))
	return float64(totalUsage)/((float64(total)/nextCpuState.Cpu_num)*float64(nextCpuState.Restricted_cpu_num))
}
