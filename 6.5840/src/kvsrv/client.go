package kvsrv

import (
	"6.5840/labrpc"
)
import "crypto/rand"
import "math/big"

type Clerk struct {
	server *labrpc.ClientEnd
	// You will have to modify this struct.
	id int64
}

// 生成一个大数作为请求id，server记录请求id以及对应执行情况，来避免重复执行
func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(server *labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.server = server

	ck.id = nrand()
	return ck
}

// fetch the current value for a key.
// returns "" if the key does not exist.
// keeps trying forever in the face of all other errors.
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
func (ck *Clerk) Get(key string) string {
	server := ck.server
	args := GetArgs{key, &BaseArgs{ck.id, nrand(), GET}}
	reply := GetReply{""}

	// 循环到成功为止
	for !server.Call("KVServer.Get", &args, &reply) {
	}

	// 确认收到response
	args.BaseArgs.RequestType = DONE
	doneArgs := DoneArgs{args.BaseArgs}
	for !server.Call("KVServer.Done", &doneArgs, &DoneReply{}) {
	}

	return reply.Value
}

// shared by Put and Append.
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
func (ck *Clerk) PutAppend(key string, value string, op string) string {
	server := ck.server
	args := PutAppendArgs{key, value, &BaseArgs{ck.id, nrand(), PUT_APPEND}}
	reply := PutAppendReply{""}

	// 循环到成功为止
	for !server.Call("KVServer."+op, &args, &reply) {
	}

	// 确认收到response
	args.BaseArgs.RequestType = DONE
	doneArgs := DoneArgs{args.BaseArgs}
	for !server.Call("KVServer.Done", &doneArgs, &DoneReply{}) {
	}

	return reply.Value
}

func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, "Put")
}

// Append value to key's value and return that value
func (ck *Clerk) Append(key string, value string) string {
	return ck.PutAppend(key, value, "Append")
}
