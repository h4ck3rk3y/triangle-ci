package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/h4ck3rk3y/triangle-ci/docker"
	"github.com/h4ck3rk3y/triangle-ci/git"
)

type newCommitForm struct {
	Repository string `json:"repository_url"`
}

// Job ...
type Job struct {
	Repository string `json:"repository_url"`
	Status     string `json:"status"`
	Branch     string `json:"branch"`
	Commit     string `json:"commit"`
	UUID       string `json:"uuid"`
	Path       string `json:"path"`
}

var statusMap map[string]string
var jobMap map[string]Job
var outputMap map[string]string
var jobChan chan Job

const (
	queued      string = "Queued"
	failed             = "Failed"
	processing         = "Processing"
	testsfailed        = "TestsFailed"
	completed          = "Completed"
	tryLater           = "Queue is full try later"
)

func newCommitHandler(c *gin.Context) {
	var form newCommitForm

	if c.ShouldBindJSON(&form) != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"message": "invalid request"})
		c.Abort()
		return
	}

	uuid := git.CreateUUID()
	outputMap[uuid] = ""

	status, job := enqueJob(form.Repository, uuid)
	jobMap[uuid] = job

	if status {
		statusMap[uuid] = queued
		c.JSON(http.StatusOK, gin.H{"message": "build queued", "id": uuid})
	} else {
		statusMap[uuid] = tryLater
		c.JSON(http.StatusOK, gin.H{"message": "queue is full try later", "id": uuid})
	}
}

func statusCheck(c *gin.Context) {
	id := c.Query("id")

	if status, ok := statusMap[id]; ok {
		c.JSON(http.StatusOK, gin.H{"status": status})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"message": "invalid id"})
		c.Abort()
	}
}

func outputCheck(c *gin.Context) {
	id := c.Query("id")

	if output, ok := outputMap[id]; ok {
		c.JSON(http.StatusOK, gin.H{"output": output})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"message": "invalid id"})
		c.Abort()
	}
}

func enqueJob(Repository string, uuid string) (bool, Job) {
	job := Job{Repository, "", "", "", uuid, ""}
	select {
	case jobChan <- job:
		return true, job
	default:
		return false, job
	}
}

func worker() {
	for job := range jobChan {
		process(job)
	}
}

func createWorkerPool(limit int) {
	for i := 0; i < limit; i++ {
		go worker()
	}
}

func process(job Job) {
	path, err := git.Clone(job.Repository, job.UUID)
	job.Path = path

	if err != nil {
		statusMap[job.UUID] = failed
	} else {
		statusMap[job.UUID] = processing
		status, output, err := docker.RunDockerFile(path, job.UUID)
		outputMap[job.UUID] = output
		if err != nil || status == false {
			statusMap[job.UUID] = testsfailed
		} else if status == true {
			statusMap[job.UUID] = completed
		}
	}

	job.Status = statusMap[job.UUID]
}

func main() {

	statusMap = make(map[string]string)
	outputMap = make(map[string]string)
	jobMap = make(map[string]Job)

	jobChan = make(chan Job, 10)
	createWorkerPool(4)

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/push/", newCommitHandler)
	r.GET("/status", statusCheck)
	r.GET("/output", outputCheck)
	r.Run()
}
