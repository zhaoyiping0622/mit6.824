package single

import (
	"context"

	"6.824/labrpc"
	"6.824/raft"
	"6.824/raftapp"
)

type SingleServer struct {
  controller raftapp.RaftController
  tickers []raftapp.Ticker
  background context.Context
  backgroundCancel context.CancelFunc
  *raftapp.RpcServer
  rf *raft.Raft
}

func MakeSingleServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int, executor raftapp.InnerExecutor) *SingleServer {
  ss:=new(SingleServer)
  DPrintf("%v executor %p", me, executor)
  c:=new(raftapp.RaftControllerImpl)
  applyCh:=make(chan raft.ApplyMsg)
  rf:=raft.Make(servers, me, persister, applyCh)

  tickers:=[]raftapp.Ticker{
    raftapp.MakeTermTicker(rf, me, c),
    raftapp.MakeSnapshotTicker(maxraftstate, persister, c),
  }
  notice:=MakeSingleNotice()
  executorId:=raftapp.ExecutorIdApp
  app := &raftapp.ExecutorImpl{
    InnerExecutor: executor,
    Notice: notice,
    ExecutorId: executorId,
  }

  ss.background, ss.backgroundCancel = context.WithCancel(context.Background())

  // init controller
  c.Init(
    ss.background,
    me,applyCh,rf,notice,
    []raftapp.TermNotice{
      notice,
    },
    []raftapp.Executable{
      app,
    },
    []raftapp.Snapshotable{
      app,
      notice,
    },
  )

  // init server
  ss.controller=c
  ss.tickers = tickers
  ss.RpcServer=raftapp.MakeRpcServer(me, []int64{executorId}, c)
  ss.rf=rf

  for _,t:=range ss.tickers {
    go t.TickerRun(ss.background)
  }
  go c.Run(ss.background)

  return ss
}

func (ss *SingleServer) GetRaft() *raft.Raft {
  return ss.rf
}

func(ss *SingleServer) Kill() {
  ss.backgroundCancel()
  ss.rf.Kill()
}

