package singleRaftapp

type Err string

const (
  OK Err = "OK"
  ErrWrongLeader = "ErrWrongLeader"
  ErrWrongGroup = "ErrWrongGroup"
)

type CommandRequestArgs struct {
  MetaData CommandRequestMetadata
  Command interface{}
}

type CommandRequestReply struct {
  Err Err
  Result interface{}
}

type CommandRequestMetadata struct {
  SessionId int64
  SeqNum int
  Shard int
}
