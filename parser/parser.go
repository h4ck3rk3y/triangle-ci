package parser

import (
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/ghodss/yaml"
)

// DockerFile ...
type DockerFile struct {
	DockerImage string       `json:"docker-image"`
	Steps       []StepStruct `json:"steps"`
}

// StepStruct ...
type StepStruct struct {
	RunCommand string `json:"run"`
}

// ConvertYAMLToDockerfile ...
func ConvertYAMLToDockerfile(path string, repo string) (err error) {
	ciPath := path + "/triangle.yml"

	readFile, err := ioutil.ReadFile(ciPath)

	if err != nil {
		return err
	}

	var dockerYml DockerFile
	err = yaml.Unmarshal(readFile, &dockerYml)
	if err != nil {
		return err
	}

	dockerFile := fmt.Sprintf("FROM %s\n", dockerYml.DockerImage)

	parsedRepo, err := url.Parse(repo)

	if err != nil {
		return err
	}

	workpath := fmt.Sprintf("/go/src/%s%s", parsedRepo.Host, parsedRepo.Path)

	dockerFile = dockerFile + fmt.Sprintf("ADD . %s\n", workpath)
	dockerFile = dockerFile + fmt.Sprintf("WORKDIR %s\n", workpath)

	for _, step := range dockerYml.Steps {
		dockerFile = dockerFile + fmt.Sprintf("RUN %s\n", step.RunCommand)
	}

	outputPath := path + "/Dockerfile"
	err = ioutil.WriteFile(outputPath, []byte(dockerFile), 0644)
	return err
}
