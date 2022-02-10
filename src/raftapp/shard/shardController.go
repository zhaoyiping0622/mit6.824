package shard

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"6.824/labgob"
	"6.824/labrpc"
	"6.824/raft"
	"6.824/raftapp"
	"6.824/shardctrler"
)

// Ticker
// -- info
// -- query shard from other group
// -- update shardctrler
// Executor
// TermNotice
type ShardController struct {
  me int
  gid int
  mu sync.Mutex
  tickers []raftapp.Ticker
  shardCtrlerCli [NShards]*shardctrler.Clerk

  ctrlerCli *shardctrler.Clerk

  currentInfo *shardctrler.InfoResult
  shardStates [NShards]ShardState
  shardVersions [NShards]int

  shardSnapshots [NShards]*raftapp.SnapshotGroup
  shardSnapshotVersions [NShards]int

  shardSessionIds [NShards]int64
  shardSeqNums [NShards]int
  InfoSessionId int64
  InfoSeqNum int

  shardNotice *ShardNotice

  shardWrappers []ShardWrapper

  controller raftapp.RaftController

  term int
  isLeader int32
  inited bool

  make_end func(string) *labrpc.ClientEnd
  make_innerClient func(servers []*labrpc.ClientEnd)*InnerClient
}

func init() {
  labgob.Register(&ShardControllerSnapshot{})
}

type ShardControllerSnapshot struct {
  CurrentInfo *shardctrler.InfoResult
  ShardStates [NShards]ShardState
  ShardVersions [NShards]int

  ShardSnapshots [NShards]*raftapp.SnapshotGroup
  ShardSnapshotVersions [NShards]int
}

func (sc *ShardController)GenerateSnapshot() raftapp.Snapshot {
  sc.mu.Lock()
  defer sc.mu.Unlock()
  snapshot:=&ShardControllerSnapshot{
    CurrentInfo: sc.currentInfo,
    ShardStates: sc.shardStates,
    ShardVersions: sc.shardVersions,
    ShardSnapshots: sc.shardSnapshots,
    ShardSnapshotVersions: sc.shardSnapshotVersions,
  }
  for i:=range snapshot.ShardSnapshots {
    if snapshot.ShardSnapshots[i] == nil {
      snapshot.ShardSnapshots[i]=&raftapp.SnapshotGroup{
        Snapshots: make([]raft.Snapshot, 0),
      }
    }
  }
  return raftapp.ValueToSnapshot(snapshot)
}
func (sc *ShardController)ApplySnapshot(i raftapp.Snapshot) {
  sc.mu.Lock()
  defer sc.mu.Unlock()
  snapshot:=new(ShardControllerSnapshot)
  raftapp.SnapshotToValue(i, &snapshot)
  sc.currentInfo=snapshot.CurrentInfo
  sc.shardStates=snapshot.ShardStates
  sc.shardVersions=snapshot.ShardVersions
  sc.shardSnapshots=snapshot.ShardSnapshots
  sc.shardSnapshotVersions=snapshot.ShardSnapshotVersions
  DPrintf("%v apply snapshot state %+v version %+v", sc.gid, sc.shardStates, sc.shardVersions)
}

func (sc *ShardController) getIsLeader() bool {
  return atomic.LoadInt32(&sc.isLeader)==1
}

func (sc *ShardController) setIsLeader(isLeader bool) {
  if isLeader {
    atomic.StoreInt32(&sc.isLeader, 1)
  } else {
    atomic.StoreInt32(&sc.isLeader, 0)
  }
}

func (sc *ShardController) UpdateTermAndLeader(term int, isLeader bool) {
  sc.term=term
  sc.setIsLeader(isLeader)
}

func (sc *ShardController) updateInfo(info *shardctrler.InfoResult) {
  sc.InfoSeqNum++
  seqNum:=sc.InfoSeqNum
  sc.controller.AsyncRequest(&raftapp.AsyncRequestArgs{
    Location: raftapp.MakeLocation(sc.InfoSessionId, raftapp.ExecutorIdShard),
    Command: &UpdateInfo{info},
    SeqNum: seqNum,
  })
}

func (sc *ShardController) UpdateInfo() {
  if !sc.getIsLeader() {
    return
  }
  newInfo:=sc.ctrlerCli.Info()
  sc.mu.Lock()
  if sc.currentInfo!=nil && sc.currentInfo.Config.Num==newInfo.Config.Num && sc.currentInfo.Version==newInfo.Version {
    sc.mu.Unlock()
    return
  }
  sc.mu.Unlock()
  sc.updateInfo(newInfo)
}

