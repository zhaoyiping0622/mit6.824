package shard

import (
	"6.824/labrpc"
	"6.824/raftapp"
	"6.824/shardctrler"
)

type ShardRaftClient struct {
  config *shardctrler.Config
  ctrlerCli *shardctrler.Clerk
  clis [NShards]*raftapp.RaftClient
  cliGroups [NShards][]string
  serverName string
  make_end func(string) *labrpc.ClientEnd
  location raftapp.NoticeLocation
}

func (cli *ShardRaftClient) Send(shard int, command interface{}) interface{} {
  inc:=true
  for {
    if cli.config != nil {
      ok,result:=cli.clis[shard].SendWithShard(command, &ShardNoticeLocation{
        NoticeLocation: cli.location,
        Shard: shard,
      }, inc)
      if ok {
        return result
      }
    }
    inc=false
    cli.config=cli.ctrlerCli.Info().Config
    gid:=cli.config.Shards[shard]
    group:=cli.config.Groups[gid]
    cli.cliGroups[shard]=group
    servers:=make([]*labrpc.ClientEnd, len(group))
    for i:=range group {
      servers[i]=cli.make_end(group[i])
    }
    cli.clis[shard].SetServer(servers)
  }
}

func MakeShardRaftClient(ctrlerServers []*labrpc.ClientEnd, name string, make_end func(string) *labrpc.ClientEnd) *ShardRaftClient {
  cli:=new(ShardRaftClient)
  sessionId:=raftapp.Nrand()
  cli.ctrlerCli=shardctrler.MakeClerk(ctrlerServers)
  for i:=0;i<NShards;i++ {
    cli.clis[i]=raftapp.MakeRaftClient([]*labrpc.ClientEnd{}, name)
    cli.cliGroups[i]=make([]string, 0)
  }
  cli.serverName=name
  cli.make_end=make_end
  cli.location=&raftapp.SessionLocation{SessionId: sessionId}
  return cli
}
