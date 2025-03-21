package kvsrv

// Put or Append

const GET int = 1
const PUT_APPEND int = 2
const DONE int = 3

type BaseArgs struct {
	ClientId    int64
	RequestId   int64
	RequestType int
}
type PutAppendArgs struct {
	Key   string
	Value string
	// You'll have to add definitions here.
	// Field names must start with capital letters,
	// otherwise RPC will break.
	BaseArgs *BaseArgs
}

type PutAppendReply struct {
	Value string
}

type GetArgs struct {
	Key string
	// You'll have to add definitions here.
	BaseArgs *BaseArgs
}

type GetReply struct {
	Value string
}

// 用于client向服务器确认收到response，类似于三次握手的第三次
type DoneArgs struct {
	BaseArgs *BaseArgs
}

type DoneReply struct {
}
