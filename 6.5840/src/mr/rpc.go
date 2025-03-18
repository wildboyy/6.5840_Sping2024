package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

// example to show how to declare the arguments
// and reply for an RPC.
type AskJobArgs struct {
	WorkerID string
}

type AskJobReply struct {
	Job Job
}

type PingArgs struct {
	Worker_ID string
}

type PingReply struct {
}

type JobDoneArgs struct {
	Type     int
	JobNum   int
	WorkerID string
}

type JobDoneReply struct {
}

// Add your RPC definitions here.

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
