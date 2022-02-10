package raftapp

import (
	"bytes"

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

func ValueToSnapshot(x interface{}) Snapshot {
  buffer:=new(bytes.Buffer)
  encoder:=labgob.NewEncoder(buffer)
  err:=encoder.Encode(x)
  if err!=nil {
    panic(err)
  }
  return buffer.Bytes()
}

func SnapshotToValue(b Snapshot, x interface{}) {
  decoder:=labgob.NewDecoder(bytes.NewBuffer(b))
  err:=decoder.Decode(x)
  if err!=nil {
    panic(err)
  }
}
