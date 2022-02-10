package shard

import (
	"6.824/labrpc"
	"6.824/raftapp/single"
)

type InnerClient struct {
  *single.SingleRaftClient
}

func MakeInnerClient(servers []*labrpc.ClientEnd, name string) *InnerClient {
  ic:=new(InnerClient)
  ic.SingleRaftClient=single.MakeSingleRaftClient(servers, name+".ShardRpcRequest")
  return ic
}

