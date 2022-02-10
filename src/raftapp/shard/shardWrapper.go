package shard

import (
	"sync"

	"6.824/labgob"
	"6.824/raftapp"
)

func init() {
  labgob.Register(&ShardWrapperImplSnapshot{})
}

type ShardSnapshotableConstructor = func(shard int)raftapp.CleanSnapshotable

type ShardWrapper interface {
  raftapp.Snapshotable
  ChangeShardState(shard int, to ShardState, snapshot raftapp.Snapshot) raftapp.Snapshot
  GetShard(location raftapp.NoticeLocation) raftapp.CleanSnapshotable
  ForAll(func(raftapp.CleanSnapshotable))
}

type ShardWrapperImpl struct {
  name string
  me int
  executorId int64
  mu sync.Mutex
  states [NShards]ShardState
  values [NShards]raftapp.CleanSnapshotable
  movedShard raftapp.CleanSnapshotable
  constructor ShardSnapshotableConstructor
}

// lock
func (s *ShardWrapperImpl) GetShard(location raftapp.NoticeLocation) raftapp.CleanSnapshotable {
  s.mu.Lock()
  defer s.mu.Unlock()
  if l,ok:=location.(*ShardNoticeLocation); ok && (s.states[l.GetShard()] == HERE) {
    return s.values[l.GetShard()]
  } else {
    if ok {
      DPrintf("%v shard %v state %+v", s.me, l.GetShard(), s.states[l.GetShard()])
    }
    return s.movedShard
  }
}

func (s *ShardWrapperImpl) ChangeShardState(shard int, to ShardState, snapshot raftapp.Snapshot) raftapp.Snapshot {
  s.mu.Lock()
  defer s.mu.Unlock()
  DPrintf("%v shard %v state change from %+v to %+v", s.me, shard, s.states[shard], to)
  snapshot1:=s.generateShardSnapshot(shard)
  if to == LOCKED || to == MOVED {
    s.values[shard].Clean()
    // DPrintf("%v %v shard %v clean", s.me, s.name, shard)
  }
  s.states[shard]=to
  s.applySnapshotToShard(shard, snapshot)
  if shard != -1 && s.states[shard]==LOCKED {
    return snapshot1
  } else {
    return nil
  }
}

type ShardWrapperImplSnapshot struct {
  States [NShards]ShardState
  Values raftapp.SnapshotGroup
}

func (s *ShardWrapperImpl) applySnapshotToShard(shard int, snapshot raftapp.Snapshot) {
  if len(snapshot) == 0 { return }
  // DPrintf("%v %v apply snapshot %+v to shard %v", s.me, s.name, raftapp.PrettyPrint(snapshot), shard)
  s.values[shard]=s.constructor(shard)
  s.values[shard].ApplySnapshot(snapshot)
}

func (s *ShardWrapperImpl) generateShardSnapshot(shard int) raftapp.Snapshot {
  if s.states[shard] == HERE || s.states[shard] == WAITED {
    snapshot:=s.values[shard].GenerateSnapshot()
    // DPrintf("%v %v generate snapshot %+v shard %v", s.me, s.name, raftapp.PrettyPrint(snapshot), shard)
    return snapshot
  } else {
    return raftapp.Snapshot{}
  }
}

// lock
func (s *ShardWrapperImpl) ApplySnapshot(i raftapp.Snapshot) {
  s.mu.Lock()
  defer s.mu.Unlock()
  var snapshot ShardWrapperImplSnapshot
  raftapp.SnapshotToValue(i, &snapshot)
  // DPrintf("%v %v shardwrapper apply snapshot %+v", s.me, s.name, raftapp.PrettyPrint(snapshot))

  for i:=range s.values {
    s.states[i]=snapshot.States[i]
    s.applySnapshotToShard(i, snapshot.Values.Snapshots[i])
  }
}

// lock
func (s *ShardWrapperImpl) GenerateSnapshot() raftapp.Snapshot {
  s.mu.Lock()
  defer s.mu.Unlock()
  var snapshot ShardWrapperImplSnapshot
  snapshot.States=s.states
  for i:=range s.values {
    snapshot.Values.AddSnapshots(s.generateShardSnapshot(i))
  }
  return raftapp.ValueToSnapshot(snapshot)
}

// lock
func (s *ShardWrapperImpl) ForAll(f func(raftapp.CleanSnapshotable)) {
  s.mu.Lock()
  defer s.mu.Unlock()
  wg:=sync.WaitGroup{}
  wg.Add(NShards)
  for i,v:=range s.values {
    go func(i int, v raftapp.CleanSnapshotable){
      defer wg.Done()
      if s.states[i] == HERE {
        f(v)
      }
    }(i, v)
  }
  wg.Wait()
}

func MakeShardWrapper(f ShardSnapshotableConstructor, movedShard raftapp.CleanSnapshotable, me int, name string) ShardWrapper {
  s:=new(ShardWrapperImpl)
  s.name=name
  s.me=me
  s.constructor=f
  s.executorId=raftapp.ExecutorIdShard
  s.movedShard=movedShard
  wg:=sync.WaitGroup{}
  wg.Add(len(s.values))
  for i:=range s.values {
    s.states[i]=MOVED
    go func(i int){
      defer wg.Done()
      s.values[i]=f(i)
    }(i)
  }
  wg.Wait()
  return s
}

