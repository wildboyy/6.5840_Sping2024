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
	// 两个锁，分别锁kv存储 和 请求cache，这样粒度更细，并发效率更高
	mu    sync.Mutex
	logMu sync.Mutex

	// kv 存储
	m map[string]string

	// 缓存请求结果，避免重复请求
	// key 为clientId，（由于单个client的单线程请求模型，一个新的请求到来，表明client已经收到前面的请求的回复
	// value 为请求日志，包含返回值和请求id
	requestLog map[int64]RequestLog
}

type RequestLog struct {
	requestId     int64
	responseValue string
}

func (kv *KVServer) checkDuplicatedRequest(baseArgs *BaseArgs) (dup bool, ans string) {
	clientId := baseArgs.ClientId
	requestId := baseArgs.RequestId

	kv.logMu.Lock()
	log, exists := kv.requestLog[clientId]
	kv.logMu.Unlock()

	if exists && log.requestId == requestId {
		return true, log.responseValue
	} else {
		return false, ""
	}
}

func (kv *KVServer) createRequestCache(baseArgs *BaseArgs, responseValue string) {
	kv.logMu.Lock()
	defer kv.logMu.Unlock()
	kv.requestLog[baseArgs.ClientId] = RequestLog{baseArgs.RequestId, responseValue}
}

func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	dup, ans := kv.checkDuplicatedRequest(args.BaseArgs)
	if dup {
		reply.Value = ans
		return
	}

	key := args.Key
	kv.mu.Lock()
	defer kv.mu.Unlock()
	value, ok := kv.m[key]
	if ok {
		reply.Value = value
	} else {
		reply.Value = ""
	}

	kv.createRequestCache(args.BaseArgs, reply.Value)
}

func (kv *KVServer) Put(args *PutAppendArgs, reply *PutAppendReply) {
	dup, _ := kv.checkDuplicatedRequest(args.BaseArgs)
	if dup {
		return
	}

	key := args.Key
	value := args.Value
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.m[key] = value

	kv.createRequestCache(args.BaseArgs, "")
}

func (kv *KVServer) Append(args *PutAppendArgs, reply *PutAppendReply) {
	dup, ans := kv.checkDuplicatedRequest(args.BaseArgs)
	if dup {
		reply.Value = ans
		return
	}

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
	kv.createRequestCache(args.BaseArgs, reply.Value)
}

func (kv *KVServer) Done(args *DoneArgs, reply *DoneReply) {
	clienId := args.BaseArgs.ClientId
	kv.logMu.Lock()
	defer kv.logMu.Unlock()
	delete(kv.requestLog, clienId)
}

func StartKVServer() *KVServer {
	kv := new(KVServer)

	// go 的map默认空值为nil，直接使用会出错，必须初始化
	kv.m = make(map[string]string)
	kv.requestLog = make(map[int64]RequestLog)
	return kv
}
