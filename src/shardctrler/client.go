package shardctrler

//
// Shardctrler clerk.
//

import (
	"6.824/labrpc"
	"6.824/raftapp/single"
)

type Clerk struct {
  *single.SingleRaftClient
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
  ck.SingleRaftClient = single.MakeSingleRaftClient(servers, "ShardCtrler")
	return ck
}

func (ck *Clerk) Query(num int) Config {
  return *ck.Send(&Query{
    Num: num,
  }).(*Config)
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

func (ck *Clerk) Info() *InfoResult {
  return ck.Send(&Info{}).(*InfoResult)
}

func (ck *Clerk) UpdateOwner(Infos *ShardInfo, Shard int) *UpdateOwnerResult {
  return ck.Send(&UpdateOwner{
    Infos:Infos,
    Shard: Shard,
  }).(*UpdateOwnerResult)
}
