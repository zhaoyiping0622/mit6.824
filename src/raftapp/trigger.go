package raftapp

import "context"

type Trigger struct {
  done context.CancelFunc
  result *CommandReply
  term int
}
