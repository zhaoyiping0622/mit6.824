package shardctrler

import "6.824/raftapp"

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

func (ct *Controller) CreateSnapshot() interface{} {
  panic("controller do not support snapshot")
}

func getController(i raftapp.RaftApp) *Controller {
  return i.(*Controller)
}
