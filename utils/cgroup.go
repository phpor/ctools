package utils

import "strings"




var cgroup_dir = "/sys/fs/cgroup"


func GetCgroupDir(subtype string) string{
	cgroup := "/proc/self/cgroup"
	dir := cgroup_dir + "/" + subtype
	tmp_dir := ""
	ForEachFile(cgroup, func(line string) (bool, error) {
		arr := strings.Split(line, ":")
		arrType := strings.Split(arr[1], ",")
		for _,v := range arrType {
			if v == subtype {
				tmp_dir = cgroup_dir + arr[2]
				return false, nil
			}
		}
		return true, nil
	})
	if ok, _ := PathExists(tmp_dir); ok {
		return tmp_dir
	}
	return dir
}

