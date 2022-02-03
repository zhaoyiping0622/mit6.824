package raftapp

import (
	"crypto/rand"
	"math/big"
	"time"

	"6.824/labrpc"
)

type RaftClient struct {
	servers    []*labrpc.ClientEnd
	seqNum     int
	lastLeader int
	name       string
}

func (cli *RaftClient) SendWithShard(command interface{}, shard int, location NoticeLocation) (bool, interface{}) {
	leader := cli.lastLeader
	cli.seqNum++
	defer func() { cli.lastLeader = leader }()
	args := &OutterRpcArgs{
		Location: location,
		SeqNum:   cli.seqNum,
		Command:  command,
	}
	for {
		if leader == len(cli.servers) {
			leader = 0
		}
		var reply OutterRpcReply
		ok := cli.servers[leader].Call(cli.name+".Request", args, &reply)
		if ok {
			switch reply.Err {
			case Ok:
				return true, reply.Result
			case WrongGroup:
				return false, nil
			case WrongLeader:
				leader++
			default:
				leader++
			}
		} else {
			leader++
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func MakeRaftClient(servers []*labrpc.ClientEnd, name string) *RaftClient {
	return &RaftClient{
		servers: servers,
		name:    name,
	}
}

func Nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}
