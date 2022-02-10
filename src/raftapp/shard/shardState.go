package shard

type ShardState int

func (s ShardState) String() string {
  switch s {
  case MOVED: return "moved"
  case LOCKED: return "locked"
  case HERE: return "here"
  case WAITED: return "waited"
  default: return "unknown"
  }
}

// state:   HERE --> LOCKED --> MOVED <--> WAITED --> HERE
// version: v    --> v      --> v     <--> v'+1   --> v'+1  (v'==remote version>=v)
//                             v'=v          
//                   ----------------------->
// command:      lock       move   apply/move    enable
const (
  WAITED ShardState = iota
  HERE
  LOCKED
  MOVED
)
