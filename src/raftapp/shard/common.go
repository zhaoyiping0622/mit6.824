package shard

import (
	"os"
	"strings"

	"6.824/raftapp"
	"6.824/shardctrler"
)

var debug bool

func init() {
  for _,s:=range strings.Split(os.Getenv("debug"), ",") {
    if s == "shardRaftapp" {
      debug=true
      return
    }
  }
}

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if debug {
    raftapp.DPrintf("shard "+format, a...)
	}
	return
}

const NShards = shardctrler.NShards
type Config = shardctrler.Config
