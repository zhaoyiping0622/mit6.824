package raftapp

import (
	"sync"

	"6.824/labgob"
)

func init() {
  labgob.Register(map[int64]*Session{})
}

type Session struct {
  SeqNum int
  Value interface{}
}

type Trigger interface {
  UpdateSession(sessionId int64, session*Session)
  GetValue(sessionId int64, seqNum int) <-chan interface{}
  HasValue(sessionId int64, seqNum int) bool
  UpdateTerm(term int)
  Snapshotable
}

type Notice struct {
  seqNum int
  valueChan []chan<- interface{}
}

type TriggerImpl struct {
  mu sync.Mutex
  sessions map[int64]*Session
  notices map[int64]*Notice
  term int
}

func (t *TriggerImpl) UpdateSession(sessionId int64, session *Session) {
  t.mu.Lock()
  defer t.mu.Unlock()
  if s,ok:=t.sessions[sessionId];ok {
    if session.SeqNum <= s.SeqNum {
      return
    }
  }
  t.sessions[sessionId] = session
  if notice,ok:=t.notices[sessionId];ok {
    if notice.seqNum == session.SeqNum {
      for _,c:=range notice.valueChan {
        c<-session.Value
      }
      delete(t.notices, sessionId)
    }
  }
}

func (t *TriggerImpl) GetValue(sessionId int64, seqNum int) <-chan interface{} {
  t.mu.Lock()
  defer t.mu.Unlock()
  ret:=make(chan interface{}, 1)
  if s,ok:=t.sessions[sessionId];ok {
    if s.SeqNum == seqNum {
      ret<-s.Value
      return ret
    } else if s.SeqNum > seqNum {
      close(ret)
      return ret
    }
  }
  if _,ok:=t.notices[sessionId];!ok {
    t.notices[sessionId]=&Notice{
      seqNum: seqNum,
      valueChan: make([]chan<- interface{}, 0),
    }
  }
  t.notices[sessionId].valueChan=append(t.notices[sessionId].valueChan, ret)
  return ret
}

func (t *TriggerImpl) HasValue(sessionId int64, seqNum int) bool {
  t.mu.Lock()
  defer t.mu.Unlock()
  s,ok:=t.sessions[sessionId]
  return ok && s.SeqNum == seqNum
}

func (t *TriggerImpl) removeNotice(sessionId int64) {
  if notice,ok:=t.notices[sessionId];ok {
    for _,c:=range notice.valueChan {
      close(c)
    }
    delete(t.notices, sessionId)
  }
}

func (t *TriggerImpl) removeAllNotices() {
  for sessionId:=range t.notices {
    t.removeNotice(sessionId)
  }
}

func (t *TriggerImpl) UpdateTerm(term int) {
  t.mu.Lock()
  defer t.mu.Unlock()
  if term <= t.term {
    return
  }
  t.removeAllNotices()
}

func (t *TriggerImpl) GenerateSnapshot() interface{} {
  t.mu.Lock()
  defer t.mu.Unlock()
  ret:=make(map[int64]*Session)
  for sessionId,session:=range t.sessions {
    s:=Session{}
    s=*session
    ret[sessionId]=&s
  }
  return ret
}

func (t *TriggerImpl) ApplySnapshot(snapshot interface{}) {
  t.mu.Lock()
  defer t.mu.Unlock()
  t.removeAllNotices()
  t.term=0
  t.sessions=snapshot.(map[int64]*Session)
}

func MakeTrigger() Trigger {
  return &TriggerImpl{
    mu: sync.Mutex{},
    sessions: make(map[int64]*Session),
    notices: make(map[int64]*Notice),
    term: 0,
  }
}
