package haikube

import (
	"fmt"
	"io"
	"net/url"

	"github.com/go-yaml/yaml"
)

type Config struct {
	Name       string                 `yaml:"name"`
	Cmd        string                 `yaml:"cmd"`
	Image      string                 `yaml:"image"`
	Tag        string                 `yaml:"tag"`
	BaseImage  string                 `yaml:"baseimage"`
	Buildpack  string                 `yaml:"buildpack"`
	Instances  int                    `yaml:"instances"`
	Ports      []int                  `yaml:"ports"`
	Env        map[string]string      `yaml:"env"`
	HelmValues map[string]interface{} `yaml:"helm_values"`
}

var supportedBuildpacks = map[string]string{
	"binary":     "https://github.com/cloudfoundry/binary-buildpack/releases/download/v1.0.21/binary-buildpack-v1.0.21.zip",
	"go":         "https://github.com/cloudfoundry/go-buildpack/releases/download/v1.8.25/go-buildpack-v1.8.25.zip",
	"java":       "https://github.com/cloudfoundry/java-buildpack/releases/download/v4.13.1/java-buildpack-v4.13.1.zip",
	"dotnetcore": "https://github.com/cloudfoundry/dotnet-core-buildpack/releases/download/v2.1.3/dotnet-core-buildpack-v2.1.3.zip",
	"node":       "https://github.com/cloudfoundry/nodejs-buildpack/releases/download/v1.6.29/nodejs-buildpack-v1.6.29.zip",
	"php":        "https://github.com/cloudfoundry/php-buildpack/releases/download/v4.3.58/php-buildpack-v4.3.58.zip",
	"python":     "https://github.com/cloudfoundry/python-buildpack/releases/download/v1.6.19/python-buildpack-v1.6.19.zip",
	"ruby":       "https://github.com/cloudfoundry/ruby-buildpack/releases/download/v1.7.21/ruby-buildpack-v1.7.21.zip",
	"staticfile": "https://github.com/cloudfoundry/staticfile-buildpack/releases/download/v1.4.30/staticfile-buildpack-v1.4.30.zip",
	"nginx":      "https://github.com/cloudfoundry/nginx-buildpack/releases/download/v1.0.0/nginx-buildpack-v1.0.0.zip",
}

func (s *Config) Parse(f io.Reader) error {
	decoder := yaml.NewDecoder(f)
	err := decoder.Decode(s)
	if err != nil {
		return fmt.Errorf("could not decode yaml: %v", err)
	}

	err = s.translateBuildpackURL()
	if err != nil {
		return fmt.Errorf("failed parsing buildpack: %v", err)
	}

	if s.BaseImage == "" {
		s.BaseImage = "cloudfoundry/cflinuxfs2"
	}
	return nil
}

func (s *Config) translateBuildpackURL() error {
	buildpackURI, ok := supportedBuildpacks[s.Buildpack]
	if ok {
		s.Buildpack = buildpackURI
		return nil
	}

	_, err := url.ParseRequestURI(s.Buildpack)
	if err != nil {
		return fmt.Errorf("buildpack url invalid: %v", err)
	}
	return nil
}
