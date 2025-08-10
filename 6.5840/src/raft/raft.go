package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"math"

	//	"bytes"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	//	"6.5840/labgob"
	"6.5840/labrpc"
)

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 3D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 3D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

type Log struct {
	Command interface{}
	Term    int
}

const LEADER int32 = 1
const CANDIDATE int32 = 2
const FOLLOWER int32 = 3

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (3A, 3B, 3C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	// Persistent state on all servers
	currentTerm int
	votedFor    int
	log         []Log // 下标从 1 开始
	role        int32 // LEADER or CANDIDATE or FOLLOWER

	// Volatile state on all servers
	commitIndex  int // 最新已提交的 log 的 index，初始为 0
	lastApplied  int
	leaderExists bool // for ticker to check whether it should start an election

	// Volatile state on leaders
	nextIndex  []int
	matchIndex []int

	appendLocks []sync.Mutex

	applyCh chan ApplyMsg
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (3A).
	term = rf.currentTerm
	isleader = (rf.role == LEADER)

	return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
// before you've implemented snapshots, you should pass nil as the
// second argument to persister.Save().
// after you've implemented snapshots, pass the current snapshot
// (or nil if there's not yet a snapshot).
func (rf *Raft) persist() {
	// Your code here (3C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// raftstate := w.Bytes()
	// rf.persister.Save(raftstate, nil)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (3C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (3D).

}

// example RequestVote RPC arguments structure.
// field names must start with capital letters!
type RequestVoteArgs struct {
	// Your data here (3A, 3B).
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	// Your data here (3A).
	Term        int
	VoteGranted bool
}

type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []Log
	LeaderCommit int
}
type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	leaderTerm := args.Term

	rf.mu.Lock()
	defer rf.mu.Unlock()

	// heartbeat && leader term check
	if leaderTerm >= rf.currentTerm {
		reply.Success = true
		reply.Term = leaderTerm
		rf.role = FOLLOWER
		rf.currentTerm = leaderTerm
		rf.leaderExists = true
	} else {
		reply.Success = false
		reply.Term = rf.currentTerm
		return
	}

	// log replication
	if len(args.Entries) != 0 {

		// prevLog不匹配，需要删除
		if args.PrevLogIndex >= len(rf.log) {
			reply.Success = false
			return
		}

		// 新日志已经在对应 index 处存在
		if args.PrevLogIndex+1 < len(rf.log) && rf.log[args.PrevLogIndex+1].Term == args.Entries[0].Term {
			reply.Success = true
			return
		}

		prevLog := rf.log[args.PrevLogIndex]
		if prevLog.Term != args.PrevLogTerm {
			reply.Success = false
		} else {
			// follwer 首次收到 lognetry 只写入 不提交，之后leader通过心跳来告知 follower 提交
			rf.log = append(rf.log[:args.PrevLogIndex+1], args.Entries...)
			reply.Success = true
			// fmt.Printf("follower:%v replicate log at index %v term %v \n", rf.me, len(rf.log)-1, args.Entries[0].Term)
		}
	} else {
		// 2nd Phase commit

		if args.LeaderCommit < len(rf.log) && rf.log[args.LeaderCommit].Term == leaderTerm && rf.commitIndex != args.LeaderCommit {
			// fmt.Printf("follower:%v commit at index %v \n", rf.me, args.LeaderCommit)

			for i := rf.commitIndex + 1; i <= args.LeaderCommit; i++ {
				rf.applyCh <- ApplyMsg{
					CommandValid: true,
					Command:      rf.log[i].Command,
					CommandIndex: i,
				}
			}
			rf.commitIndex = args.LeaderCommit

		}
	}

}

