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

// Status ...
type Status struct {
	Path   string `json:"path"`
	Status string `json:"status"`
	ID     string `json:"id"`
}

var statusMap map[string]Status
var urlIDMap map[string]string
var outputMap map[string]string

const (
	failed      string = "Failed"
	processing         = "Processing"
	testsfailed        = "TestsFailed"
	completed          = "Completed"
)

func newCommitHandler(c *gin.Context) {
	var form newCommitForm

	if c.ShouldBindJSON(&form) != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"message": "invalid request"})
		c.Abort()
		return
	}

	path, uuid, err := git.Clone(form.Repository)

	if err != nil {
		statusMap[uuid] = Status{path, failed, uuid}
	} else {
		statusMap[uuid] = Status{path, processing, uuid}
		// @ToDo put this on a separate worker
		status, output, err := docker.RunDockerFile(path, uuid)
		outputMap[uuid] = output
		if err != nil || status == false {
			statusMap[uuid] = Status{path, testsfailed, uuid}
		} else if status == true {
			statusMap[uuid] = Status{path, completed, uuid}
		}
	}

	urlIDMap[form.Repository] = uuid

	c.JSON(http.StatusOK, gin.H{"message": "your build is under progress", "id": uuid})
}

func statusCheck(c *gin.Context) {
	id := c.Query("id")

	if status, ok := statusMap[id]; ok {
		c.JSON(http.StatusOK, gin.H{"status": status.Status})
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

func main() {

	statusMap = make(map[string]Status)
	urlIDMap = make(map[string]string)
	outputMap = make(map[string]string)
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
