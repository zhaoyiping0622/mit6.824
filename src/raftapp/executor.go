package raftapp

import (
	"6.824/labgob"
)

func init() {
	labgob.Register(&ExecutorImplSnapshot{})
}

type Executor interface {
	Snapshotable
	RunCommand(*AsyncRequestArgs)
	SetExecutorId(int64)
}

type InnerExecutor interface {
	Snapshotable
	Run(interface{}) interface{}
}

type ExecutorImpl struct {
	InnerExecutor InnerExecutor
	ExecutorId    int64
	Notice        NoticeProducer
}

type ExecutorImplSnapshot struct {
	Inner      interface{}
	ExecutorId int64
}

func (e *ExecutorImpl) SetExecutorId(id int64) {
	e.ExecutorId = id
}

func (e *ExecutorImpl) SetNotice(n NoticeProducer) {
	e.Notice = n
}

func (e *ExecutorImpl) RunCommand(r *AsyncRequestArgs) {
	for _, x := range r.Location.GetExecutorIds() {
		if x == e.ExecutorId {
      res := &AsyncRequestReply{
        Err: Ok,
        Result: e.InnerExecutor.Run(r.Command),
      }
			e.Notice.SetValue(r.Location, r.SeqNum, res)
		}
	}
}

func (e *ExecutorImpl) GenerateSnapshot() interface{} {
	s := &ExecutorImplSnapshot{ExecutorId: e.ExecutorId}
	s.Inner = e.InnerExecutor.GenerateSnapshot()
	return s
}

func (e *ExecutorImpl) ApplySnapshot(s interface{}) {
	ss := s.(*ExecutorImplSnapshot)
	e.ExecutorId = ss.ExecutorId
	e.InnerExecutor.ApplySnapshot(ss.Inner)
}
