package shardctrler

import (
	"log"
	"os"
	"sort"
	"strings"

	"6.824/labgob"
)

var Debug bool

func init() {
  debug:=os.Getenv("debug")
  for _,s:=range strings.Split(debug, ",") {
     if s == "shardctrler" {
       Debug = true
       break
     }
  }
}

func DPrintf(format string, a ...interface{}) (n int, err error) {
  if Debug {
    log.Printf(format, a...)
  }
  return
}

//
// Shard controler: assigns shards to replication groups.
//
// RPC interface:
// Join(servers) -- add a set of groups (gid -> server-list mapping).
// Leave(gids) -- delete a set of groups.
// Move(shard, gid) -- hand off one shard from current owner to gid.
// Query(num) -> fetch Config # num, or latest config if num==-1.
//
// A Config (configuration) describes a set of replica groups, and the
// replica group responsible for each shard. Configs are numbered. Config
// #0 is the initial configuration, with no groups and all shards
// assigned to group 0 (the invalid group).
//
// You will need to add fields to the RPC argument structs.
//

// The number of shards.
const NShards = 10

func init() {
  labgob.Register(&Config{})
}

// A configuration -- an assignment of shards to groups.
// Please don't change this.
type Config struct {
	Num    int              // config number
	Shards [NShards]int     // shard -> gid
	Groups map[int][]string // gid -> servers[]
}

type ShardInfo struct {
  Owner int
  Version int
}

func (c *Config) balance() {
  if len(c.Groups) == 0 {
    return
  }
  // gid --> shards
  cnt:=make(map[int][]int)
  for k:=range c.Groups {
    cnt[k]=make([]int, 0)
  }
  orphan:=make([]int,0)
  for i,gid:=range c.Shards {
    if _,ok:=cnt[gid]; ok {
      cnt[gid]=append(cnt[gid], i)
    } else {
      orphan = append(orphan, i)
    }
  }

  avg:=len(c.Shards)/len(cnt)
  other:=len(c.Shards)%len(cnt)
  // gids
  in:=make([]int,0)
  out:=make([]int,0)
  for k,v:=range cnt {
    if len(v)<=avg {
      in=append(in, k)
    } else {
      out=append(out, k)
    }
  }
  sort.Ints(in)
  sort.Ints(out)
  for i,k:=range out {
    array:=cnt[k]
    var length int
    if i>=other {
      length = avg
    } else {
      length = avg+1
    }
    orphan = append(orphan, array[length:]...)
    cnt[k]=array[:length]
  }
  for _,k:=range in {
    array:=cnt[k]
    for len(array)<avg {
      array=append(array, orphan[len(orphan)-1])
      orphan=orphan[:len(orphan)-1]
    }
    cnt[k]=array
  }
  for _,k:=range in {
    if len(orphan) == 0 {
      break
    }
    cnt[k] = append(cnt[k], orphan[len(orphan)-1])
    orphan=orphan[:len(orphan)-1]
  }
  if len(orphan) != 0 {
    panic("error in balance")
  }
  for k,v:=range cnt {
    for _,shard:=range v {
      c.Shards[shard]=k
    }
  }
}
