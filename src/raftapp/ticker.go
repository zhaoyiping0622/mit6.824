package raftapp

import (
	"context"
	"time"
)

func (rs *RaftServer) statusTicker() {
  for {
    select {
    case <-rs.background.Done():return
    case <-time.After(StatusTickerInterval):
    }
    wasLeader:=(rs.leaderCtx!=nil)
    term,isLeader:=rs.rf.GetState()
    if isLeader != wasLeader || term != rs.term {
      ctx,cancel:=context.WithCancel(rs.background)
      go rs.sendEvent(&StatusChangeEvent{
        isLeader: isLeader,
        term: term,
        done: cancel,
      })
      <-ctx.Done()
    }
  }
}

type StatusChangeEvent struct {
  isLeader bool
  term int
  done context.CancelFunc
}

func (e *StatusChangeEvent) Run(rs *RaftServer) {
  if e.done != nil {
    defer e.done()
  }
  wasLeader:=(rs.leaderCtx!=nil)
  if wasLeader == e.isLeader && rs.term == e.term {
    return
  }
  if wasLeader {
    rs.leaderCancel()
    rs.leaderCtx = nil
  }
  for sessionId,trigger:=range rs.triggers {
    if trigger.term < e.term {
      trigger.result.Err = ErrWrongLeader
      trigger.done()
      delete(rs.triggers, sessionId)
    }
  }
  rs.term = e.term
  if e.isLeader {
    rs.leaderCtx, rs.leaderCancel=context.WithCancel(rs.background)
  }
}
