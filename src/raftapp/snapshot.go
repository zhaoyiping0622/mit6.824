package raftapp

import (
	"bytes"
	"context"
	"time"

	"6.824/labgob"
	"6.824/raft"
)

type RaftServerSnapshotState struct {
  AppState interface{}
  Sessions map[int64]*Session
}

func (rs *RaftServer) loadSnapshot(snapshot raft.Snapshot, index int) {
  var state RaftServerSnapshotState
  buffer:=bytes.NewBuffer(snapshot)
  decoder:=labgob.NewDecoder(buffer)
  decoder.Decode(&state)
  rs.app.ApplySnapshot(state.AppState)
  rs.sessions=state.Sessions
  rs.lastApplied = index
}

func (rs *RaftServer) generateSnapshot() (raft.Snapshot, int) {
  index:=rs.lastApplied
  state:=RaftServerSnapshotState{
    AppState: rs.app.CreateSnapshot(),
    Sessions: rs.sessions,
  }
  buffer:=new(bytes.Buffer)
  encoder:=labgob.NewEncoder(buffer)
  encoder.Encode(state)
  data:=buffer.Bytes()
  return data,index
}

func (rs *RaftServer) needSnapshot() bool {
  return rs.maxraftstate != -1 && rs.persister.RaftStateSize() >= rs.maxraftstate
}

func (rs *RaftServer) snapshotLoop() {
  for {
    select {
    case <-rs.background.Done(): return
    case <-time.After(time.Microsecond*10):
    }
    if rs.needSnapshot() {
      ctx,cancel:=context.WithCancel(rs.background)
      go rs.sendEvent(&SnapshotEvent{cancel})
      <-ctx.Done()
    }
  }
}

type SnapshotEvent struct {
  done context.CancelFunc
}

func (e *SnapshotEvent) Run(rs *RaftServer) {
  DPrintf("%v request snapshot", rs.me)
  snapshot,index:=rs.generateSnapshot()
  rs.snapshotFinish = e.done
  go func() {
    if !rs.rf.Snapshot(index,snapshot) {
      e.done()
    }
  }()
}
