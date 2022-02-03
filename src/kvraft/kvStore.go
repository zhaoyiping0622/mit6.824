package kvraft

import (
	"fmt"

	"6.824/labgob"
	"6.824/raftapp"
)

func init() {
	labgob.Register(&GET{})
	labgob.Register(&APPEND{})
	labgob.Register(&PUT{})
	labgob.Register(map[string]string{})
}

type GET struct {
	Key string
}

type APPEND struct {
	Key   string
	Value string
}

type PUT struct {
	Key   string
	Value string
}

type kvStore struct {
	m map[string]string
}

func (k *kvStore) Run(c interface{}) *raftapp.AsyncRequestReply {
	ret := new(raftapp.AsyncRequestReply)
	ret.Err = raftapp.Ok
	switch cc := c.(type) {
	case *GET:
		{
			kk := cc.Key
			if v, ok := k.m[kk]; ok {
				ret.Result = v
			} else {
				ret.Result = ""
			}
		}
	case *PUT:
		k.m[cc.Key] = cc.Value
	case *APPEND:
		{
			kk := cc.Key
			if v, ok := k.m[kk]; ok {
				k.m[kk] = v + cc.Value
			} else {
				k.m[kk] = cc.Value
			}
		}
	default:
		panic(fmt.Sprintf("unknown command %+v", c))
	}
	return ret
}

func (k *kvStore) GenerateSnapshot() interface{} {
	return k.m
}

func (k *kvStore) ApplySnapshot(s interface{}) {
	k.m = s.(map[string]string)
}

func MakeKvStore() *kvStore {
	return &kvStore{make(map[string]string)}
}
