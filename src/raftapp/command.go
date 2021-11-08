package raftapp

type Command interface {
  Apply(app RaftApp) *CommandReply
}

