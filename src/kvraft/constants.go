package kvraft

import "time"

const LeaderCheckTime time.Duration = 10 * time.Millisecond
const EventLoopLength int = 1000
const ServerRpcTimeout time.Duration = 1 * time.Second
const TriggerRemoveTryTimes int = 1
const MaxSessionNumber int = 100
const ClientRequestChanSize int = 200
