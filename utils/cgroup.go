package utils

import "strings"

// 说明： 参考过 /proc/self/cgroup   /proc/self/mounts   /proc/self/mountinfo  ,似乎都不保证是正确的


const  UNLIMITTED  = 9223372036854771712


func GetCgroupDir(subtype string) string{
	cgroupDir := ""
	//cgroup /sys/fs/cgroup/cpu,cpuacct cgroup rw,nosuid,nodev,noexec,relatime,cpu,cpuacct 0 0
	err := ForEachFile("/proc/self/mounts", func(line string)(bool, error) {
		arr := strings.Split(line, " ")
		if arr[2] != "cgroup" {
			return true,nil
		}
		arr2 := strings.Split(arr[3], ",")
		for _,v := range arr2 {
			if v == subtype {
				// 可能出现一个subsystem的mount条目有多个的情况，而有的目录是不存在的
				if ok, _ := PathExists(arr[1]); ok {
					cgroupDir = arr[1]
					return false, nil
				}
			}
		}
		return true, nil
	})
	if err != nil {
		panic("No cgroup mounted")
	}

	cgroup := "/proc/self/cgroup"
	tmp_dir := ""
	ForEachFile(cgroup, func(line string) (bool, error) {
		arr := strings.Split(line, ":")
		arrType := strings.Split(arr[1], ",")
		for _,v := range arrType {
			if v == subtype {
				tmp_dir = cgroupDir + "/" + arr[2]  // 不知道这样写是不是对
				return false, nil
			}
		}
		return true, nil
	})
	if tmp_dir != "" {
		if ok, _ := PathExists(tmp_dir); ok {
			return tmp_dir
		}
	}
	if cgroupDir != "" {
		if ok, _ := PathExists(cgroupDir); ok {
			return cgroupDir
		}
	}


	panic("No cgroup mounted")
}

