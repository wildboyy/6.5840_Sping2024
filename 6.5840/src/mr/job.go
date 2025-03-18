package mr

import "time"

// job status
const IDLE int = 0
const PROCESSING int = 1
const FINISHED int = 2

// job type
const JOB_NONE int = 0
const JOB_MAP int = 1
const JOB_REDUCE int = 2

type Job struct {
	Status         int
	Type           int
	JobNum         int
	InputPath      string
	InputFileName  string
	OutputPath     string
	OutputFileName string
	WorkerID       string
	NReduce        int
	StartTime      time.Time
}
