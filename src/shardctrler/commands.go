package shardctrler

import (
	"6.824/labgob"
)

func init() {
  labgob.Register(&Join{})
  labgob.Register(&Leave{})
  labgob.Register(&Move{})
  labgob.Register(&Query{})
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
