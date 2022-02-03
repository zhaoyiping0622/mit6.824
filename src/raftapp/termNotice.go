package raftapp

type TermNotice interface {
	UpdateTermAndLeader(term int, isLeader bool)
}