// 将日志复制到 follower
func (rf *Raft) logReplication(index int, term int) {

	n := len(rf.peers)
	me := rf.me

	//fmt.Printf("leader %v，current term %v, append new log, index%v \n", rf.me, rf.currentTerm, index)
	replicateChan := make(chan bool, len(rf.peers))
	for i := 0; i < n; i++ {
		if i != me {
			go rf.tryReplicate(i, replicateChan, term)
		}
	}

	// 半数复制完成，则提交
	cnt := 0
	for i := 0; i < n-1; i++ {
		success := <-replicateChan
		if success {
			cnt++
		}
		if cnt >= n/2 {
			// 强制按顺序提交
			for rf.commitIndex != index-1 {
			}

			// 提交
			rf.mu.Lock()
			rf.commitIndex = index
			rf.applyCh <- ApplyMsg{
				CommandValid: true,
				Command:      rf.log[index].Command,
				CommandIndex: rf.commitIndex,
			}
			rf.mu.Unlock()

			// fmt.Printf("leader:%v commit at index %v term %v\n", rf.me, rf.commitIndex, rf.currentTerm)
			break
		}
	}

}

func (rf *Raft) tryReplicate(i int, replicateChan chan bool, leaderTerm int) {

	rf.appendLocks[i].Lock()
	defer rf.appendLocks[i].Unlock()
	for rf.nextIndex[i] < len(rf.log) && rf.role == LEADER && leaderTerm == rf.currentTerm {

		appendIndex := 1
		if rf.nextIndex[i] < len(rf.log) {
			appendIndex = rf.nextIndex[i]
		} else {
			break
		}

		args := AppendEntriesArgs{
			Term:         leaderTerm,
			LeaderId:     rf.me,
			PrevLogIndex: appendIndex - 1,
			PrevLogTerm:  rf.log[appendIndex-1].Term,
			Entries:      []Log{rf.log[appendIndex]},
			LeaderCommit: rf.commitIndex,
		}
		reply := AppendEntriesReply{}

		ok := rf.sendAppendEntries(i, &args, &reply)

		if ok && reply.Success {
			// 如果用 rf.nextIndex[i]++ ，则不幂等，如果先后复制 log3，log2，则 log2会复制两次，rf.nextIndex[i]++也会执行两次
			rf.nextIndex[i] = args.PrevLogIndex + 2
		} else if ok {
			if reply.Term > rf.currentTerm {
				rf.role = FOLLOWER
				rf.currentTerm = reply.Term
				replicateChan <- false
				return
			} else {
				rf.nextIndex[i] = int(math.Max(1, float64(args.PrevLogIndex)))
			}
		}
	}

	if rf.role == LEADER {
		replicateChan <- true
	} else {
		replicateChan <- false
	}
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) heartBeat() {
	for !rf.killed() {
		if rf.role != LEADER {
			continue
		}

		size := len(rf.peers)
		args := AppendEntriesArgs{}
		args.Entries = make([]Log, 0)
		args.Term = rf.currentTerm
		args.LeaderCommit = rf.commitIndex

		me := rf.me

		for i := 0; i < size; i++ {
			if i != me {
				//fmt.Printf("heartBeat:%v \n", i)
				reply := AppendEntriesReply{}

				go func(i int) {
					ok := rf.sendAppendEntries(i, &args, &reply)

					// 成功则已，失败则下台，并停止发送心跳
					if ok && !reply.Success && reply.Term > rf.currentTerm {
						rf.currentTerm = reply.Term
						rf.role = FOLLOWER
					}
				}(i)
			}
		}

		// 心跳间隔 100ms(lab3A 的提示中说不要超过 10次每秒）
		time.Sleep(100 * time.Millisecond)
	}
}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (3A, 3B).
	rf.mu.Lock()
	defer rf.mu.Unlock()

	candidateTerm := args.Term
	candidateLastLogIndex := args.LastLogIndex
	candidateLastLogTerm := args.LastLogTerm
	voterLastLog := rf.log[len(rf.log)-1]

	// 无论如何，发现更新的 term，就要更新
	if candidateTerm <= rf.currentTerm {
		reply.VoteGranted = false
	} else if candidateLastLogTerm < voterLastLog.Term {
		rf.currentTerm = candidateTerm
		reply.VoteGranted = false
	} else if candidateLastLogTerm == voterLastLog.Term && candidateLastLogIndex < len(rf.log)-1 {
		rf.currentTerm = candidateTerm
		reply.VoteGranted = false
	} else {
		rf.currentTerm = candidateTerm
		rf.role = FOLLOWER
		rf.votedFor = args.CandidateId
		reply.VoteGranted = true
		rf.leaderExists = true
	}

	reply.Term = rf.currentTerm

}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply, voteChan chan *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	voteChan <- reply
	return ok
}

