package raftapp

import (
	"fmt"

	"6.824/labgob"
)

func init() {
	labgob.Register(&SessionLocation{})
	labgob.Register(&ExecutorLocationImpl{})
}

type ExecutorLocation interface {
	GetExecutorIds() []int64
}

type NoticeLocation interface {
	GetSessionId() int64
}

type Location interface {
	ExecutorLocation
	NoticeLocation
}

type SessionLocation struct {
	SessionId int64
}

func (l *SessionLocation) String() string {
	return fmt.Sprintf("{SessionId: %v}", l.SessionId)
}

func (l *SessionLocation) GetSessionId() int64 {
	return l.SessionId
}

type ExecutorLocationImpl struct {
	NoticeLocation
	ExecutorIds []int64
}

func (l *ExecutorLocationImpl) GetExecutorIds() []int64 {
	return l.ExecutorIds
}

func MakeLocation(sessionId int64, executorId ...int64) Location {
  return &ExecutorLocationImpl{
    NoticeLocation: &SessionLocation{sessionId},
    ExecutorIds: executorId,
  }
}
