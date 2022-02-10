package raftapp

type NoticeProducer interface {
	SetValue(location NoticeLocation, seqNum int, value *AsyncRequestReply)
}

type NoticeConsumer interface {
	GetValue(location NoticeLocation, seqNum int) <-chan *AsyncRequestReply
	HasValue(location NoticeLocation, seqNum int) bool
}

type Notice interface {
	NoticeProducer
	NoticeConsumer
	CleanSnapshotable
	TermNotice
}

type WrongGroupNotice struct {}

func (n *WrongGroupNotice) SetValue(location NoticeLocation, seqNum int, value *AsyncRequestReply) {
  panic("impossible")
}

func (n *WrongGroupNotice) GetValue(location NoticeLocation, seqNum int) <-chan *AsyncRequestReply {
  ch:=make(chan *AsyncRequestReply, 1)
  ch<-&AsyncRequestReply{ Err: WrongGroup }
  close(ch)
  return ch
}

func (n *WrongGroupNotice) HasValue(location NoticeLocation, seqNum int) bool { return true }

func (n *WrongGroupNotice) UpdateTermAndLeader(term int, isLeader bool) {}

func (n *WrongGroupNotice) GenerateSnapshot() Snapshot { return nil }

func (n *WrongGroupNotice) ApplySnapshot(Snapshot) {}

func (n *WrongGroupNotice) Clean() {}

func MakeWrongGroupNotice() Notice { return &WrongGroupNotice{} }
