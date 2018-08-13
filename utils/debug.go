package utils

import (
	"github.com/davecgh/go-spew/spew"
	"os"
)

func Debug(val ...interface{}) {
	if os.Getenv("debug_ctools") == "on" {
		spew.Dump(val...)
	}
}