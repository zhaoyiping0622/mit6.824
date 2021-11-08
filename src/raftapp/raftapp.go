package raftapp

type RaftApp interface {
  ApplySnapshot(interface{})
  CreateSnapshot() interface{}
}

