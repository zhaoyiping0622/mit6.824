package shardctrler

import (
	"6.824/labrpc"
	"6.824/raft"
	"6.824/singleRaftapp"
)


type ShardCtrler struct {
  rf *raft.Raft
  *singleRaftapp.SingleServer
}

// needed by shardkv tester
func (sc *ShardCtrler) Raft() *raft.Raft {
	return sc.rf
}

//
// servers[] contains the ports of the set of
// servers that will cooperate via Raft to
// form the fault-tolerant shardctrler service.
// me is the index of the current server in servers[].
//
func StartServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister) *ShardCtrler {
  sc:=new(ShardCtrler)
  sc.SingleServer = singleRaftapp.MakeSingleServer(servers, me, persister, -1, MakeController(me))
  sc.rf = sc.GetRaft()
  return sc
}
