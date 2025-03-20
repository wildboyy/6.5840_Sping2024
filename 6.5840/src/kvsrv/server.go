package kvsrv

import (
	"log"
	"sync"
)

const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

type KVServer struct {
	mu sync.Mutex

	// Your definitions here.
	m map[string]string
}

func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	key := args.Key
	kv.mu.Lock()
	defer kv.mu.Unlock()
	value, ok := kv.m[key]
	if ok {
		reply.Value = value
	} else {
		reply.Value = ""
	}
}

func (kv *KVServer) Put(args *PutAppendArgs, reply *PutAppendReply) {
	key := args.Key
	value := args.Value
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.m[key] = value
}

func (kv *KVServer) Append(args *PutAppendArgs, reply *PutAppendReply) {
	key := args.Key
	value := args.Value
	kv.mu.Lock()
	defer kv.mu.Unlock()
	oldValue, ok := kv.m[key]
	if ok {
		kv.m[key] = oldValue + value
		reply.Value = oldValue
	} else {
		kv.m[key] = value
		reply.Value = ""
	}
}

func StartKVServer() *KVServer {
	kv := new(KVServer)
	kv.m = make(map[string]string) // map必须初始化

	return kv
}
