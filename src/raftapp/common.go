package raftapp

import (
	"context"
	"log"
	"os"
	"strings"
	"sync/atomic"
)

func (rs *RaftServer) Kill() {
  ctx,cancel:=context.WithCancel(rs.background)
  go rs.sendEvent(&KillEvent{cancel})
  <-ctx.Done()
}

type KillEvent struct {
  done context.CancelFunc
}

func (e *KillEvent) Run(rs *RaftServer) {
  defer rs.backgroundCancel()
  if e.done != nil {
    defer e.done()
  }
  atomic.StoreInt32(&rs.dead,1)
  rs.rf.Kill()
}

var Debug bool

func init() {
  debug:=os.Getenv("debug")
  for _,s:= range strings.Split(debug, ",") {
    if s == "raftapp" {
      Debug = true
      return
    }
  }
}

func DPrintf(format string, a ...interface{}) (n int, err error) {
  if Debug {
    log.Printf(format, a...)
  }
  return
}
