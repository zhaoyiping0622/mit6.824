package single

import (
	"sync"

	"6.824/labgob"
	"6.824/raftapp"
)

func init() {
  labgob.Register(&SingleNoticeLocation{})
  labgob.Register(&Session{})
  labgob.Register(map[int64]*Session{})
}

type SingleNoticeLocation struct {
  SessionId int64
}

func (s *SingleNoticeLocation) GetSessionId() int64 {
  return s.SessionId
}

type Session struct {
  SeqNum int
  Value *raftapp.AsyncRequestReply
}

type NoticeWatingQueue struct {
  SeqNum int
  queue []chan<- *raftapp.AsyncRequestReply
}

type SingleNotice struct {
  mu sync.Mutex
  sessions map[int64]*Session
  queues map[int64]*NoticeWatingQueue
  term int
}

func (s *SingleNotice) SetValue(location raftapp.NoticeLocation, seqNum int, value *raftapp.AsyncRequestReply){
  s.mu.Lock()
  defer s.mu.Unlock()
  sessionId:=location.GetSessionId()
  var session *Session
  if ss,ok:=s.sessions[sessionId]; ok {
    session=ss
  } else {
    session=&Session{}
    s.sessions[sessionId]=session
  }
  if session.SeqNum > seqNum {
    return
  }
  session.SeqNum=seqNum
  session.Value=value
  if qs,ok:=s.queues[sessionId]; ok {
    if qs.SeqNum == seqNum {
      for q:= range qs.queue {
        qs.queue[q]<-value
      }
      delete(s.queues, sessionId)
    }
  }
}

func (s *SingleNotice) GetValue(location raftapp.NoticeLocation, seqNum int) <-chan *raftapp.AsyncRequestReply {
  s.mu.Lock()
  defer s.mu.Unlock()
  ret:=make(chan *raftapp.AsyncRequestReply, 1)
  sessionId := location.GetSessionId()
  if s,ok:=s.sessions[sessionId];ok {
    if s.SeqNum == seqNum {
      ret<-s.Value
      return ret
    } else if s.SeqNum > seqNum {
      close(ret)
      return ret
    }
  }
  if _,ok:=s.queues[sessionId];!ok {
    s.queues[sessionId]=&NoticeWatingQueue{
      SeqNum: seqNum,
      queue: make([]chan<-*raftapp.AsyncRequestReply, 0),
    }
  }
  s.queues[sessionId].queue=append(s.queues[sessionId].queue, ret)
  return ret

}

func (s *SingleNotice) HasValue(location raftapp.NoticeLocation, seqNum int) bool {
  s.mu.Lock()
  defer s.mu.Unlock()
  ss,ok:=s.sessions[location.GetSessionId()]
  return ok && ss.SeqNum == seqNum
}


func (s *SingleNotice) removeQueue(sessionId int64) {
  if notice,ok:=s.queues[sessionId];ok {
    for _,c:=range notice.queue {
      close(c)
    }
    delete(s.queues, sessionId)
  }
}

func (s *SingleNotice) removeAllQueue() {
  for sessionId:=range s.queues {
    s.removeQueue(sessionId)
  }
}

func (s *SingleNotice) UpdateTermAndLeader(term int, isLeader bool) {
  s.mu.Lock()
  defer s.mu.Unlock()
  if term <= s.term {
    return
  }
  s.term=term
  s.removeAllQueue()
}


func (s *SingleNotice) GenerateSnapshot() interface{} {
  s.mu.Lock()
  defer s.mu.Unlock()
  ret:=make(map[int64]*Session)
  for sessionId,session:=range s.sessions {
    s:=Session{}
    s=*session
    ret[sessionId]=&s
  }
  return ret
}

func (s *SingleNotice) ApplySnapshot(snapshot interface{}) {
  s.mu.Lock()
  defer s.mu.Unlock()
  s.removeAllQueue()
  s.term=0
  s.sessions=snapshot.(map[int64]*Session)
}

func MakeSingleNotice() raftapp.Notice {
  return &SingleNotice{
    sessions: make(map[int64]*Session),
    queues: make(map[int64]*NoticeWatingQueue),
  }
}
