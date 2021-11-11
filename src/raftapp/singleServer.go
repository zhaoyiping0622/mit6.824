package raftapp

import (
	"context"
	"sync"

	"6.824/labrpc"
	"6.824/raft"
)

type SingleServer struct {
  me int
  mu sync.Mutex
  applier Applier
  background context.Context
  finish context.CancelFunc
  maxraftstate int
  rf *raft.Raft
}

func(ss *SingleServer) Kill() {
  ss.mu.Lock()
  defer ss.mu.Unlock()
  ss.finish()
  ss.rf.Kill()
}

func MakeSingleServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int, app RaftApp) *SingleServer {
  ss:=new(SingleServer)
  ss.me=me
  applyCh:=make(chan raft.ApplyMsg)
  ss.rf=raft.Make(servers, me, persister, applyCh)
  ss.applier=MakeSingleApplier(me, ss.rf, app, applyCh)
  ss.background, ss.finish = context.WithCancel(context.Background())
  ss.maxraftstate = maxraftstate
  go MakeStatusTicker(ss.rf, ss.applier, me).Run(ss.background)
  go MakeSnapshotTicker(persister, ss.applier, maxraftstate).Run(ss.background)
  go ss.applier.Run(ss.background)
  return ss
}

func (ss *SingleServer) CommandRequest(args *CommandRequestArgs, reply *CommandRequestReply) {
  DPrintf("%v get request %+v", ss.me, args)
  ss.applier.CommandRequest(args, reply)
}

func (ss *SingleServer) GetRaft() *raft.Raft {
  return ss.rf
}
