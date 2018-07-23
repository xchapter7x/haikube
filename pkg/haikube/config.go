package haikube

import (
	"io"

	"github.com/go-yaml/yaml"
)

type Config struct {
	Name            string            `json:"name"`
	Cmd             string            `json:"cmd"`
	Image           string            `json:"image"`
	Tag             string            `json:"tag"`
	BaseImage       string            `json:"baseimage"`
	Buildpack       string            `json:"buildpack"`
	Instances       int               `json:"instances"`
	Ports           []int             `json:"ports"`
	Env             map[string]string `json:"env"`
	DeploymentPatch interface{}       `json:"deploymentpatch"`
	ServicePatch    interface{}       `json:"servicepatch"`
}

func (s *Config) Parse(f io.Reader) error {
	decoder := yaml.NewDecoder(f)
	return decoder.Decode(s)
}
