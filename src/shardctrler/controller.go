package shardctrler

import (
	"fmt"

	"6.824/singleRaftapp"
)

type Controller struct {
  Configs []Config
  me int
}

func MakeController(me int) *Controller {
  rt:=new(Controller)
  rt.me=me
  rt.Configs=make([]Config, 1)
  rt.Configs[0].Groups=make(map[int][]string)
  return rt
}

func (ct *Controller) getLastConfig() *Config {
  return ct.getConfig(len(ct.Configs)-1)
}

// deep copy the config
func (ct *Controller) getConfig(idx int) *Config {
  raw:=ct.Configs[idx]
  newmap:=make(map[int][]string)
  for k,v:=range raw.Groups {
    s:=make([]string, len(v))
    copy(s, v)
    newmap[k]=s
  }
  raw.Groups=newmap
  return &raw
}

func (ct *Controller) getNewConfig() *Config {
  ret:=ct.getLastConfig()
  ret.Num++
  ct.Configs = append(ct.Configs, *ret)
  return &ct.Configs[len(ct.Configs)-1]
}

func (ct *Controller) ApplySnapshot(interface{}) {
  panic("controller do not support snapshot")
}

func (ct *Controller) GenerateSnapshot() interface{} {
  panic("controller do not support snapshot")
}

func (ct *Controller) ApplyCommand(command interface{}) interface{} {
  switch command:=command.(type) {
  case *Join:
    ne:=ct.getNewConfig()
    for k,v:=range command.Servers {
      if vv,ok:=ne.Groups[k];ok {
        ne.Groups[k] = append(vv, v...)
      } else {
        ne.Groups[k] = v
      }
    }
    ne.balance()
    return nil
  case *Leave:
    ne:=ct.getNewConfig()
    for i:=range command.GIDs {
      delete(ne.Groups, command.GIDs[i])
    }
    ne.balance()
    return nil
  case *Move:
    ne:=ct.getNewConfig()
    ne.Shards[command.Shard]=command.GID
    return nil
  case *Query:
    if command.Num < 0 || command.Num >= len(ct.Configs) {
      command.Num = len(ct.Configs) - 1
    }
    return *ct.getConfig(command.Num)
  default:
    panic(fmt.Sprintf("unknown command %T%+v",command,command))
  }
}

func getController(i singleRaftapp.RaftApp) *Controller {
  return i.(*Controller)
}
