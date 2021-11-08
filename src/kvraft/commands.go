package kvraft

import (
	"6.824/labgob"
	"6.824/raftapp"
)

func init() {
  labgob.Register(GetCommand{})
  labgob.Register(PutCommand{})
  labgob.Register(AppendCommand{})
}

type GetCommand struct {
  Key string
}


func(c *GetCommand) Apply(app raftapp.RaftApp) *raftapp.CommandReply {
  kvs:=getKVStore(app)
  if value,ok:=kvs.Data[c.Key]; ok {
    return &raftapp.CommandReply{
      Err: raftapp.OK,
      Result: value,
    }
  } else {
    return &raftapp.CommandReply{
      Err: raftapp.OK,
      Result: "",
    }
  }
}

type PutCommand struct {
  Key string
  Value string
}

func (c *PutCommand) Apply(app raftapp.RaftApp) *raftapp.CommandReply {
  kvs:=getKVStore(app)
  kvs.Data[c.Key]=c.Value
  return &raftapp.CommandReply{
    Err: raftapp.OK,
  }
}

type AppendCommand struct {
  Key string
  Value string
}

func (c *AppendCommand) Apply(app raftapp.RaftApp) *raftapp.CommandReply {
  kvs:=getKVStore(app)
  if _,ok:=kvs.Data[c.Key]; ok {
    kvs.Data[c.Key]+=c.Value
  } else {
    kvs.Data[c.Key]=c.Value
  }
  return &raftapp.CommandReply{
    Err: raftapp.OK,
  }
}
