package git

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/src-d/go-git.v4"
)

// Clone ...
func Clone(url string) (path string, err error) {

	_, err = git.PlainClone("/tmp/foo", false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})

	if err != nil {
		fmt.Println("Something failed while cloning")
		return "", err
	}

	fmt.Println("Cloning finished succesfully")

	return "/tmp/foo", nil
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
