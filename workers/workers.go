package workers

import (
	"github.com/h4ck3rk3y/triangle-ci/docker"
	"github.com/h4ck3rk3y/triangle-ci/git"
)

// Job ...
type Job struct {
	Repository  string `json:"repository_url"`
	Status      string `json:"status"`
	Branch      string `json:"branch"`
	Commit      string `json:"commit"`
	UUID        string `json:"uuid"`
	Path        string `json:"path"`
	Output      string `json:"output"`
	CloneOutput string `json:"clone_output"`
}

const (
	// Queued ...
	Queued string = "Queued"
	// Cloned ...
	Cloned string = "Cloned"
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

// CreateWorkerPool ...
func CreateWorkerPool(limit int, jobChan chan *Job) {
	for i := 0; i < limit; i++ {
		go worker(jobChan)
	}
}

// EnqueJob ...
func EnqueJob(repository string, branch string, uuid string, jobChan chan *Job) (string, *Job) {
	job := Job{repository, "", branch, "", uuid, "", "", ""}
	select {
	case jobChan <- &job:
		return Queued, &job
	default:
		return TryLater, &job
	}
}

func worker(jobChan chan *Job) {
	for job := range jobChan {
		process(job)
	}
}

func process(job *Job) {
	path, output, err := git.Clone(job.Repository, job.Branch, job.UUID)
	job.Path = path
	job.Status = Cloned
	job.CloneOutput = output

	if err != nil {
		job.Status = Failed
	} else {
		job.Status = Processing
		status, output, err := docker.RunDockerFile(path, job.UUID)
		job.Output = output
		if err != nil || status == false {
			job.Status = TestsFailed
		} else if status == true {
			job.Status = Completed
		}
	}
}
