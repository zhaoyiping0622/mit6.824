package raftapp

import (
	"context"

	"6.824/raft"
)

func (rs *RaftServer) applyLoop() {
  for {
    select {
    case <-rs.background.Done():return
    case msg,ok:=<-rs.applyCh:
      if !ok {
        return
      }
      DPrintf("%v get msg %+v", rs.me, msg)
      if msg.CommandValid {
        if msg.CommandIndex <= rs.lastApplied {
          DPrintf("%v ignore msg %+v", rs.me, msg)
          continue
        } else if msg.CommandIndex == rs.lastApplied + 1 {
          if op,ok:=msg.Command.(Op); ok {
            ctx,cancel:=context.WithCancel(rs.background)
            go rs.sendEvent(&ApplyCommandEvent{
              done: cancel,
              op: op,
            })
            <-ctx.Done()
          } else {
            DPrintf("%v unknown command %+v", rs.me, msg)
          }
        } else {
          DPrintf("%v msg out of order lastApplied %v index %v", rs.me, rs.lastApplied, msg.CommandIndex)
        }
      } else if msg.SnapshotValid && rs.rf.CondInstallSnapshot(msg.SnapshotTerm, msg.SnapshotIndex, msg.Snapshot) {
        ctx,cancel:=context.WithCancel(rs.background)
        go rs.sendEvent(&ApplySnapshotEvent{
          done: cancel,
          snapshot: msg.Snapshot,
          index: msg.SnapshotIndex,
        })
        <-ctx.Done()
      }
    }
  }
}

type ApplyCommandEvent struct {
  done context.CancelFunc
  op Op
}

func (e *ApplyCommandEvent) Run(rs *RaftServer) {
  if e.done != nil {
    defer e.done()
  }
  var session *Session
  op:=&e.op
  if s,ok:=rs.sessions[op.SessionId]; ok {
    session = s
  } else {
    session = new(Session)
    rs.sessions[op.SessionId] = session
  }
  rs.lastApplied++
  if session.SeqNum+1 != op.SeqNum {
    return
  }
  session.SeqNum++
  reply := op.Command.Apply(rs.app)
  session.Result = reply
  if trigger,ok:=rs.triggers[op.SessionId];ok && op.SeqNum == trigger.seqNum {
    *trigger.result=*reply
    if trigger.done != nil {
      trigger.done()
      trigger.done = nil
    }
    delete(rs.triggers, op.SessionId)
  }
}

type ApplySnapshotEvent struct {
  done context.CancelFunc
  snapshot raft.Snapshot
  index int
}
func (e *ApplySnapshotEvent) Run(rs *RaftServer) {
  DPrintf("%v apply snapshot with index %v", rs.me, e.index)
  if e.done != nil {
    defer e.done()
  }
  if rs.snapshotFinish != nil {
    finish:=rs.snapshotFinish
    defer finish()
    rs.snapshotFinish = nil
  }
  if e.index <= rs.lastApplied {
    return
  }
  rs.loadSnapshot(e.snapshot, e.index)
}
