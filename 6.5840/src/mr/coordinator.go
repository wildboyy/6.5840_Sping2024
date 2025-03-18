package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

// For this lab, have the coordinator wait for ten seconds;
// after that the coordinator should assume the worker has died (of course, it might not have).
// 我这里不是主动检查任务是否超时，而是在收到worker的任务请求时，去检查是否有超时的任务可以分配。
// 毕竟，没有worker来领任务的话，将任务标记为超时是无意义的。
const JOB_MAX_EXECUTION_TIME = 10 * time.Second

// The coordinator should notice if a worker hasn't completed
// its task in a reasonable amount of time (for this lab, use ten seconds),
// and give the same task to a different worker.
// master 有两个线程，主线程，和rpc线程
type Coordinator struct {
	Map_jobs    []Job
	Reduce_jobs []Job
	NReduce     int
	Files       []string
	MR_Finished bool

	kill chan bool
	mu   sync.Mutex
}

// Your code here -- RPC handlers for the worker to call.

func (c *Coordinator) AskForJob(args *AskJobArgs, reply *AskJobReply) error {

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	workerID := args.WorkerID
	if !c.MapDone() {
		for idx := range c.Map_jobs {
			job := &c.Map_jobs[idx]
			if job.Status == IDLE {
				job.Status = PROCESSING
				job.WorkerID = workerID
				job.StartTime = now
				reply.Job = *job
				fmt.Printf("job %d send to %s \n", idx, workerID)
				break
			} else if job.Status == PROCESSING {

				if now.Sub(job.StartTime) > JOB_MAX_EXECUTION_TIME {
					// 超时了
					fmt.Printf("job %d send to %s \n", idx, workerID)
					job.WorkerID = workerID
					job.StartTime = now
					reply.Job = *job
					break
				}
			}
		}

	} else if !c.ReduceDone() {
		for idx := range c.Reduce_jobs {
			job := &c.Reduce_jobs[idx]
			if job.Status == IDLE {
				job.Status = PROCESSING
				job.WorkerID = workerID
				job.StartTime = now
				reply.Job = *job
				break
			} else if job.Status == PROCESSING {

				if now.Sub(job.StartTime) > JOB_MAX_EXECUTION_TIME {
					// 超时了
					job.WorkerID = workerID
					job.StartTime = now
					reply.Job = *job
					break
				}
			}
		}
	}

	return nil
}

func (c *Coordinator) JobDone(args *JobDoneArgs, reply *JobDoneReply) error {
	jobType := args.Type
	jobNum := args.JobNum
	//workerId := args.WorkerID

	if jobType == JOB_MAP {
		c.Map_jobs[jobNum].Status = FINISHED
	} else {
		c.Reduce_jobs[jobNum].Status = FINISHED
	}

	return nil
}

func (c *Coordinator) Ping(args *PingArgs, reply *PingReply) error {
	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// 定期检查worker是否掉线，是否任务超时
func (c *Coordinator) scheduler() {

}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	return c.MR_Finished
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
	c.NReduce = nReduce
	c.Files = files
	c.MR_Finished = false
	c.createJobs()
	c.server()
	c.scheduler()
	return &c
}

func (c *Coordinator) createJobs() {
	for idx, file := range c.Files {
		job := Job{}
		job.Status = IDLE
		job.JobNum = idx
		job.Type = JOB_MAP
		job.InputFileName = file
		job.NReduce = c.NReduce
		c.Map_jobs = append(c.Map_jobs, job)
	}

	for i := 0; i < c.NReduce; i++ {
		job := Job{}
		job.Status = IDLE
		job.JobNum = i
		job.Type = JOB_REDUCE
		job.NReduce = c.NReduce
		c.Reduce_jobs = append(c.Reduce_jobs, job)
	}
}

func (c *Coordinator) MapDone() bool {
	if c.MR_Finished {
		return true
	}
	for idx := range c.Map_jobs {
		job := &c.Map_jobs[idx]
		if job.Status != FINISHED {
			return false
		}
	}
	return true
}

func (c *Coordinator) ReduceDone() bool {
	if c.MR_Finished {
		return true
	}
	for idx := range c.Reduce_jobs {
		job := &c.Reduce_jobs[idx]
		if job.Status != FINISHED {
			return false
		}
	}

	c.MR_Finished = true
	return true
}
