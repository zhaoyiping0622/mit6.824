package raftapp

type Err int

const (
	Ok Err = iota
	WrongLeader
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
	rpcExecutorId int64
}

func (s *RpcServer) Request(request *OutterRpcArgs, reply *OutterRpcReply) {
	args := &AsyncRequestArgs{
		SeqNum:  request.SeqNum,
		Command: request.Command,
		Location: &ExecutorLocationImpl{
			NoticeLocation: request.Location,
			ExecutorIds:    []int64{s.rpcExecutorId},
		},
	}
	DPrintf("%v get request %+v", s.me, request)
	r := s.controller.AsyncRequest(args)
	reply.Err = r.Err
	reply.Result = r.Result
	DPrintf("%v reply %+v", s.me, r.Result)
}

func MakeRpcServer(me int, executorId int64, controller RaftController) *RpcServer {
	return &RpcServer{
		me:            me,
		controller:    controller,
		rpcExecutorId: executorId,
	}
}
