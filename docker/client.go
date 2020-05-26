package docker

import (
	"bufio"
	"errors"
	"os/exec"

	"github.com/h4ck3rk3y/triangle-ci/git"
)

// RunDockerFile ...
func RunDockerFile(path string, id string, output *string) (status bool, err error) {
	_, err = git.GetCIFile(path)

	if err != nil {
		return false, errors.New("invalid file passed")
	}

	cmd := exec.Command("docker", "build", "-t", id, ".")
	cmd.Dir = path
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	var foo string
	for scanner.Scan() {
		foo = foo + "\n" + scanner.Text()
		*output = foo
	}
	err = cmd.Wait()
	if err != nil {
		return false, err
	}
	return true, nil
}

// ComposeUp ...
func ComposeUp(path string, id string, output *string) (status bool, err error) {
	_, err = git.GetDockerComposeFile((path))

	if err != nil {
		return false, errors.New("cant find docker compose")
	}

	cmd := exec.Command("docker-compose", "up")
	cmd.Dir = path
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	var foo string
	for scanner.Scan() {
		foo = foo + "\n" + scanner.Text()
		*output = foo
	}
	err = cmd.Wait()
	if err != nil {
		return false, err
	}
	return true, nil
}
