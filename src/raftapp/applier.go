package raftapp

import (
	"context"

	"6.824/labgob"
)

func init() {
  labgob.Register(&Op{})
}

// from raft to app
type Applier interface {
  CommandRequest(args *CommandRequestArgs,reply *CommandRequestReply)
  Run(ctx context.Context)
  UpdateTerm(term int)
  // block function
  Snapshot(context.Context)
}

type Op struct {
  Metadata CommandRequestMetadata
  Command interface{}
}

