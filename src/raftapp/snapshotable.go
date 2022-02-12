package raftapp

import (
	"6.824/labgob"
	"6.824/raft"
)

func init() {
  labgob.Register(&SnapshotGroup{})
}

type Snapshot=raft.Snapshot

type SnapshotGroup struct {
  Snapshots []Snapshot
}

func (s *SnapshotGroup) AddSnapshots(ss ...Snapshot) *SnapshotGroup {
  s.Snapshots=append(s.Snapshots, ss...)
  return s
}

func (s *SnapshotGroup) SetSnapshots(ss ...Snapshot) *SnapshotGroup {
  s.Snapshots=ss
  return s
}

type Snapshotable interface {
	GenerateSnapshot() Snapshot
	ApplySnapshot(Snapshot)
}

type Cleanable interface {
  Clean()
}

type CleanSnapshotable interface {
  Snapshotable
  Cleanable
}
