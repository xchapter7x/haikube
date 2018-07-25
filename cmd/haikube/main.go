package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/xchapter7x/haikube/pkg/docker"
	"github.com/xchapter7x/haikube/pkg/haikube"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	hkConfig    = kingpin.Flag("config", "config file path").Short('c').Default(".haikube.yml").String()
	hkSourceDir = kingpin.Flag("source", "path to your code").Short('s').Default(".").String()
	build       = kingpin.Command("build", "Build a container image from a buildpack and your code")
	upload      = kingpin.Command("upload", "Build & Push your image to dockerhub.")
	deploy      = kingpin.Command("deploy", "Deploy your application container to kubernetes.")
	push        = kingpin.Command("push", "Build Push and Deploy your code")
)

func main() {
	switch kingpin.Parse() {
	case build.FullCommand():
		_, err := buildDockerImage(*hkConfig, *hkSourceDir)
		if err != nil {
			log.Panicf("buildDockerImage failed: %v", err)
		}
	case upload.FullCommand():
		err := uploadDockerImage(*hkConfig, *hkSourceDir)
		if err != nil {
			log.Panicf("uploadDockerImage failed: %v", err)
		}
	case deploy.FullCommand():
		err := createDeployment(*hkConfig)
		if err != nil {
			log.Panicf("create deployment failed: %v", err)
		}
	case push.FullCommand():
		err := uploadDockerImage(*hkConfig, *hkSourceDir)
		if err != nil {
			log.Panicf("uploadDockerImage failed: %v", err)
		}
		err = createDeployment(*hkConfig)
		if err != nil {
			log.Panicf("create deployment failed: %v", err)
		}
	}
}

func createDeployment(config string) error {
	cfg := new(haikube.Config)
	yamlFilePath, err := filepath.Abs(config)
	if err != nil {
		return fmt.Errorf("absolute path to config not found: %v", err)
	}

	f, err := os.Open(yamlFilePath)
	if err != nil {
		return fmt.Errorf("file read failed %v", err)
	}

	cfg.Parse(f)
	err = docker.HelmInstall(cfg.Name, cfg.Image, cfg.Tag, fmt.Sprint(cfg.Ports[0]))
	if err != nil {
		return fmt.Errorf("helm install failed: %v", err)
	}
	return nil
}

func uploadDockerImage(config, source string) error {
	imageName, err := buildDockerImage(config, source)
	if err != nil {
		return fmt.Errorf("failed building image: %v", err)
	}

	return docker.PushImage(imageName)
}

func buildDockerImage(config, source string) (string, error) {
	fmt.Println("Creating image from your code")
	cfg := new(haikube.Config)
	yamlFilePath, err := filepath.Abs(config)
	if err != nil {
		return "", fmt.Errorf("absolute path to config not found: %v", err)
	}

	f, err := os.Open(yamlFilePath)
	if err != nil {
		return "", fmt.Errorf("file read failed %v", err)
	}

	cfg.Parse(f)
	sourcePath, err := filepath.Abs(source)
	if err != nil {
		return "", fmt.Errorf("absolute path to your code not found: %v", err)
	}
	originalPath, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cant find current path: %v", err)
	}

	defer func() {
		os.Chdir(originalPath)
	}()
	err = os.Chdir(sourcePath)
	if err != nil {
		return "", fmt.Errorf("chdir failed: %v", err)
	}

	dockerfileReader, cleanup, err := docker.CreateDockerfile(
		cfg.Buildpack,
		cfg.BaseImage,
		fmt.Sprintf("%v", cfg.Ports[0]),
		".",
		cfg.Cmd,
		cfg.Env,
		docker.URIDownloader,
	)
	defer cleanup()
	if err != nil {
		return "", fmt.Errorf("create dockerfile failed: %v", err)
	}

	fullImageName := fmt.Sprintf("%s:%s", cfg.Image, cfg.Tag)
	err = docker.BuildImage(
		dockerfileReader,
		fullImageName,
	)
	if err != nil {
		return "", fmt.Errorf("build image fialed: ", err)
	}
	return fullImageName, nil
}