func (sc *ShardController) ChangeShardState(shard int, to ShardState, snapshot *raftapp.SnapshotGroup, version int, newInfo *shardctrler.InfoResult) {
  sc.shardSeqNums[shard]++
  // DPrintf("%v request to change shard %v shardstate to %+v", sc.gid, shard, to)
  if snapshot == nil {
    snapshot=&raftapp.SnapshotGroup{ Snapshots: make([]raft.Snapshot, len(sc.shardWrappers)) }
  }
  sc.controller.AsyncRequest(&raftapp.AsyncRequestArgs{
    Location: raftapp.MakeLocation(sc.shardSessionIds[shard], raftapp.ExecutorIdShard),
    Command: &ChangeShardState{
      Shard: shard,
      To: to,
      Snapshot: snapshot,
      Version: version,
      Info: newInfo,
    },
    SeqNum: sc.shardSeqNums[shard],
  })
}

func (sc *ShardController) UpdateShard(shard int) raftapp.TickerFunc {
  var cli *InnerClient
  return func(ctx context.Context) {
    sc.mu.Lock()
    if sc.currentInfo == nil || !sc.getIsLeader() || !sc.inited {
      sc.mu.Unlock()
      return
    }
    info:=sc.currentInfo
    state:=sc.shardStates[shard]
    version:=sc.shardVersions[shard]
    snapshot:=sc.shardSnapshots[shard]
    snapshotVersion:=sc.shardSnapshotVersions[shard]

    config:=info.Config
    shardInfo:=info.ShardInfos[shard]
    groups:=info.Groups
    sc.mu.Unlock()
    // DPrintf("%v in UpdateShard shard %v version %v state %v", sc.me, shard, version, state)
    switch state {
      case HERE:
        if config.Shards[shard] != sc.gid {
          sc.ChangeShardState(shard, LOCKED, nil, version, nil)
        }
      case LOCKED:
        if config.Shards[shard] == sc.gid && shardInfo.Version == version && version == snapshotVersion {
          sc.ChangeShardState(shard, WAITED, snapshot, version+1, nil)
        } else if shardInfo.Owner != sc.gid || shardInfo.Version > version {
          sc.ChangeShardState(shard, MOVED, nil, version, nil)
        }
      case MOVED:
        if shardInfo.Owner == sc.gid {
          // panic(fmt.Sprintf("%v shard %v info %+v", sc.me, shard, raftapp.PrettyPrint(info)))
          return
        } else if config.Shards[shard] != sc.gid {
          return
        } else if shardInfo.Owner == -1 {
          // uninitialized
          if shardInfo.Version != 0 {
            panic("impossible")
          }
          sc.ChangeShardState(shard, WAITED, nil, 1, nil)
        } else {
          targetGID:=shardInfo.Owner
          groups:=groups[targetGID]
          servers:=make([]*labrpc.ClientEnd, len(groups))
          for i:=range groups {
            servers[i]=sc.make_end(groups[i])
          }
          if cli == nil {
            cli=sc.make_innerClient(servers)
          } else {
            cli.SetServer(servers)
          }
          getShardResult1:=cli.Send(&GetShard{
            InfoNum: config.Num,
            InfoVersion: info.Version,
            From: sc.gid,
            Shard: shard,
            Version: shardInfo.Version,
          })
          if getShardResult1 != nil {
            // change state from moved to locked
            getShardResult:=getShardResult1.(*GetShardResult)
            sc.ChangeShardState(shard, WAITED, getShardResult.Snapshot, getShardResult.Version+1, nil)
          } else {
            DPrintf("%v shard %v get unknown shard result", sc.me, shard)
          }
        }
      case WAITED:
        if shardInfo.Owner == sc.gid && shardInfo.Version == version {
          sc.ChangeShardState(shard, HERE, nil, shardInfo.Version, nil)
        } else {
          updateResult:=sc.shardCtrlerCli[shard].UpdateOwner(&shardctrler.ShardInfo{
            Owner: sc.gid,
            Version: version,
          }, shard)
          if updateResult.Success {
            sc.ChangeShardState(shard, HERE, nil, version, updateResult.InfoResult)
          } else if updateResult.ShardInfos[shard].Owner!=sc.gid && updateResult.ShardInfos[shard].Version >= version {
            sc.ChangeShardState(shard, MOVED, nil, version, updateResult.InfoResult)
          }
        }
    }
  }
}

