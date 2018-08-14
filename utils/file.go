package utils

import (
	"os"
	"bufio"
	"strconv"
	"io/ioutil"
	"strings"
	"io"
)


func ReadString(filename string) (string, error) {
	if content, err := ioutil.ReadFile(filename); err != nil {
		return "", err
	} else {
		return strings.Trim(string(content), "\n"), nil
	}
}

func ReadUint64(filename string) (uint64, error) {
	val, err := ReadString(filename)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(val, 10, 64)
}


func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ForEachFile(filename string, fn func(line string) (bool, error)) error{
	var line string
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
	}()
	buf := bufio.NewReader(f)
	err = nil
	for err == nil {
		line, err = buf.ReadString('\n')
		if err != nil {
			break
		}
		if continu, _ := fn(strings.Trim(line, "\n")); !continu {
			break
		}
	}

	if err == io.EOF {
		return nil
	}
	return err
}
