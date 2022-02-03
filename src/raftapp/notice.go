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
	Snapshotable
	TermNotice
}
