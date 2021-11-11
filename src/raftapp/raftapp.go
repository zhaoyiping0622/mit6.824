package raftapp

type RaftApp interface {
  Snapshotable
  ApplyCommand(command interface{}) interface{}
}
