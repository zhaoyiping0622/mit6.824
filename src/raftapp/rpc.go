package raftapp

import (
	"context"
)

type Err string

const (
  OK Err = "OK"
  ErrWrongLeader = "ErrWrongLeader"
  ErrWrongGroup = "ErrWrongGroup"
)

type CommandArgs struct {
  SessionId int64
  SeqNum int
  Command Command
}

type CommandReply struct {
  Err Err
  Result interface{}
}

func (rs *RaftServer) CommandRequest(args *CommandArgs, reply *CommandReply) {
  if args.Command == nil {
    panic(args)
  }
  DPrintf("%v get CommandRequest args %+v", rs.me, args)
  op:=Op{
    SessionId: args.SessionId,
    SeqNum: args.SeqNum,
    Command: args.Command,
  }
  _,term,isLeader:=rs.rf.Start(op)
  if !isLeader {
    reply.Err = ErrWrongLeader
    return
  }
  ctx,cancel:=context.WithCancel(rs.background)
  go rs.sendEvent(&CommandRequestEvent{
    sessionId: args.SessionId,
    seqNum: args.SeqNum,
    term: term,
    done: cancel,
    reply: reply,
  })
  <-ctx.Done()
}

type CommandRequestEvent struct {
  sessionId int64
  seqNum int
  term int
  done context.CancelFunc
  reply *CommandReply
}

func (e *CommandRequestEvent) Run(rs *RaftServer) {
  if e.term < rs.term {
    e.reply.Err = ErrWrongLeader
    if e.done != nil {
      e.done()
    }
    return
  }
  var session *Session
  if s,ok:=rs.sessions[e.sessionId]; ok {
    session = s
  } else {
    session = new(Session)
    rs.sessions[e.sessionId] = session
  }
  if session.SeqNum == e.seqNum {
    *e.reply = *session.Result
    if e.done != nil {
      e.done()
    }
    return
  }
  trigger:=&Trigger{
    done: e.done,
    result: e.reply,
    term: rs.term,
    seqNum: e.seqNum,
  }
  if t,ok:=rs.triggers[e.sessionId];ok {
    if t.done != nil {
      t.result.Err = ErrWrongLeader
      t.done()
    }
  }
  rs.triggers[e.sessionId]=trigger
}
