package kvraft

import (
	"6.824/labrpc"
	"6.824/raftapp/single"
)

type Clerk struct {
	*single.SingleRaftClient
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.SingleRaftClient = single.MakeSingleRaftClient(servers, "KVServer")
	return ck
}

func (ck *Clerk) Get(key string) string {
	return ck.Send(&GET{
		Key: key,
	}).(string)
}

func (ck *Clerk) Put(key string, value string) {
	ck.Send(&PUT{
		Key:   key,
		Value: value,
	})
}
func (ck *Clerk) Append(key string, value string) {
	ck.Send(&APPEND{
		Key:   key,
		Value: value,
	})
}
