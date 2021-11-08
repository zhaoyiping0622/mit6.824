package raftapp

type Session struct {
  SeqNum int
  Result *CommandReply
}
