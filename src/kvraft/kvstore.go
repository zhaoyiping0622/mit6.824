package kvraft

import (
	"fmt"

	"6.824/labgob"
	"6.824/singleRaftapp"
)

func init() {
  labgob.Register(&KVStore{})
  labgob.Register(map[string]string{})
}

type KVStore struct {
  Data map[string]string
}

func (kvs *KVStore) ApplySnapshot(snapshot interface{}) {
  if ma,ok:=snapshot.(*KVStore);ok {
    kvs.Data = ma.Data
  } else {
    panic(fmt.Sprintf("KVStore should apply snapshot with type KVStore, snapshot type %T", snapshot))
  }
}
func (kvs *KVStore) GenerateSnapshot()interface{} {
  return kvs
}

func (kvs *KVStore) ApplyCommand(command interface{}) interface{} {
  switch command:=command.(type) {
  case *GetCommand:
    if v,ok:=kvs.Data[command.Key]; ok {
      return v
    } else {
      return ""
    }
  case *PutCommand:
    kvs.Data[command.Key]=command.Value
    return nil
  case *AppendCommand:
    if _,ok:=kvs.Data[command.Key]; ok {
      kvs.Data[command.Key]+=command.Value
    } else {
      kvs.Data[command.Key]=command.Value
    }
    return nil
  default:
    panic(fmt.Sprintf("unknown command %T%+v", command, command))
  }
}

func MakeKVStore() *KVStore {
  return &KVStore{
    Data: make(map[string]string),
  }
}

func getKVStore(i singleRaftapp.RaftApp) *KVStore {
  return i.(*KVStore)
}
