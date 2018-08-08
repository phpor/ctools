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

const  UNLIMITTED  = 9223372036854771712


var cpu_tick = uint64(100)
const nanoSecondsPerSecond = 1e9


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


func getCgroupCpuUsage() (map[string]uint64, error){

	cgroup_cpu_dir := utils.GetCgroupDir("cpu")
	cpustat := map[string]uint64{}
	cpustat["total"], _ = getSystemCPUUsage()

	var err error
	var quota_us = ""
	if _quota_us, err := ioutil.ReadFile(cgroup_cpu_dir + "/cpu.cfs_quota_us"); err != nil {
		return nil, err
	} else {
		quota_us = strings.Trim(string(_quota_us), "\n")
	}

	if cpustat["period_us"],err = utils.ReadUint64(cgroup_cpu_dir + "/cpu.cfs_period_us");err != nil {
		return nil, err
	}

	cgroup_cpuacct_dir := utils.GetCgroupDir("cpuacct")


	if cpustat["total_usage"],err = utils.ReadUint64(cgroup_cpuacct_dir + "/cpuacct.usage");err != nil {
		return nil, err
	}

	if usage_percpu, err := ioutil.ReadFile(cgroup_cpuacct_dir + "/cpuacct.usage_percpu"); err != nil {
		return nil, err
	} else {
		cpustat["cpu_num"] = uint64(len(strings.Split(strings.Trim(string(usage_percpu), "\n "), " ")))
	}

	cpu_num := cpustat["cpu_num"]

	if quota_us == "-1" {
		cpustat["restricted_cpu_num"] = cpu_num
	} else {
		cpustat["restricted_cpu_num"] = cpustat["quota_us"] / cpustat["period_us"]

	}

	utils.ForEachFile(cgroup_cpuacct_dir + "/cpuacct.stat", func(line string)(bool, error) {
		arr := strings.Split(line, " ")
		if arr[0] == "user" {
			cpustat["usage_user"],_ = strconv.ParseUint(arr[1], 10, 64)
		} else
		if arr[0] == "system" {
			cpustat["usage_system"],_ = strconv.ParseUint(arr[1], 10, 64)
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
	//spew.Dump(preCpuState)
	//spew.Dump(nextCpuState)
	totalUsage := nextCpuState["total_usage"] - preCpuState["total_usage"]
	totalCpu := nextCpuState["total"] - preCpuState["total"]
	return float64(totalUsage)/float64(totalCpu)*float64(nextCpuState["cpu_num"])/float64(nextCpuState["restricted_cpu_num"])
}
