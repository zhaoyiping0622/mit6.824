package shard

import "6.824/raftapp"

type ShardExecutor struct {
  ShardWrapper
}

func (se *ShardExecutor) RunCommand(args *raftapp.AsyncRequestArgs) {
  for _,id:=range args.Location.GetExecutorIds() {
    if id == raftapp.ExecutorIdApp {
      location:=args.Location.(*raftapp.ExecutorLocationImpl).NoticeLocation
      e:=se.GetShard(location).(raftapp.Executor)
      e.RunCommand(args)
    }
  }
}

func (se *ShardExecutor) Clean() {}

func MakeShardExecutor(f func(shard int)raftapp.InnerExecutor, notice raftapp.Notice, executorId int64, me int) *ShardExecutor {
  return &ShardExecutor{
    ShardWrapper:MakeShardWrapper(
      func(shard int) raftapp.CleanSnapshotable {
        return raftapp.MakeExecutor(f(shard), notice, executorId)
      }, raftapp.MakeWrongGroupExecutor(), me, "ShardExecutor"),
  }
}