func (sc *ShardController) RunCommand(args *raftapp.AsyncRequestArgs) {
  for _,id:=range args.Location.GetExecutorIds() {
    if id == raftapp.ExecutorIdInit {
      func() {
        sc.mu.Lock()
        defer sc.mu.Unlock()
        sc.inited=true
        DPrintf("%v inited", sc.me)
      }()
      continue
    }
    if id == raftapp.ExecutorIdShard {
      // DPrintf("%v get shard request %+v", sc.gid, args)
      switch c:=args.Command.(type) {
      case *ChangeShardState: func(){
        sc.mu.Lock()
        defer sc.mu.Unlock()
        defer sc.shardNotice.SetValue(args.Location, args.SeqNum, &raftapp.AsyncRequestReply{ Err: raftapp.Ok, })
        shard:=c.Shard
        currentState:=sc.shardStates[shard]
        currentVersion:=sc.shardVersions[shard]
        if c.Info!=nil {
          info:=c.Info
          if info.Config.Num > sc.currentInfo.Config.Num || info.Version > sc.currentInfo.Version {
            sc.currentInfo=info
          }
        }
        if c.Version < currentVersion || (c.Version==currentVersion && currentState >= c.To){
          return
        }
        snapshots:=&raftapp.SnapshotGroup{}
        // DPrintf("%v shard %v state change to %+v", sc.gid, shard, c.To)
        for i:=range sc.shardWrappers {
          snapshots.AddSnapshots(sc.shardWrappers[i].ChangeShardState(shard, c.To, c.Snapshot.Snapshots[i]))
        }
        sc.shardStates[shard]=c.To
        sc.shardVersions[shard]=c.Version
        if c.To == LOCKED {
          sc.shardSnapshots[shard]=snapshots
          sc.shardSnapshotVersions[shard]=c.Version
        } else if c.To == MOVED || c.To == HERE {
          sc.shardSnapshots[shard]=nil
          sc.shardSnapshotVersions[shard]=0
        }
        // DPrintf("%v shard %v change to %+v", sc.gid, shard, c.To)
      }()
      case *UpdateInfo: func() {
        sc.mu.Lock()
        defer sc.mu.Unlock()
        defer sc.shardNotice.SetValue(args.Location, args.SeqNum, &raftapp.AsyncRequestReply{ Err: raftapp.Ok, })
        if sc.currentInfo!=nil && sc.currentInfo.Config.Num == c.Info.Config.Num && sc.currentInfo.Version == c.Info.Version {
          return
        }
        sc.currentInfo=c.Info
        DPrintf("%v updated info to version %v num %v", sc.me, c.Info.Version, c.Info.Config.Num)
      }()
      }
    }
  }
}

func MakeShardController(ctx context.Context, gid int, ctrlerServers []*labrpc.ClientEnd, shardNotice *ShardNotice, shardExecutor *ShardExecutor, raftController raftapp.RaftController, make_end func(string)*labrpc.ClientEnd, name string, me int) *ShardController {
  sc:=new(ShardController)

  sc.me=me
  sc.shardNotice=shardNotice
  sc.gid=gid
  sc.tickers=make([]raftapp.Ticker, 0)
  sc.tickers = append(sc.tickers,
    raftapp.DefaultTicker().SetTickerFunc(func(context.Context) { sc.UpdateInfo() }).SetTickerDuration(50*time.Millisecond),
  )
  for i:=0;i<NShards;i++ {
    sc.tickers = append(sc.tickers,
      raftapp.DefaultTicker().SetTickerFunc(sc.UpdateShard(i)).SetTickerDuration(50*time.Millisecond),
    )
  }

  for i:=0;i<NShards;i++ {
    sc.shardCtrlerCli[i]=shardctrler.MakeClerk(ctrlerServers)
    sc.shardStates[i]=MOVED
  }
  sc.ctrlerCli=shardctrler.MakeClerk(ctrlerServers)

  for i:=0;i<NShards;i++ {
    sc.shardSessionIds[i]=raftapp.Nrand()
  }
  sc.InfoSessionId=raftapp.Nrand()

  sc.shardWrappers=[]ShardWrapper{shardNotice, shardExecutor}
  sc.controller=raftController

  sc.make_end=make_end
  sc.make_innerClient=func(servers []*labrpc.ClientEnd) *InnerClient { return MakeInnerClient(servers, name) }

  for _,t:=range sc.tickers {
    go t.TickerRun(ctx)
  }

  return sc
}
