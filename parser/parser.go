package parser

import (
	"fmt"
	"io/ioutil"

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
func ConvertYAMLToDockerfile(path string) (err error) {
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
	dockerFile = dockerFile + "ADD . /ciarea\n"
	dockerFile = dockerFile + "WORKDIR /ciarea\n"
	for _, step := range dockerYml.Steps {
		dockerFile = dockerFile + fmt.Sprintf("RUN %s\n", step.RunCommand)
	}

	outputPath := path + "/Dockerfile"
	err = ioutil.WriteFile(outputPath, []byte(dockerFile), 0644)
	return err
}
