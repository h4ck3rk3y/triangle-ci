package docker

import (
	"errors"
	"os"
	"os/exec"

	"github.com/h4ck3rk3y/triangle-ci/git"
)

// RunDockerFile ...
func RunDockerFile(path string, id string) (passed bool, output string, err error) {
	os.Chdir(path)
	_, err = git.GetCIFile(path)

	if err != nil {
		return false, "", errors.New("invalid file passed")
	}

	cmd := exec.Command("docker", "build", "-t", id, ".")
	stdout, err := cmd.Output()

	if err != nil {
		return false, string(stdout), err
	}

	return true, string(stdout), nil
}
