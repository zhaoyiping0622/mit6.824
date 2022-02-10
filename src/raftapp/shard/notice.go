package shard

import (
	"6.824/raftapp"
	"6.824/raftapp/single"
)

type ShardNotice struct {
  ShardWrapper
  Notice raftapp.Notice
}

func (s *ShardNotice) getNotice(location raftapp.NoticeLocation) raftapp.Notice {
  location=location.(*raftapp.ExecutorLocationImpl).NoticeLocation
  if _,ok:=location.(*ShardNoticeLocation); ok {
    notice:=s.GetShard(location).(raftapp.Notice)
    // if _,ok=notice.(*raftapp.WrongGroupNotice); ok {
    //   DPrintf("location %+v get wrong group notice", location)
    // } else {
    //   DPrintf("location %+v get normal notice", location)
    // }
    return notice
  } else {
    return s.Notice
  }
}

func (s *ShardNotice) SetValue(location raftapp.NoticeLocation, seqNum int, value *raftapp.AsyncRequestReply) {
  s.getNotice(location).SetValue(location, seqNum, value)
}

func (s *ShardNotice) HasValue(location raftapp.NoticeLocation, seqNum int) bool {
  return s.getNotice(location).HasValue(location, seqNum)
}

func (s *ShardNotice) GetValue(location raftapp.NoticeLocation, seqNum int) <-chan *raftapp.AsyncRequestReply {
  return s.getNotice(location).GetValue(location, seqNum)
}

func (s *ShardNotice) UpdateTermAndLeader(term int, isLeader bool) {
  s.Notice.UpdateTermAndLeader(term, isLeader)
  s.ForAll(func(i raftapp.CleanSnapshotable) {
    if i != nil {
      i.(raftapp.Notice).UpdateTermAndLeader(term, isLeader)
    }
  })
}

func (s *ShardNotice) GenerateSnapshot() raftapp.Snapshot {
  return raftapp.ValueToSnapshot((&raftapp.SnapshotGroup{}).SetSnapshots(
    s.ShardWrapper.GenerateSnapshot(),
    s.Notice.GenerateSnapshot(),
  ))
}

func (s *ShardNotice) ApplySnapshot(i raftapp.Snapshot) {
  var ii raftapp.SnapshotGroup
  raftapp.SnapshotToValue(i, &ii)
  s.ShardWrapper.ApplySnapshot(ii.Snapshots[0])
  s.Notice.ApplySnapshot(ii.Snapshots[1])
}

func (s *ShardNotice) Clean() {
  // TODO: when needed
}

func MakeShardNotice(me int) *ShardNotice {
  return &ShardNotice{
    Notice: single.MakeSingleNotice(),
    ShardWrapper: MakeShardWrapper(
      func(shard int) raftapp.CleanSnapshotable {
        return single.MakeSingleNotice()
    }, raftapp.MakeWrongGroupNotice(), me, "ShardNotice")}
}
