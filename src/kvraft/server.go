package kvraft

import (
	"context"

	"6.824/labrpc"
	"6.824/raft"
	"6.824/raftapp"
)

type KVServer struct {
  rf *raft.Raft
  *raftapp.SingleServer
}

type KillEvent struct {
	done context.CancelFunc
}
//
// servers[] contains the ports of the set of
// servers that will cooperate via Raft to
// form the fault-tolerant key/value service.
// me is the index of the current server in servers[].
// the k/v server should store snapshots through the underlying Raft implementation, which should call persister.SaveStateAndSnapshot() to
// atomically save the Raft state along with the snapshot.
// the k/v server should snapshot when Raft's saved state exceeds maxraftstate bytes,
// in order to allow Raft to garbage-collect its log. if maxraftstate is -1,
// you don't need to snapshot.
// StartKVServer() must return quickly, so it should start goroutines
// for any long-running work.
//
func StartKVServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int) *KVServer {
  kv:=new(KVServer)
  kv.SingleServer = raftapp.MakeSingleServer(servers, me, persister, maxraftstate, MakeKVStore())
  kv.rf=kv.GetRaft()

	return kv
}
