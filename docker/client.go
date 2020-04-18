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
	cmd.Wait()
	return true, nil
}
