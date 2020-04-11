package git

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"

	"gopkg.in/src-d/go-git.v4/plumbing"

	"gopkg.in/src-d/go-git.v4"
)

// Clone ...
func Clone(url string, branch string, uuid string) (path string, err error) {

	path = CreatePath(uuid)

	_, err = git.PlainClone(path, false, &git.CloneOptions{
		URL:           url,
		Progress:      os.Stdout,
		ReferenceName: plumbing.ReferenceName(branch),
	})

	if err != nil {
		return "failure while cloning", err
	}

	fmt.Println("Cloning finished succesfully")

	return path, nil
}

// GetCIFile ...
func GetCIFile(path string) (dockerfile string, err error) {

	path = path + "/Dockerfile"

	_, err = os.Open(path)

	if err != nil {
		log.Fatal("couldn't open file")
		return "", err
	}

	return path, err
}

// CreateUUID ...
func CreateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return uuid
}

// CreatePath ...
func CreatePath(uuid string) (path string) {
	return "/tmp/" + uuid
}
