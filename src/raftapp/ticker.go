package raftapp

import (
	"context"
	"time"

	"6.824/raft"
)

type TickerFunc = func(ctx context.Context)

type Ticker interface {
	SetTickerFunc(TickerFunc) Ticker
	SetTickerDuration(duration time.Duration) Ticker
	TickerRun(ctx context.Context)
}

type TickerImpl struct {
	f TickerFunc
	d time.Duration
}

func (t *TickerImpl) SetTickerFunc(f TickerFunc) Ticker { t.f = f; return t }

func (t *TickerImpl) SetTickerDuration(d time.Duration) Ticker { t.d = d; return t }

func (t *TickerImpl) TickerRun(ctx context.Context) {
	if t.f == nil {
		panic("set func and raft controller before run the ticker")
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(t.d):
		}
		t.f(ctx)
	}
}

func DefaultTicker() Ticker {
	return &TickerImpl{
		d: time.Millisecond * 50,
	}
}

func MakeTermTicker(rf *raft.Raft, me int, c RaftController) Ticker {
	term, isLeader := 0, false
	return DefaultTicker().SetTickerFunc(func(ctx context.Context) {
		termA, isLeaderA := rf.GetState()
		if termA != term || isLeaderA != isLeader {
			if isLeaderA {
				DPrintf("%v change to leader", me)
			}
			term = termA
			isLeader = isLeaderA
			c.SyncRequest(&SyncRequestArgs{
				Command: &TermRequest{
					Term:     term,
					IsLeader: isLeader,
				}})
		}
	}) // term ticker
}

func MakeSnapshotTicker(maxraftstate int, persister *raft.Persister, c RaftController) Ticker {
	return DefaultTicker().SetTickerFunc(func(ctx context.Context) {
		if maxraftstate == -1 {
			return
		}
		if persister.RaftStateSize() >= maxraftstate {
			c.SyncRequest(&SyncRequestArgs{Command: &SnapshotRequest{Ctx: ctx}})
		}
	})
}
