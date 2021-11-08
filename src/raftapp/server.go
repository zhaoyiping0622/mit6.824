package raftapp

import (
	"context"

	"6.824/labgob"
	"6.824/labrpc"
	"6.824/raft"
)

type RaftConfig struct {
  EventLength int
}

func MakeRaftConfig() *RaftConfig {
  return &RaftConfig{
    EventLength: RaftEventLength,
  }
}

type RaftServer struct {
  me int
  rf *raft.Raft
  dead int32
  applyCh chan raft.ApplyMsg
  persister *raft.Persister
  maxraftstate int

  background context.Context
  backgroundCancel context.CancelFunc
  leaderCtx context.Context
  leaderCancel context.CancelFunc
  term int

  snapshotFinish context.CancelFunc

  events chan Event
  sessions map[int64]*Session
  triggers map[int64]*Trigger

  lastApplied int
  app RaftApp
}

func MakeRaftServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int, app RaftApp) *RaftServer {
  labgob.Register(Op{})
  rs:=new(RaftServer)
  rs.applyCh = make(chan raft.ApplyMsg)
  rs.rf = raft.Make(servers,me,persister,rs.applyCh)
  rs.persister = persister
  rs.maxraftstate = maxraftstate
  rs.app = app

  rs.background, rs.backgroundCancel = context.WithCancel(context.Background())

  go rs.statusTicker()
  go rs.eventLoop()
  go rs.applyLoop()

  return rs
}

func (rs *RaftServer) GetRaft() *raft.Raft {
  return rs.rf
}
