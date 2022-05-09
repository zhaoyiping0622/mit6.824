package shardkv

import (
	"6.824/kvraft"
	"6.824/labrpc"
	"6.824/raft"
	"6.824/raftapp"
	"6.824/raftapp/shard"
)

type ShardKV struct {
	rf *raft.Raft
	// Your definitions here.
	*shard.ShardServer
}

//
// the tester calls Kill() when a ShardKV instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (kv *ShardKV) Kill() {
	// Your code here, if desired.
	kv.ShardServer.Kill()
}

//
// servers[] contains the ports of the servers in this group.
//
// me is the index of the current server in servers[].
//
// the k/v server should store snapshots through the underlying Raft
// implementation, which should call persister.SaveStateAndSnapshot() to
// atomically save the Raft state along with the snapshot.
//
// the k/v server should snapshot when Raft's saved state exceeds
// maxraftstate bytes, in order to allow Raft to garbage-collect its
// log. if maxraftstate is -1, you don't need to snapshot.
//
// gid is this group's GID, for interacting with the shardctrler.
//
// pass ctrlers[] to shardctrler.MakeClerk() so you can send
// RPCs to the shardctrler.
//
// make_end(servername) turns a server name from a
// Config.Groups[gid][i] into a labrpc.ClientEnd on which you can
// send RPCs. You'll need this to send RPCs to other groups.
//
// look at client.go for examples of how to use ctrlers[]
// and make_end() to send RPCs to the group owning a specific shard.
//
// StartServer() must return quickly, so it should start goroutines
// for any long-running work.
//
func StartServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int, gid int, ctrlers []*labrpc.ClientEnd, make_end func(string) *labrpc.ClientEnd) *ShardKV {

	kv := new(ShardKV)
	kv.ShardServer = shard.MakeShardServer(servers, me, persister, maxraftstate, gid, ctrlers, make_end, "ShardKV", func(shard int) raftapp.InnerExecutor {
		return kvraft.MakeKvStore()
	})
	kv.rf = kv.ShardServer.GetRaft()

	return kv
}
