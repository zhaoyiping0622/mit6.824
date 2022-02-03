package raftapp

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"6.824/labgob"
	"6.824/raft"
)

func init() {
	labgob.Register(&Op{})
}

type Op = AsyncRequestArgs

func init() {
	labgob.Register(&AsyncRequestArgs{})
	labgob.Register(&AsyncRequestReply{})
}

type RaftController interface {
	// communicate with notice
	// not block controller
	// check group and leader
	AsyncRequest(*AsyncRequestArgs) *AsyncRequestReply
	// block controller
	// not check group and leader
	SyncRequest(*SyncRequestArgs) *SyncRequestReply
	Run(ctx context.Context)
}

type AsyncRequestArgs struct {
	Location Location
	SeqNum   int
	Command  interface{}
}

type AsyncRequestReply struct {
	Err    Err
	Result interface{}
}

type SyncRequestArgs struct {
	Command interface{}
}

type SyncRequestReply struct {
}

type SnapshotRequest struct {
	Ctx context.Context
}
type TermRequest struct {
	Term     int
	IsLeader bool
}

type RaftControllerImpl struct {
	mu             sync.Mutex
	me             int
	applyCh        <-chan raft.ApplyMsg
	lastApplied    int
	rf             *raft.Raft
	notice         Notice
	snapshotFinish context.CancelFunc
	snapshotCtx    context.Context
	term           int
	isLeader       bool
	executors      []Executor
	termNotice     []TermNotice
	snapshotables  []Snapshotable
	ctx            context.Context
}

func (s *RaftControllerImpl) Run(ctx context.Context) {
	s.ctx = ctx
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-s.applyCh:
			if ok {
				if msg.CommandValid {
					s.applyCommand(msg)
				} else if msg.SnapshotValid {
					s.applySnapshot(msg)
				} else {
					DPrintf("%v get invalid msg %+v", s.me, msg)
					continue
				}
			} else {
				DPrintf("%v applyCh closed", s.me)
				return
			}
		}
	}
}

func (s *RaftControllerImpl) applyCommand(msg raft.ApplyMsg) {
	s.mu.Lock()
	defer s.mu.Unlock()
	DPrintf("%v get msg %+v", s.me, msg)
	if msg.CommandIndex <= s.lastApplied {
		return
	} else if msg.CommandIndex > s.lastApplied+1 {
		panic(fmt.Sprintf("%v CommandIndex %v lastApplied %v", s.me, msg.CommandIndex, s.lastApplied))
	}
	s.lastApplied++
	op := msg.Command.(*Op)
	if !s.notice.HasValue(op.Location, op.SeqNum) {
		// DPrintf("%v apply Command %T%+v", s.me, op.Command, op.Command)
		wg := sync.WaitGroup{}
		wg.Add(len(s.executors))
		for _, e := range s.executors {
			go func(e Executor) {
				defer wg.Done()
				e.RunCommand(op)
			}(e)
		}
		wg.Wait()
	}
}

func (s *RaftControllerImpl) applySnapshot(msg raft.ApplyMsg) {
	s.mu.Lock()
	defer s.mu.Unlock()
	DPrintf("%v get msg %+v", s.me, msg)
	if s.rf.CondInstallSnapshot(msg.SnapshotTerm, msg.SnapshotIndex, msg.Snapshot) && s.lastApplied < msg.SnapshotIndex {
		var snapshot []interface{}
		buffer := bytes.NewBuffer(msg.Snapshot)
		decoder := labgob.NewDecoder(buffer)
		decoder.Decode(&snapshot)
		wg := sync.WaitGroup{}
		wg.Add(len(s.snapshotables))
		for i, e := range s.snapshotables {
			go func(i int, e Snapshotable) {
				defer wg.Done()
				e.ApplySnapshot(snapshot[i])
			}(i, e)
		}
		wg.Wait()
		s.lastApplied = msg.SnapshotIndex
	}
	if s.snapshotFinish != nil {
		s.snapshotFinish()
	}
}

func (s *RaftControllerImpl) AsyncRequest(args *AsyncRequestArgs) *AsyncRequestReply {
	ch := s.notice.GetValue(args.Location, args.SeqNum)
	var value interface{}
	var ok bool
	select {
	case value, ok = <-ch:
	default:
		_, _, isLeader := s.rf.Start(args)
		if !isLeader {
			break
		}
		select {
		case value, ok = <-ch:
		case <-s.ctx.Done():
			return &AsyncRequestReply{Err: WrongLeader}
		}
	}
	if ok {
		return value.(*AsyncRequestReply)
	} else {
		return &AsyncRequestReply{
			Err: WrongLeader,
		}
	}
}

func (s *RaftControllerImpl) SyncRequest(args *SyncRequestArgs) *SyncRequestReply {
	switch r := args.Command.(type) {
	case *SnapshotRequest:
		return s.snapshot(r)
	case *TermRequest:
		return s.updateTerm(r)
	default:
		panic("unknown sync request")
	}
}

// no need to lock
func (s *RaftControllerImpl) updateTerm(r *TermRequest) *SyncRequestReply {
	wg := sync.WaitGroup{}
	wg.Add(len(s.termNotice))
	for _, n := range s.termNotice {
		go func(n TermNotice) {
			defer wg.Done()
			n.UpdateTermAndLeader(r.Term, r.IsLeader)
		}(n)
	}
	wg.Wait()
	return &SyncRequestReply{}
}

func (s *RaftControllerImpl) snapshot(r *SnapshotRequest) *SyncRequestReply {
	s.mu.Lock()
	snapshot := make([]interface{}, len(s.snapshotables))
	wg := sync.WaitGroup{}
	wg.Add(len(s.snapshotables))
	for i, e := range s.snapshotables {
		go func(i int, e Snapshotable) {
			defer wg.Done()
			snapshot[i] = e.GenerateSnapshot()
		}(i, e)
	}
	wg.Wait()
	DPrintf("snapshot %+v", snapshot)
	buffer := new(bytes.Buffer)
	encoder := labgob.NewEncoder(buffer)
	encoder.Encode(snapshot)
	index := s.lastApplied
	if s.rf.Snapshot(index, buffer.Bytes()) {
		s.snapshotCtx, s.snapshotFinish = context.WithCancel(r.Ctx)
		s.mu.Unlock()
		<-s.snapshotCtx.Done()
	} else {
		s.mu.Unlock()
	}
	return &SyncRequestReply{}
}

func (c *RaftControllerImpl) Init(me int, applyCh <-chan raft.ApplyMsg, rf *raft.Raft, notice Notice, termNotices []TermNotice, executors []Executor, snapshotables []Snapshotable) {
	c.me = me
	c.applyCh = applyCh
	c.rf = rf
	c.notice = notice
	c.termNotice = termNotices
	c.executors = executors
	c.snapshotables = snapshotables
}
