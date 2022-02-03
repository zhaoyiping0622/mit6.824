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

  term,isLeader:=0,false
  tickers:=[]raftapp.Ticker{
    raftapp.DefaultTicker().SetFunc(func(ctx context.Context) {
      termA,isLeaderA:=rf.GetState()
      if termA != term {
        if isLeaderA {
          DPrintf("%v change to leader", me)
        }
        term=termA
        isLeader=isLeaderA
        c.SyncRequest(&raftapp.SyncRequestArgs{
          Command: &raftapp.TermRequest{
            Term: term,
            IsLeader: isLeader,
        }})
      }
    }), // term ticker
    raftapp.DefaultTicker().SetFunc(func(ctx context.Context) {
      if maxraftstate == -1 {
        return
      }
      if persister.RaftStateSize() >= maxraftstate {
        c.SyncRequest(&raftapp.SyncRequestArgs{ Command: &raftapp.SnapshotRequest{ Ctx: ctx } })
      }
    }),
  }
  notice:=MakeSingleNotice()
  var executorId int64 =1
  app := &raftapp.ExecutorImpl{
      InnerExecutor: executor,
      Notice: notice,
      ExecutorId: executorId,
    }

  // init controller
  c.Init(
    me,applyCh,rf,notice,
    []raftapp.TermNotice{
      notice,
    },
    []raftapp.Executor{
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
  ss.background, ss.backgroundCancel = context.WithCancel(context.Background())
  ss.RpcServer=raftapp.MakeRpcServer(me, executorId, c)
  ss.rf=rf

  for _,t:=range ss.tickers {
    go t.Run(ss.background)
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

