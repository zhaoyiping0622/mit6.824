package raftapp

import (
	"log"
	"os"
	"strings"
)

var debug bool

func init() {
  for _,s:=range strings.Split(os.Getenv("debug"), ",") {
    if s == "raftapp" {
      debug=true
      return
    }
  }
}

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if debug {
		log.Printf(format, a...)
	}
	return
}
