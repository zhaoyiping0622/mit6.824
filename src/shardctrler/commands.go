package shardctrler

import (
	"6.824/labgob"
)

func init() {
  labgob.Register(&Join{})
  labgob.Register(&Leave{})
  labgob.Register(&Move{})
  labgob.Register(&Query{})
  labgob.Register(&Info{})
  labgob.Register(&InfoResult{})
  labgob.Register(&ShardInfo{})
  labgob.Register(&UpdateOwner{})
  labgob.Register(&UpdateOwnerResult{})
  labgob.Register(map[int]ShardInfo{})
  labgob.Register(map[int]bool{})
}

type Join struct {
  Servers map[int][]string
}

type Leave struct {
  GIDs []int
}

type Move struct {
  Shard int
  GID int
}

type Query struct {
  Num int
}

type Info struct {}

type UpdateOwner struct {
  Shard int
  Infos *ShardInfo
}

type InfoResult struct {
  Version int
  Config *Config
  ShardInfos [NShards]ShardInfo
  Groups map[int][]string
}

type UpdateOwnerResult struct {
  *InfoResult
  Success bool
}
