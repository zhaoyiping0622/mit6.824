package singleRaftapp

import (
	"context"
	"time"

	"6.824/raft"
)

type Ticker interface {
  Run(ctx context.Context)
}

type StatusTicker struct {
  rf *raft.Raft
  applier Applier
  term int
  me int
}

func MakeStatusTicker(rf *raft.Raft, applier Applier, me int) *StatusTicker {
  return &StatusTicker{
    rf: rf,
    applier: applier,
    me: me,
  }
}

func (st *StatusTicker)Run(ctx context.Context) {
  for {
    select {
    case <-ctx.Done():return
    case <-time.After(10*time.Millisecond):
    }
    term,isLeader:=st.rf.GetState()
    if term != st.term {
      if isLeader {
        DPrintf("%v change to leader", st.me)
      }
      st.term = term
      st.applier.UpdateTerm(term)
    }
  }
}

type SnapshotTicker struct {
  persister *raft.Persister
  applier Applier
  limit int
}

func MakeSnapshotTicker(persister *raft.Persister, applier Applier, limit int) *SnapshotTicker {
  return &SnapshotTicker{
    persister: persister,
    applier: applier,
    limit: limit,
  }
}

func (st *SnapshotTicker) Run(ctx context.Context) {
  if st.limit == -1 {
    return
  }
  for {
    select {
    case <-ctx.Done():return
    case <-time.After(10*time.Millisecond):
    }
    if st.persister.RaftStateSize() >= st.limit {
      st.applier.Snapshot(ctx)
    }
  }
}
