package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type newCommitForm struct {
	Repository string `json:"repository_url"`
}

var statusMap map[string]string

func newCommitHandler(c *gin.Context) {
	var form newCommitForm

	if c.ShouldBindJSON(&form) != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"message": "invalid request"})
		c.Abort()
		return
	}

	statusMap["foo"] = "processing"

	c.JSON(http.StatusOK, gin.H{"message": "your build is under progress", "id": "foo"})
}

func statusCheck(c *gin.Context) {
	statusID := c.Query("id")

	if status, ok := statusMap[statusID]; ok {
		c.JSON(http.StatusOK, gin.H{"status": status})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"message": "invalid status id"})
		c.Abort()
	}
}

func main() {

	statusMap = make(map[string]string)
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/push/", newCommitHandler)
	r.GET("/status", statusCheck)
	r.Run()
}
