package workers

import (
	"github.com/h4ck3rk3y/triangle-ci/docker"
	"github.com/h4ck3rk3y/triangle-ci/git"
)

// Job ...
type Job struct {
	Repository string `json:"repository_url"`
	Status     string `json:"status"`
	Branch     string `json:"branch"`
	Commit     string `json:"commit"`
	UUID       string `json:"uuid"`
	Path       string `json:"path"`
}

const (
	// Queued ...
	Queued string = "Queued"
	// Failed ...
	Failed = "Failed"
	// Processing ...
	Processing = "Processing"
	// TestsFailed ...
	TestsFailed = "TestsFailed"
	// Completed ...
	Completed = "Completed"
	// TryLater ...
	TryLater = "Queue is full try later"
)

// StatusMap ...
type StatusMap map[string]string
type jobMap map[string]Job

// OutputMap ...
type OutputMap map[string]string

// CreateWorkerPool ...
func CreateWorkerPool(limit int, jobChan chan Job, statusMap StatusMap, outputMap OutputMap) {
	for i := 0; i < limit; i++ {
		go worker(jobChan, statusMap, outputMap)
	}
}

// EnqueJob ...
func EnqueJob(Repository string, uuid string, jobChan chan Job) (string, Job) {
	job := Job{Repository, "", "", "", uuid, ""}
	select {
	case jobChan <- job:
		return Queued, job
	default:
		return TryLater, job
	}
}

func worker(jobChan chan Job, statusMap StatusMap, outputMap OutputMap) {
	for job := range jobChan {
		process(job, statusMap, outputMap)
	}
}

func process(job Job, statusMap StatusMap, outputMap OutputMap) {
	path, err := git.Clone(job.Repository, job.UUID)
	job.Path = path

	if err != nil {
		statusMap[job.UUID] = Failed
	} else {
		statusMap[job.UUID] = Processing
		status, output, err := docker.RunDockerFile(path, job.UUID)
		outputMap[job.UUID] = output
		if err != nil || status == false {
			statusMap[job.UUID] = TestsFailed
		} else if status == true {
			statusMap[job.UUID] = Completed
		}
	}

	job.Status = statusMap[job.UUID]
}
