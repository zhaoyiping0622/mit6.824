package raftapp

import (
	"crypto/rand"
	"math/big"
	"time"

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

func (cli *RaftClient) Send(command interface{}) (bool, interface{}) {
  return cli.SendWithShard(command, 0)
}

func (cli *RaftClient) SendWithShard(command interface{}, shard int) (bool, interface{}) {
  DPrintf("client %v get command %T%+v", cli.sessionId, command, command)
  leader:=cli.lastLeader
  cli.seqNum++
  defer func() { cli.lastLeader = leader }()
  args:=&CommandRequestArgs{
    MetaData: CommandRequestMetadata{
      SessionId: cli.sessionId,
      SeqNum: cli.seqNum,
      Shard: shard,
    },
    Command: command,
  }
  for {
    if leader == len(cli.servers) {
      leader = 0
    }
    var reply CommandRequestReply
    ok := cli.servers[leader].Call(cli.name+".CommandRequest", args, &reply)
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
    time.Sleep(10*time.Millisecond)
  }
}

func MakeSingleRaftClient(servers []*labrpc.ClientEnd, name string) *RaftClient {
  cli := &RaftClient{
    servers: servers,
    name: name,
    sessionId: nrand(),
  }
  DPrintf("client %p sessionId %v", cli, cli.sessionId )
  return cli
}
