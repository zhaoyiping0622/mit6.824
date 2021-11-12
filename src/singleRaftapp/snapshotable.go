package singleRaftapp

type Snapshotable interface {
  GenerateSnapshot() interface{}
  ApplySnapshot(interface{})
}
