package single

import (
	"6.824/labrpc"
	"6.824/raftapp"
)

type SingleRaftClient struct {
  location raftapp.NoticeLocation
  *raftapp.RaftClient
}

func (cli *SingleRaftClient) Send(command interface{}) (bool, interface{}) {
  return cli.SendWithShard(command, 0, cli.location)
}

func MakeSingleRaftClient(servers []*labrpc.ClientEnd, name string) *SingleRaftClient {
  sessionId:=raftapp.Nrand()
  cli:=raftapp.MakeRaftClient(servers, name)
  DPrintf("client %p sessionId %v", cli, sessionId)
  return &SingleRaftClient{
    RaftClient: cli,
    location: &raftapp.SessionLocation{
      SessionId: sessionId,
    },
  }
}

