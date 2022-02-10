package raftapp

import (
	"6.824/labgob"
)

func init() {
	labgob.Register(&ExecutorImplSnapshot{})
}

const (
  ExecutorIdApp int64 = iota + 1
  ExecutorIdShard
  ExecutorIdInit
)

type Executable interface {
	RunCommand(*AsyncRequestArgs)
}

type Executor interface {
	CleanSnapshotable
  Executable
}

type InnerExecutor interface {
	CleanSnapshotable
	Run(interface{}) interface{}
}

type ExecutorImpl struct {
	InnerExecutor InnerExecutor
	ExecutorId    int64
	Notice        Notice
}

type ExecutorImplSnapshot struct {
	Inner      Snapshot
	ExecutorId int64
}

func (e *ExecutorImpl) SetNotice(n Notice) {
	e.Notice = n
}

func (e *ExecutorImpl) RunCommand(r *AsyncRequestArgs) {
  if e.Notice.HasValue(r.Location, r.SeqNum) {
    return
  }
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

func (e *ExecutorImpl) GenerateSnapshot() Snapshot {
	s := ExecutorImplSnapshot{
    ExecutorId: e.ExecutorId,
    Inner: e.InnerExecutor.GenerateSnapshot(),
  }
	return ValueToSnapshot(s)
}

func (e *ExecutorImpl) ApplySnapshot(s Snapshot) {
  if s != nil {
    x:=ExecutorImplSnapshot{}
    SnapshotToValue(s, &x)
    e.ExecutorId = x.ExecutorId
    e.InnerExecutor.ApplySnapshot(x.Inner)
  }
}

func (e *ExecutorImpl) Clean() {
  e.InnerExecutor.Clean()
}

func MakeExecutor(innerExecutor InnerExecutor, notice Notice, executorId int64) Executor {
  return &ExecutorImpl{
    InnerExecutor: innerExecutor,
    Notice: notice,
    ExecutorId: executorId,
  }
}

type TmpExecutor struct {
  f func(r *AsyncRequestArgs)
}

func (e *TmpExecutor) RunCommand(r *AsyncRequestArgs) { e.f(r) }

func (e *TmpExecutor) GenerateSnapshot() Snapshot { return nil }

func (e *TmpExecutor) ApplySnapshot(Snapshot) {}

func (e *TmpExecutor) Clean() {}

func MakeTmpExecutor(f func(r *AsyncRequestArgs)) Executor { return &TmpExecutor{ f: f } }

func MakeWrongGroupExecutor() Executor { return MakeTmpExecutor(func(r *AsyncRequestArgs) {}) }

func MakeInitExecutor(c *RaftControllerImpl, n Notice) Executor {
  return MakeTmpExecutor(func(r *AsyncRequestArgs) {
    for _,x:=range r.Location.GetExecutorIds() {
      if x == ExecutorIdInit{
        c.SetInited(true)
        DPrintf("%v inited", c.me)
        n.SetValue(r.Location, r.SeqNum, nil)
      }
    }
  })
}
