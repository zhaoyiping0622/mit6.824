package raftapp

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

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
	isLeader       int32
	executors      []Executable
	termNotice     []TermNotice
	snapshotables  []Snapshotable
	ctx            context.Context
	inited         int32
}

func (s *RaftControllerImpl) SetInited(inited bool) {
	if inited {
		atomic.StoreInt32(&s.inited, 1)
	}
}

func (s *RaftControllerImpl) GetInited() bool {
	return atomic.LoadInt32(&s.inited) == 1
}

func (s *RaftControllerImpl) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-s.applyCh:
			var copyMsg raft.ApplyMsg
			DeepCopy(msg, &copyMsg)
			if ok {
				if copyMsg.CommandValid {
					s.applyCommand(copyMsg)
				} else if copyMsg.SnapshotValid {
					s.applySnapshot(copyMsg)
				} else {
					DPrintf("%v get invalid msg %+v", s.me, copyMsg)
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
	// DPrintf("%v get msg %+v", s.me, PrettyPrint(msg))
	if msg.CommandIndex <= s.lastApplied {
		return
	} else if msg.CommandIndex > s.lastApplied+1 {
		panic(fmt.Sprintf("%v CommandIndex %v lastApplied %v", s.me, msg.CommandIndex, s.lastApplied))
	}
	s.lastApplied++
	op := msg.Command.(*Op)

	// DPrintf("%v apply Command %T%+v", s.me, op.Command, op.Command)
	wg := sync.WaitGroup{}
	wg.Add(len(s.executors))
	for _, e := range s.executors {
		go func(e Executable) {
			defer wg.Done()
			e.RunCommand(op)
		}(e)
	}
	wg.Wait()
}

func (s *RaftControllerImpl) applySnapshot(msg raft.ApplyMsg) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// DPrintf("%v get msg %+v", s.me, PrettyPrint(msg))
	if s.rf.CondInstallSnapshot(msg.SnapshotTerm, msg.SnapshotIndex, msg.Snapshot) && s.lastApplied < msg.SnapshotIndex {
		// TODO: unzip msg.Snapshot
		var snapshot SnapshotGroup
		SnapshotToValue(UnzipData(msg.Snapshot), &snapshot)
		wg := sync.WaitGroup{}
		wg.Add(len(s.snapshotables))
		for i, e := range s.snapshotables {
			go func(i int, e Snapshotable) {
				defer wg.Done()
				e.ApplySnapshot(snapshot.Snapshots[i])
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
	var copyArgs AsyncRequestArgs
	DeepCopy(args, &copyArgs)
	ch := s.notice.GetValue(copyArgs.Location, copyArgs.SeqNum)
	var value interface{}
	var ok bool
	select {
	case value, ok = <-ch:
	default:
		_, _, isLeader := s.rf.Start(copyArgs)
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
	// DPrintf("%v update TermRequest %+v", s.me, *r)
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
	snapshot := SnapshotGroup{
		Snapshots: make([]raft.Snapshot, len(s.snapshotables)),
	}
	wg := sync.WaitGroup{}
	wg.Add(len(s.snapshotables))
	for i, e := range s.snapshotables {
		go func(i int, e Snapshotable) {
			defer wg.Done()
			snapshot.Snapshots[i] = e.GenerateSnapshot()
		}(i, e)
	}
	wg.Wait()
	// TODO: zip snapshot
	index := s.lastApplied
	if s.rf.Snapshot(index, ZipData(ValueToSnapshot(snapshot))) {
		s.snapshotCtx, s.snapshotFinish = context.WithCancel(r.Ctx)
		s.mu.Unlock()
		<-s.snapshotCtx.Done()
	} else {
		s.mu.Unlock()
	}
	return &SyncRequestReply{}
}

func (c *RaftControllerImpl) Init(ctx context.Context, me int, applyCh <-chan raft.ApplyMsg, rf *raft.Raft, notice Notice, termNotices []TermNotice, executors []Executable, snapshotables []Snapshotable) {
	c.ctx = ctx
	c.me = me
	c.applyCh = applyCh
	c.rf = rf
	c.notice = notice
	c.termNotice = termNotices
	c.executors = append(executors, MakeInitExecutor(c, notice))
	c.snapshotables = snapshotables
	sessionId := Nrand()
	seqNum := 0
	go DefaultTicker().SetTickerDuration(100 * time.Millisecond).SetTickerFunc(func(ctx context.Context) {
		if c.GetInited() {
			return
		}
		seqNum++
		c.AsyncRequest(&AsyncRequestArgs{
			Location: MakeLocation(sessionId, ExecutorIdInit),
			SeqNum:   seqNum,
			Command:  nil,
		})
	}).TickerRun(ctx)
}
