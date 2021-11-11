package raftapp

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"6.824/labgob"
	"6.824/raft"
)

type SingleSnapshot struct {
  App interface{}
  Trigger interface{}
}

type SingleApplier struct {
  mu sync.Mutex
  me int
  applyCh <-chan raft.ApplyMsg
  app RaftApp
  lastApplied int
  rf *raft.Raft
  trigger Trigger
  snapshotFinish context.CancelFunc
  snapshotCtx context.Context
}

func (sa *SingleApplier) CommandRequest(args *CommandRequestArgs,reply *CommandRequestReply)  {
  metadata:=getSingleCommandRequestMetatadata(args.MetaData)
  ch:=sa.trigger.GetValue(metadata.SessionId, metadata.SeqNum)
  var value interface{}
  var ok bool
  select {
    case value,ok=<-ch:
    default:
      _,_,isLeader:=sa.rf.Start(&Op{
        Metadata: metadata,
        Command: args.Command,
      })
      if !isLeader {
        break
      }
      value,ok=<-ch
  }
  if ok {
    reply.Err = OK
    reply.Result = value
  } else {
    reply.Err = ErrWrongLeader
  }
}

func (sa *SingleApplier) UpdateTerm(term int) {
  sa.trigger.UpdateTerm(term)
}

func (sa *SingleApplier) applyCommand(msg raft.ApplyMsg) {
  sa.mu.Lock()
  defer sa.mu.Unlock()
  DPrintf("%v get msg %+v", sa.me, msg)
  if msg.CommandIndex <= sa.lastApplied {
    return
  } else if msg.CommandIndex > sa.lastApplied + 1 {
    panic(fmt.Sprintf("%v CommandIndex %v lastApplied %v", sa.me, msg.CommandIndex, sa.lastApplied))
  }
  sa.lastApplied++
  op:=msg.Command.(*Op)
  metadata:=getSingleCommandRequestMetatadata(op.Metadata)
  if !sa.trigger.HasValue(metadata.SessionId, metadata.SeqNum) {
    result:=sa.app.ApplyCommand(op.Command)
    sa.trigger.UpdateSession(metadata.SessionId, &Session{
      SeqNum: metadata.SeqNum,
      Value: result,
    })
  }
}

func (sa *SingleApplier) applySnapshot(msg raft.ApplyMsg) {
  sa.mu.Lock()
  defer sa.mu.Unlock()
  DPrintf("%v get msg %+v", sa.me, msg)
  if sa.rf.CondInstallSnapshot(msg.SnapshotTerm, msg.SnapshotIndex, msg.Snapshot) && sa.lastApplied < msg.SnapshotIndex {
    var snapshot SingleSnapshot
    buffer:=bytes.NewBuffer(msg.Snapshot)
    decoder:=labgob.NewDecoder(buffer)
    decoder.Decode(&snapshot)
    sa.trigger.ApplySnapshot(snapshot.Trigger)
    sa.app.ApplySnapshot(snapshot.App)
    sa.lastApplied=msg.SnapshotIndex
  }
  if sa.snapshotFinish != nil {
    sa.snapshotFinish()
  }
}

func (sa *SingleApplier) Run(ctx context.Context) {
  for {
    select {
    case <-ctx.Done():return
    case msg,ok:=<-sa.applyCh:
      if ok {
        if msg.CommandValid {
          sa.applyCommand(msg)
        } else if msg.SnapshotValid {
          sa.applySnapshot(msg)
        } else {
          DPrintf("%v get invalid msg %+v", sa.me, msg)
          continue
        }
      } else {
        DPrintf("%v applyCh closed", sa.me)
        return
      }
    }
  }
}

func (sa *SingleApplier) Snapshot(ctx context.Context) {
  sa.mu.Lock()
  snapshot:=SingleSnapshot{
    App: sa.app.GenerateSnapshot(),
    Trigger: sa.trigger.GenerateSnapshot(),
  }
  DPrintf("snapshot %+v", snapshot)
  buffer:=new(bytes.Buffer)
  encoder:=labgob.NewEncoder(buffer)
  encoder.Encode(snapshot)
  index:=sa.lastApplied
  if sa.rf.Snapshot(index, buffer.Bytes()) {
    sa.snapshotCtx,sa.snapshotFinish = context.WithCancel(ctx)
    sa.mu.Unlock()
    <-sa.snapshotCtx.Done()
  } else {
    sa.mu.Unlock()
  }
}

func MakeSingleApplier(me int, rf *raft.Raft, app RaftApp, applyCh <-chan raft.ApplyMsg) *SingleApplier {
  return &SingleApplier{
    me: me,
    applyCh: applyCh,
    app: app,
    rf: rf,
    trigger: MakeTrigger(),
  }
}
