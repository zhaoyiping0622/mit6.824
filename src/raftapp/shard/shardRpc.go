package shard

import (
	"6.824/labgob"
	"6.824/raftapp"
	"6.824/shardctrler"
)

func init() {
  labgob.Register(&GetShard{})
  labgob.Register(&GetShardResult{})
  labgob.Register(&ChangeShardState{})
  labgob.Register(&UpdateInfo{})
}

func (sc *ShardController) ShardRpcRequest(request *raftapp.OutterRpcArgs, reply *raftapp.OutterRpcReply) {
  sc.mu.Lock()
  defer sc.mu.Unlock()
  if !sc.getIsLeader() || sc.currentInfo == nil {
    reply.Err=raftapp.WrongLeader
    return
  }
  info:=sc.currentInfo
  config:=info.Config
  switch c:=request.Command.(type) {
    case *GetShard:
      shard:=c.Shard
      if c.From != config.Shards[shard] && (c.InfoNum < config.Num || c.InfoVersion < info.Version) {
        reply.Err=raftapp.Ok
        reply.Result=nil
      } else if c.Version == sc.shardSnapshotVersions[shard] {
        reply.Err=raftapp.Ok
        reply.Result=&GetShardResult{
          Shard: shard,
          Version: sc.shardSnapshotVersions[shard],
          Snapshot: sc.shardSnapshots[shard],
        }
      } else {
        DPrintf("%v get request to shard %+v current snapshot version %+v", sc.me, c, sc.shardSnapshotVersions[shard])
        reply.Err=raftapp.WrongLeader
      }
  }
  DPrintf("%v get rpc request %T%+v reply %+v", sc.me, request.Command, raftapp.PrettyPrint(request), raftapp.PrettyPrint(reply))
}

type GetShard struct {
  InfoNum int
  InfoVersion int
  From int
  Version int
  Shard int
}

type GetShardResult struct {
  Version int
  Shard int
  Snapshot *raftapp.SnapshotGroup
}

type ChangeShardState struct {
  Shard int
  To ShardState
  Snapshot *raftapp.SnapshotGroup
  Version int
  Info *shardctrler.InfoResult
}

type UpdateInfo struct {
  Info *shardctrler.InfoResult
}
