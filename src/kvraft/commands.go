package kvraft

import (
	"6.824/labgob"
)

func init() {
  labgob.Register(&GetCommand{})
  labgob.Register(&PutCommand{})
  labgob.Register(&AppendCommand{})
}

type GetCommand struct {
  Key string
}

type PutCommand struct {
  Key string
  Value string
}

type AppendCommand struct {
  Key string
  Value string
}
