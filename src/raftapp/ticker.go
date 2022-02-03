package raftapp

import (
	"context"
	"time"
)

type TickerFunc = func(ctx context.Context)

type Ticker interface {
	SetFunc(TickerFunc) Ticker
	SetDuration(duration time.Duration) Ticker
	Run(ctx context.Context)
}

type TickerImpl struct {
	f TickerFunc
	d time.Duration
}

func (t *TickerImpl) SetFunc(f TickerFunc) Ticker { t.f = f; return t }

func (t *TickerImpl) SetDuration(d time.Duration) Ticker { t.d = d; return t }

func (t *TickerImpl) Run(ctx context.Context) {
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
