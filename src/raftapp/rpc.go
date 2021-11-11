package raftapp

import "6.824/labgob"

type Err string

func init() {
  labgob.Register(&SingleCommandRequestMetadata{})
}

const (
  OK Err = "OK"
  ErrWrongLeader = "ErrWrongLeader"
  ErrWrongGroup = "ErrWrongGroup"
)

type CommandRequestArgs struct {
  MetaData interface{}
  Command interface{}
}

type CommandRequestReply struct {
  Err Err
  Result interface{}
}

type SingleCommandRequestMetadata struct {
  SessionId int64
  SeqNum int
}

func getSingleCommandRequestMetatadata(i interface{}) *SingleCommandRequestMetadata {
  return i.(*SingleCommandRequestMetadata)
}
