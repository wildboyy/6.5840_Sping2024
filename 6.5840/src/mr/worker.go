package mr

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

type Worker struct {
	uuid    string
	mapf    func(string, string) []KeyValue
	reducef func(string, []string) string
	kill    chan bool
}

// main/mrworker.go calls this function.
// The workers will talk to the coordinator via RPC.
// Each worker process will, in a loop,
// ask the coordinator for a task, read the task's input from one or more files,
// execute the task, write the task's output to one or more files,
// and again ask the coordinator for a new task.
func StartWorker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your implementation here.

	w := createWorker(mapf, reducef)

	// 心跳
	// go w.ping()

	for {
		ok, job := w.askForJob()
		if ok {
			w.doJob(job)
		} else {
			break
		}
		time.Sleep(time.Second)
	}

	w.shutdown()
}

func createWorker(mapf func(string, string) []KeyValue, reducef func(string, []string) string) Worker {
	w := Worker{}

	id, _ := generateUUID()

	w.uuid = id
	w.mapf = mapf
	w.reducef = reducef
	w.kill = make(chan bool)

	return w

}

func (w *Worker) shutdown() {
	w.kill <- true
}

func (w *Worker) doJob(job Job) {
	if job.Type == JOB_MAP {
		//fmt.Println("do map")
		w.doMap(job)
	} else if job.Type == JOB_REDUCE {
		//fmt.Println("do reduce")
		w.doReduce(job)
	} else {
		fmt.Errorf("getting an empty job!")
	}
}

func (w *Worker) doMap(job Job) {
	filename := job.InputFileName
	intermediate := []KeyValue{}

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("cannot open %v", filename)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", filename)
	}
	file.Close()
	kva := w.mapf(filename, string(content))
	intermediate = append(intermediate, kva...)

	// mr-X-Y, X: job number, Y: reduce number
	for i := 0; i < job.NReduce; i++ {
		oname := "mr-" + strconv.Itoa(job.JobNum) + "-" + strconv.Itoa(i)
		ofile, _ := os.Create(oname)

		for idx := range intermediate {
			kv := &intermediate[idx]
			if ihash(kv.Key)%job.NReduce == i {
				fmt.Fprintf(ofile, "%v %v\n", kv.Key, kv.Value)
			}
		}
		ofile.Close()
	}

	args := JobDoneArgs{}
	args.JobNum = job.JobNum
	args.Type = JOB_MAP
	args.WorkerID = w.uuid

	reply := JobDoneReply{}

	ok := call("Coordinator.JobDone", &args, &reply)

	if !ok {
		fmt.Errorf("JobDoneRpc failed")
	}

}

func (w *Worker) doReduce(job Job) {
	Y := job.JobNum
	inputFilePattern := "mr-[0-9]*-" + strconv.Itoa(Y)
	oname := "mr-out-" + strconv.Itoa(Y)
	ofile, _ := os.Create(oname)

	//fmt.Println("打开文件")
	files, err := filepath.Glob(inputFilePattern)
	//fmt.Println("打开文结束")

	if err != nil {
		fmt.Printf("匹配文件时出现错误: %v\n", err)
		return
	}

	// 遍历匹配到的文件
	intermediate := []KeyValue{}
	for idx := range files {
		file := &files[idx]
		f, err := os.Open(*file)

		if err != nil {
			fmt.Printf("打开文件 %s 时出错: %v\n", *file, err)
			continue
		}

		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			words := strings.Fields(line)
			intermediate = append(intermediate, KeyValue{words[0], words[1]})
		}

		// 检查扫描过程中是否有错误
		if err := scanner.Err(); err != nil {
			fmt.Printf("读取文件 %s 时出错: %v\n", *file, err)
		}
	}

	sort.Sort(ByKey(intermediate))

	i := 0
	for i < len(intermediate) {
		j := i + 1
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}
		output := w.reducef(intermediate[i].Key, values)

		// this is the correct format for each line of Reduce output.
		fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)

		i = j
	}

	ofile.Close()

	args := JobDoneArgs{}
	args.JobNum = job.JobNum
	args.Type = JOB_REDUCE
	args.WorkerID = w.uuid

	reply := JobDoneReply{}

	ok := call("Coordinator.JobDone", &args, &reply)

	if !ok {
		fmt.Errorf("JobDoneRpc failed")
	}
}

func (w *Worker) askForJob() (bool, Job) {
	args := AskJobArgs{}
	args.WorkerID = w.uuid

	reply := AskJobReply{}

	ok := call("Coordinator.AskForJob", &args, &reply)

	return ok, reply.Job
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	//fmt.Println(err)
	return false
}

func generateUUID() (string, error) {
	// 生成 16 字节的随机数据
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	// 将随机数据转换为十六进制字符串
	return hex.EncodeToString(bytes), nil
}

// func (w *Worker) ping() {
// 	for {
// 		select {
// 		case <-w.kill:
// 			fmt.Println("worker shutdonw")
// 			return
// 		default:
// 			pingMaster(w.uuid)
// 		}
// 	}
// }

// func pingMaster(uuid string) {
// 	args := PingArgs{}
// 	args.Worker_ID = uuid

// 	reply := PingReply{}

// 	ok := call("Coordinator.Ping", &args, &reply)
// 	if ok {

// 	} else {
// 		//fmt.Printf("ping failed")
// 	}
// }

// func jobFinished() bool {
// 	panic("unimplemented")
// }
