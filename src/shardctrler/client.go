package shardctrler

//
// Shardctrler clerk.
//

import (
	"6.824/labrpc"
	"6.824/raftapp"
)

type Clerk struct {
  *raftapp.RaftClient
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
  ck.RaftClient = raftapp.MakeRaftClient(servers, "ShardCtrler")
	return ck
}

func (ck *Clerk) Query(num int) Config {
  _,ret:=ck.Send(&Query{
    Num: num,
  })
  return ret.(Config)
}

func (ck *Clerk) Join(servers map[int][]string) {
  ck.Send(&Join{
    Servers: servers,
  })
}

func (ck *Clerk) Leave(gids []int) {
  ck.Send(&Leave{
    GIDs: gids,
  })
}

func (ck *Clerk) Move(shard int, gid int) {
  ck.Send(&Move{
    Shard: shard,
    GID: gid,
  })
}
