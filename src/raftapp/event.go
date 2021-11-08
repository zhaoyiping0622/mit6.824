package raftapp

type Event interface {
  Run(*RaftServer)
}

func (rs *RaftServer) sendEvent(e Event) {
  rs.events<-e
}

func (rs *RaftServer) eventLoop() {
  for {
    select {
    case e,ok:=<-rs.events:
      if ok {
        e.Run(rs)
      }
    case <-rs.background.Done():
      return
    }
  }
}
