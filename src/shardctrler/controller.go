package shardctrler

import (
	"fmt"
	"sort"

	"6.824/raftapp"
)

type Controller struct {
  Configs []Config
  Me int
  ShardInfos [NShards]ShardInfo
  Version int
  Groups map[int][]string
}

func MakeController(me int) *Controller {
  rt:=new(Controller)
  rt.Me=me
  rt.Groups=make(map[int][]string)
  rt.Configs=make([]Config, 1)
  rt.Configs[0].Groups=make(map[int][]string)
  for i:=0;i<NShards;i++ {
    rt.ShardInfos[i]=ShardInfo{
      Owner: -1,
      Version: 0,
    }
  }
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

type ControllerSnapshot struct {
  Configs []Config
  ShardInfo [NShards]ShardInfo
  Version int
}

func (ct *Controller) ApplySnapshot(i raftapp.Snapshot) {
  // if i != nil {
  //   x:=i.(*ControllerSnapshot)
  //   ct.Configs=make([]Config, len(x.Configs))
  //   for i:=range x.Configs {
  //     ct.Configs[i]=x.Configs[i]
  //   }
  //   ct.ShardInfos=x.ShardInfo
  //   ct.Version=x.Version
  // }
}

func (ct *Controller) GenerateSnapshot() raftapp.Snapshot {
  panic("impossible")
  // x:=new(ControllerSnapshot)
  // x.Configs=make([]Config, len(ct.Configs))
  // for i:=range ct.Configs {
  //   x.Configs[i]=*ct.getConfig(i)
  // }
  // x.ShardInfo=ct.ShardInfos
  // x.Version=ct.Version
  // return x
}

func(ct *Controller) Clean() {}

func (ct *Controller) getInfoResult() *InfoResult {
  groups:=make(map[int][]string)
  for k,v:=range ct.Groups {
    groups[k]=make([]string, 0)
    groups[k]=append(groups[k], v...)
  }
  return &InfoResult{
    Groups: groups,
    Version: ct.Version,
    Config: ct.getLastConfig(),
    ShardInfos: ct.ShardInfos,
  }
}

func (ct *Controller) Run(command interface{}) interface{} {
  l:=len(ct.Configs)
  version:=ct.Version
  defer func() {
    if len(ct.Configs)!=l || version != ct.Version {
      DPrintf("just run %T%+v info result %v", command, command, raftapp.PrettyPrint(ct.getInfoResult()))
    }
  }()
  switch command:=command.(type) {
  case *Join:
    ne:=ct.getNewConfig()
    for k,v:=range command.Servers {
      sort.Strings(v)
      ne.Groups[k] = v
      if vv,ok:=ct.Groups[k]; ok && EqualStringSlice(vv,v) {
      } else {
        ct.Groups[k]=v
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
  case *Info:
    return ct.getInfoResult()
  case *UpdateOwner:
    config:=ct.getLastConfig()
    shard:=command.Shard
    info:=command.Infos
    // DPrintf("get %T%+v info %v currentInfo %+v", command, command, command.Infos, *ct.getInfoResult())
      if config.Shards[shard] != info.Owner || info.Version != ct.ShardInfos[shard].Version + 1 {
        // in this config the owner of the shard is not this one
        // or
        // the version is not greater than the controller's version
        return &UpdateOwnerResult{
          Success: false,
          InfoResult: ct.getInfoResult(),
        }
      } else {
        ct.ShardInfos[shard]=*info
        ct.Version++
        return &UpdateOwnerResult{
          Success: true,
          InfoResult: ct.getInfoResult(),
        }
      }
  default:
    panic(fmt.Sprintf("unknown command %T%+v",command,command))
  }
  return nil
}

func getController(i raftapp.InnerExecutor) *Controller {
  return i.(*Controller)
}
