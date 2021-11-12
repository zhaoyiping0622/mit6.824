package kvraft

import (
	"crypto/rand"
	"math/big"

	"6.824/labrpc"
	"6.824/singleRaftapp"
)

type Clerk struct {
  *singleRaftapp.SingleRaftClient
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
  return &Clerk{
    singleRaftapp.MakeSingleRaftClient(servers, "KVServer"),
  }
}

func (ck *Clerk) Get(key string) string {
  _,ret:=ck.Send(&GetCommand{
    Key:key,
  })
  return ret.(string)
}

func (ck *Clerk) Put(key string, value string) {
  ck.Send(&PutCommand{
    Key: key,
    Value: value,
  })
}
func (ck *Clerk) Append(key string, value string) {
  ck.Send(&AppendCommand{
    Key: key,
    Value: value,
  })
}
