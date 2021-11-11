package raftapp

type Snapshotable interface {
  GenerateSnapshot() interface{}
  ApplySnapshot(interface{})
}
