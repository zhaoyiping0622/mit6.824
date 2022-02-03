package shardctrler

import (
	"fmt"

	"6.824/raftapp"
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
  ret:=raw
  newmap:=make(map[int][]string)
  for k,v:=range raw.Groups {
    s:=make([]string, len(v))
    copy(s, v)
    newmap[k]=s
  }
  ret.Groups=newmap
  return &ret
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

func (ct *Controller) Run(command interface{}) interface{} {
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
  case *Leave:
    ne:=ct.getNewConfig()
    for i:=range command.GIDs {
      delete(ne.Groups, command.GIDs[i])
    }
    ne.balance()
  case *Move:
    ne:=ct.getNewConfig()
    ne.Shards[command.Shard]=command.GID
  case *Query:
    var num int
    if command.Num < 0 || command.Num >= len(ct.Configs) {
      num = len(ct.Configs) - 1
    } else {
      num = command.Num
    }
    return ct.getConfig(num)
  default:
    panic(fmt.Sprintf("unknown command %T%+v",command,command))
  }
  return nil
}

func getController(i raftapp.InnerExecutor) *Controller {
  return i.(*Controller)
}
