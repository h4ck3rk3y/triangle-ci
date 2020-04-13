package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/h4ck3rk3y/triangle-ci/git"
	"github.com/h4ck3rk3y/triangle-ci/workers"
)

type newCommitForm struct {
	Repository string `json:"repository_url"`
	Branch     string `json:"branch"`
}

const (
	// JobQueueSize ...
	JobQueueSize int = 10
	// WorkerLimit ...
	WorkerLimit = 4
)

var jobMap map[string]*workers.Job
var outputMap workers.OutputMap
var jobChan chan *workers.Job

func newCommitHandler(c *gin.Context) {
	var form newCommitForm

	if c.ShouldBindJSON(&form) != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"message": "invalid request"})
		c.Abort()
		return
	}

	uuid := git.CreateUUID()

	status, job := workers.EnqueJob(form.Repository, form.Branch, uuid, jobChan)
	jobMap[uuid] = job
	outputMap[uuid] = ""

	if status == workers.Queued {
		job.Status = workers.Queued
		c.JSON(http.StatusOK, gin.H{"message": "build queued", "id": uuid})
	} else {
		job.Status = workers.TryLater
		c.JSON(http.StatusOK, gin.H{"message": "queue is full try later", "id": uuid})
	}
}

func statusCheck(c *gin.Context) {
	id := c.Query("id")

	if job, ok := jobMap[id]; ok {
		c.JSON(http.StatusOK, gin.H{"status": job.Status})
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

func channelSize(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"current": len(jobChan), "capacity": 10})
}

func allRunningJobs(c *gin.Context) {
	runningJobs := make(map[string]workers.Job)
	for k, v := range jobMap {
		if v.Status == workers.Processing || v.Status == workers.Queued {
			runningJobs[k] = *v
		}
	}

	c.JSON(http.StatusOK, runningJobs)
}

func allFinishedJobs(c *gin.Context) {
	finishedJobs := make(map[string]workers.Job)
	for k, v := range jobMap {
		if v.Status != workers.Processing && v.Status != workers.Queued {
			finishedJobs[k] = *v
		}
	}

	c.JSON(http.StatusOK, finishedJobs)
}

func allJobs(c *gin.Context) {
	c.JSON(http.StatusOK, jobMap)
}

func main() {

	outputMap = make(workers.OutputMap)
	jobMap = make(map[string]*workers.Job)

	jobChan = make(chan *workers.Job, JobQueueSize)
	workers.CreateWorkerPool(WorkerLimit, jobChan, outputMap)

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/push/", newCommitHandler)
	r.GET("/status", statusCheck)
	r.GET("/output", outputCheck)
	r.GET("/stats", channelSize)
	r.GET("/finished", allFinishedJobs)
	r.GET("/pending", allRunningJobs)
	r.GET("/all", allJobs)
	r.Run()
}