// 向所有节点发送竞选消息，并等待结果
func (rf *Raft) startElection() {
	size := len(rf.peers)

	rf.mu.Lock()

	rf.currentTerm++

	rf.role = CANDIDATE
	rf.votedFor = rf.me
	args := RequestVoteArgs{}
	args.Term = rf.currentTerm
	args.LastLogIndex = len(rf.log) - 1
	args.LastLogTerm = rf.log[args.LastLogIndex].Term

	rf.mu.Unlock()

	// 并发发送竞选请求，以防 坏节点/网络问题 阻塞选举进程
	voteChan := make(chan *RequestVoteReply, len(rf.peers))
	for i := 0; i < size; i++ {
		if i != rf.me {
			reply := RequestVoteReply{}
			go rf.sendRequestVote(i, &args, &reply, voteChan)
		}
	}

	// 统计投票信息
	success := false
	for voteCnt, falseCnt := 1, 0; ; {
		v := <-voteChan
		if v.VoteGranted {
			voteCnt++
		} else {
			falseCnt++
		}

		if falseCnt > size/2 {
			break
		}

		if voteCnt > size/2 {
			success = true
			break
		}
	}
	// 在选举过程中，可能会收到其他 leader 的请求，转为 FOLLOWER
	// 此时，无视选举结果，直接返回
	if rf.role == FOLLOWER {
		return
	}

	if success {
		rf.role = LEADER
		// 成为 leader 后，更新 nextIndex
		for i := 0; i < len(rf.peers); i++ {
			rf.nextIndex[i] = rf.commitIndex + 1
		}
		//fmt.Printf("%v become leader \n", rf.me)
	} else {
		rf.role = FOLLOWER
	}
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first  return value is the index that the command will appear at if it's ever committed.
// the second return value is the current term.
// the third  return value is true if this server believes it is the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := false

	rf.mu.Lock()
	defer rf.mu.Unlock()

	// Your code here (3B).
	if rf.role == LEADER {
		isLeader = true
		newEntry := Log{command, rf.currentTerm}
		index = len(rf.log)
		rf.log = append(rf.log, newEntry)
		term = rf.currentTerm

		// fmt.Printf("leader %v, start replicate index%v, term%v \n", rf.me, index, term)

		// replication start
		// term表示发出同步请求时 leader 的 term，以便其他节点能判断这是否是个过期的同步请求
		go rf.logReplication(len(rf.log)-1, term)

	}

	return index, term, isLeader
}

// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) ticker() {
	for rf.killed() == false {
		// Your code here (3A)
		// Check if a leader election should be started.
		if rf.role == FOLLOWER && !rf.leaderExists {
			rf.leaderExists = false
			rf.startElection()
		} else {
			// 每次 tick 都要刷新节点的 leaderExist，用于判断两次 tick 之间是否收到过 leader 的消息
			rf.leaderExists = false
		}

		// pause for a random amount of time between 50 and 350
		// milliseconds.
		ms := 150 + (rand.Int63() % 450)
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (3A, 3B, 3C).
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.log = make([]Log, 1) // raft 的 log index 从 1 开始，所以初始化的时候已经包含一个空值，方便之后下标访问。
	rf.log[0].Term = 0
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.applyCh = applyCh
	rf.role = FOLLOWER
	rf.leaderExists = false
	rf.nextIndex = make([]int, len(rf.peers))
	for i := 0; i < len(rf.peers); i++ {
		rf.nextIndex[i] = 1
	}
	rf.appendLocks = make([]sync.Mutex, len(rf.peers))

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.ticker()

	// start heartBeat
	go rf.heartBeat()

	return rf
}
