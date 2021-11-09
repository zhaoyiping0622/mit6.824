package raftapp

import (
	"crypto/rand"
	"math/big"

	"6.824/labrpc"
)

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

type RaftClient struct {
  servers []*labrpc.ClientEnd
  sessionId int64
  seqNum int
  lastLeader int
  name string
}

func (cli *RaftClient) Send(command Command) (bool, interface{}) {
  DPrintf("client %v get command %T%+v", cli.sessionId, command, command)
  leader:=cli.lastLeader
  cli.seqNum++
  defer func() { cli.lastLeader = leader }()
  args:=CommandArgs{
    SessionId: cli.sessionId,
    SeqNum: cli.seqNum,
    Command: command,
  }
  for {
    if leader == len(cli.servers) {
      leader = 0
    }
    var reply CommandReply
    ok := cli.servers[leader].Call(cli.name+".CommandRequest", &args, &reply)
    if ok {
      switch reply.Err {
      case OK: return true,reply.Result
      case ErrWrongGroup: return false,nil
      case ErrWrongLeader: leader++
      default: leader++
      }
    } else {
      leader++
    }
  }
}

func MakeRaftClient(servers []*labrpc.ClientEnd, name string) *RaftClient {
  cli := &RaftClient{
    servers: servers,
    name: name,
    sessionId: nrand(),
  }
  DPrintf("client %p sessionId %v", cli, cli.sessionId )
  return cli
}
