package shard

import (
	"6.824/labgob"
	"6.824/raftapp"
)

func init() {
  labgob.Register(&ShardNoticeLocation{})
}

type ShardNoticeLocation struct {
  raftapp.NoticeLocation
  Shard int
}

func (l *ShardNoticeLocation) GetShard() int {
  return l.Shard
}
