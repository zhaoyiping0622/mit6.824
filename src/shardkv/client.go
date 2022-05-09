package shardkv

//
// client code to talk to a sharded key/value service.
//
// the client first talks to the shardctrler to find out
// the assignment of shards (keys) to groups, and then
// talks to the group that holds the key's shard.
//

import (
	"6.824/kvraft"
	"6.824/labrpc"
	"6.824/raftapp/shard"
	"6.824/shardctrler"
)

//
// which shard is a key in?
// please use this function,
// and please do not change it.
//
func key2shard(key string) int {
	shard := 0
	if len(key) > 0 {
		shard = int(key[0])
	}
	shard %= shardctrler.NShards
	return shard
}

type Clerk struct {
	*shard.ShardRaftClient
}

//
// the tester calls MakeClerk.
//
// ctrlers[] is needed to call shardctrler.MakeClerk().
//
// make_end(servername) turns a server name from a
// Config.Groups[gid][i] into a labrpc.ClientEnd on which you can
// send RPCs.
//
func MakeClerk(ctrlers []*labrpc.ClientEnd, make_end func(string) *labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.ShardRaftClient = shard.MakeShardRaftClient(ctrlers, "ShardKV", make_end)
	return ck
}

func (ck *Clerk) Get(key string) string {
	return ck.Send(key2shard(key), &kvraft.GET{
		Key: key,
	}).(string)
}

func (ck *Clerk) Put(key string, value string) {
	ck.Send(key2shard(key), &kvraft.PUT{
		Key:   key,
		Value: value,
	})
}
func (ck *Clerk) Append(key string, value string) {
	ck.Send(key2shard(key), &kvraft.APPEND{
		Key:   key,
		Value: value,
	})
}
