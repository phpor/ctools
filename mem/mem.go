package mem

import (
	"github.com/phpor/ctools/utils"
	"strings"
	"strconv"
)

type MemStat struct {
	Total uint64
	Used uint64
	Cached uint64
	Available uint64
}


type sysMemStat struct {
	Total     uint64
	Free      uint64
	Cached    uint64
	Available uint64
}

func Usage() (*MemStat, error) {
	cgroup_memory_dir := utils.GetCgroupDir("memory")

	memstat := &MemStat{}

	mem_total, _ := utils.ReadUint64(cgroup_memory_dir + "/memory.limit_in_bytes")
	mem_used, _ := utils.ReadUint64(cgroup_memory_dir + "/memory.usage_in_bytes")
	mem_cached := uint64(0)
	mem_free := uint64(0)
	mem_available := uint64(0)
	if mem_total == utils.UNLIMITTED {
		sys_memstat,err := getSystemMemStat()
		if err != nil {
			panic(err)
		}
		mem_total = sys_memstat.Total
		mem_free = sys_memstat.Free
		mem_available = sys_memstat.Available
	} else {
		err := utils.ForEachFile(cgroup_memory_dir + "/memory.stat", func(line string) (bool, error) {
			arr := strings.Split(line, " ")
			switch arr[0] {
			case "cache":
				mem_cached, _ = strconv.ParseUint(arr[1], 10, 64)
			}
			return true, nil
		})
		if err != nil {
			panic(err)
			return nil, err
		}
		mem_free = mem_total - mem_used
		mem_available = mem_free + mem_cached
	}
	memstat.Total = mem_total
	memstat.Cached = mem_cached
	memstat.Used = mem_used
	memstat.Available = mem_available
	return memstat, nil
}


func getSystemMemStat() (*sysMemStat, error) {
	memstat := &sysMemStat{}
	err := utils.ForEachFile("/proc/meminfo", func(line string)(bool, error){
		arr := strings.Split(line, " ")
		switch arr[0] {
		case "MemTotal:":
			val,_ := strconv.ParseUint(arr[1], 10, 64)
			memstat.Total = val * 1000
		case "MemFree:":
			val,_ := strconv.ParseUint(arr[1], 10, 64)
			memstat.Free = val * 1000
		case "MemAvailable:":
			val,_ := strconv.ParseUint(arr[1], 10, 64)
			memstat.Available = val * 1000

		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}
	return memstat, nil

}


