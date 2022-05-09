package raftapp

import (
	"crypto/rand"
	"math/big"
	"strings"
	"time"

	"6.824/labrpc"
)

type RaftClient struct {
	servers    []*labrpc.ClientEnd
	seqNum     int
	lastLeader int
	name       string
}

func (cli *RaftClient) SendWithShard(command interface{}, location NoticeLocation, inc bool) (bool, interface{}) {
	leader := cli.lastLeader
	if inc {
		cli.seqNum++
	}
	defer func() { cli.lastLeader = leader }()
	args := &OutterRpcArgs{
		Location: location,
		SeqNum:   cli.seqNum,
		Command:  command,
	}
	if len(cli.servers) == 0 {
		return false, nil
	}
	for {
		if leader >= len(cli.servers) {
			leader = 0
		}
		var reply OutterRpcReply
		ok := cli.servers[leader].Call(cli.name, args, &reply)
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

func (cli *RaftClient) SetServer(servers []*labrpc.ClientEnd) {
	cli.servers = servers
}

func MakeRaftClient(servers []*labrpc.ClientEnd, name string) *RaftClient {
	if !strings.Contains(name, ".") {
		name = name + ".Request"
	}
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
