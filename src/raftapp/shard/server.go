package shard

import (
	"context"
	"fmt"

	"6.824/labrpc"
	"6.824/raft"
	"6.824/raftapp"
)

type ShardServer struct {
  me int
  rf *raft.Raft
  controller raftapp.RaftController
  shardController *ShardController
  tickers []raftapp.Ticker
  background context.Context
  backgroundCancel context.CancelFunc
  *raftapp.RpcServer
}

func (ss *ShardServer) Kill() {
  DPrintf("%v server killed", ss.me)
  ss.backgroundCancel()
  ss.rf.Kill()
}

func (ss *ShardServer) GetRaft() *raft.Raft {
  return ss.rf
}

func MakeShardServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int, gid int, ctrlers []*labrpc.ClientEnd, make_end func(string) *labrpc.ClientEnd, name string, executorBuilder func(shard int)raftapp.InnerExecutor) *ShardServer {
  ss:=new(ShardServer)
  c:=new(raftapp.RaftControllerImpl)
  applyCh:=make(chan raft.ApplyMsg)
  DPrintf("gid %d", gid)
  rf:=raft.MakeRaft(servers, me, persister, applyCh, fmt.Sprint(gid*10+me))
  me=gid*10+me
  ss.me=me
  ss.background, ss.backgroundCancel = context.WithCancel(context.Background())
  ss.rf=rf

  tickers:=[]raftapp.Ticker{
    raftapp.MakeTermTicker(rf, me, c),
    raftapp.MakeSnapshotTicker(maxraftstate, persister, c),
  }
  ss.tickers=tickers
  notice:=MakeShardNotice(me)
  executor:=MakeShardExecutor(executorBuilder, notice, raftapp.ExecutorIdApp, me)

  ss.controller=c
  ss.RpcServer=raftapp.MakeRpcServer(me, []int64{raftapp.ExecutorIdApp}, c)

  shardController:=MakeShardController(ss.background, gid, ctrlers, notice, executor, c, make_end, name, me)
  ss.shardController=shardController

  c.Init(ss.background, me, applyCh, rf, notice,
    []raftapp.TermNotice{
      notice, shardController,
    },
    []raftapp.Executable{
      executor,shardController,
    },
    []raftapp.Snapshotable{
      notice, executor, shardController,
    },
  )

  for _,t:=range ss.tickers {
    go t.TickerRun(ss.background)
  }
  go c.Run(ss.background)

  DPrintf("%v started", ss.me)

  return ss
}

func (ss *ShardServer) ShardRpcRequest(request *raftapp.OutterRpcArgs, reply *raftapp.OutterRpcReply) {
  ss.shardController.ShardRpcRequest(request, reply)
}
