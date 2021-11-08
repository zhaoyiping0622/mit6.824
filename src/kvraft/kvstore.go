package kvraft

import (
	"fmt"

	"6.824/labgob"
	"6.824/raftapp"
)

func init() {
  labgob.Register(KVStore{})
}

type KVStore struct {
  Data map[string]string
}

func (kvs *KVStore) ApplySnapshot(snapshot interface{}) {
  if ma,ok:=snapshot.(KVStore);ok {
    kvs.Data = ma.Data
  } else {
    panic(fmt.Sprintf("KVStore should apply snapshot with type KVStore, snapshot type %T", snapshot))
  }
}
func (kvs *KVStore) CreateSnapshot()interface{} {
  return kvs
}

func MakeKVStore() *KVStore {
  return &KVStore{
    Data: make(map[string]string),
  }
}

func getKVStore(i raftapp.RaftApp) *KVStore {
  return i.(*KVStore)
}
