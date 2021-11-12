package singleRaftapp

type RaftApp interface {
  Snapshotable
  ApplyCommand(command interface{}) interface{}
}
