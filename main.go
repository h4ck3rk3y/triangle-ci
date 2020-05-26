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
	Host       bool   `json:"host" form:"host,default=false"`
}

const (
	// JobQueueSize ...
	JobQueueSize int = 10
	// WorkerLimit ...
	WorkerLimit = 4
)

var jobMap map[string]*workers.Job
var jobChan chan *workers.Job

func newCommitHandler(c *gin.Context) {
	var form newCommitForm

	if c.ShouldBindJSON(&form) != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"message": "invalid request"})
		c.Abort()
		return
	}

	uuid := git.CreateUUID()

	status, job := workers.EnqueJob(form.Repository, form.Branch, uuid, form.Host, jobChan)
	jobMap[uuid] = job

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

	if job, ok := jobMap[id]; ok {
		c.JSON(http.StatusOK, gin.H{"output": job.Output})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"message": "invalid id"})
		c.Abort()
	}
}

func channelSize(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"current": len(jobChan), "capacity": 10})
}

func allRunningJobs(c *gin.Context) {
	runningJobs := make(map[string]*workers.Job)
	for k, v := range jobMap {
		if v.Status == workers.Processing || v.Status == workers.Queued {
			runningJobs[k] = v
		}
	}
	v := make([]*workers.Job, 0, len(runningJobs))
	for _, value := range runningJobs {
		v = append(v, value)
	}
	c.JSON(http.StatusOK, v)
}

func allFinishedJobs(c *gin.Context) {
	finishedJobs := make(map[string]*(workers.Job))
	for k, v := range jobMap {
		if v.Status != workers.Processing && v.Status != workers.Queued {
			finishedJobs[k] = v
		}
	}
	v := make([]*workers.Job, 0, len(finishedJobs))
	for _, value := range finishedJobs {
		v = append(v, value)
	}
	c.JSON(http.StatusOK, v)
}

func allJobs(c *gin.Context) {
	v := make([]*workers.Job, 0, len(jobMap))
	for _, value := range jobMap {
		v = append(v, value)
	}
	c.JSON(http.StatusOK, v)
}

func main() {

	jobMap = make(map[string]*workers.Job)

	jobChan = make(chan *workers.Job, JobQueueSize)
	workers.CreateWorkerPool(WorkerLimit, jobChan)

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
	r.LoadHTMLFiles("static/*")
	r.StaticFile("/ui/all", "static/status.html")
	r.StaticFile("/ui/output", "static/output.html")
	r.Run()
}
