package raftapp

import "math/rand"

type Err int

const (
  WrongLeader	Err = iota
	Ok
	WrongGroup
)

func (e Err) String() string {
	switch e {
	case Ok:
		return "ok"
	case WrongLeader:
		return "wrong leader"
	case WrongGroup:
		return "wrong group"
	default:
		return "unknown error"
	}
}

type OutterRpcArgs struct {
	Location NoticeLocation
	SeqNum   int
	Command  interface{}
}

type OutterRpcReply = AsyncRequestReply

type RpcServer struct {
	me            int
	controller    RaftController
	rpcExecutorId []int64
}

func (s *RpcServer) Request(request *OutterRpcArgs, reply *OutterRpcReply) {
	args := &AsyncRequestArgs{
		SeqNum:  request.SeqNum,
		Command: request.Command,
		Location: &ExecutorLocationImpl{
			NoticeLocation: request.Location,
			ExecutorIds:    append(make([]int64,0), s.rpcExecutorId...),
		},
	}
  id:=rand.Int31()
  pr:=PrettyPrint(request)
  if s.me >= 1000 {
    DPrintf("id %v %v get request %+v command %T", id, s.me, pr, request.Command)
  }
	r := s.controller.AsyncRequest(args)
	reply.Err = r.Err
	reply.Result = r.Result
  if s.me >= 1000 {
	  DPrintf("id %v %v get request %+v command %T reply %+v result %+v", id, s.me, pr, request.Command, r, r.Result)
  }
}

func MakeRpcServer(me int, executorId []int64, controller RaftController) *RpcServer {
	return &RpcServer{
		me:            me,
		controller:    controller,
		rpcExecutorId: executorId,
	}
}
