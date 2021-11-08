package raftapp

type Op struct {
  SessionId int64
  SeqNum int
  Command Command
}
