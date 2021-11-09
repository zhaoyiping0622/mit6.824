package shardctrler

import (
	"6.824/labgob"
	"6.824/raftapp"
)

func init() {
  labgob.Register(&Join{})
  labgob.Register(&Leave{})
  labgob.Register(&Move{})
  labgob.Register(&Query{})
}

type Join struct {
  Servers map[int][]string
}

type Leave struct {
  GIDs []int
}

type Move struct {
  Shard int
  GID int
}

type Query struct {
  Num int
}

func (c *Join) Apply(app raftapp.RaftApp) *raftapp.CommandReply {
  ct:=getController(app)
  ne:=ct.getNewConfig()
  for k,v:=range c.Servers {
    if vv,ok:=ne.Groups[k];ok {
      ne.Groups[k] = append(vv, v...)
    } else {
      ne.Groups[k] = v
    }
  }
  ne.balance()
  return &raftapp.CommandReply{
    Err: raftapp.OK,
    Result: nil,
  }

}
func (c *Leave) Apply(app raftapp.RaftApp) *raftapp.CommandReply {
  ct:=getController(app)
  ne:=ct.getNewConfig()
  for i:=range c.GIDs {
    delete(ne.Groups, c.GIDs[i])
  }
  ne.balance()
  return &raftapp.CommandReply{
    Err: raftapp.OK,
    Result: nil,
  }
}
func (c *Move) Apply(app raftapp.RaftApp) *raftapp.CommandReply {
  ct:=getController(app)
  ne:=ct.getNewConfig()
  ne.Shards[c.Shard]=c.GID
  return &raftapp.CommandReply{
    Err: raftapp.OK,
    Result: nil,
  }
}
func (c *Query) Apply(app raftapp.RaftApp) *raftapp.CommandReply {
  ct:=getController(app)
  if c.Num < 0 || c.Num >= len(ct.Configs) {
    c.Num = len(ct.Configs) - 1
  }
  return &raftapp.CommandReply{
    Err: raftapp.OK,
    Result: *ct.getConfig(c.Num),
  }
}
