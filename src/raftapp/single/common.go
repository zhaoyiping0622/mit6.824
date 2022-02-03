package single

import (
	"os"
	"strings"

	"6.824/raftapp"
)

var debug bool

func init() {
  for _,s:=range strings.Split(os.Getenv("debug"), ",") {
    if s == "singleRaftapp" {
      debug=true
      return
    }
  }
}

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if debug {
    raftapp.DPrintf(format, a...)
	}
	return
}

