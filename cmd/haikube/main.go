package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/xchapter7x/haikube/pkg/docker"
	"github.com/xchapter7x/haikube/pkg/haikube"
	"github.com/xchapter7x/haikube/pkg/k8s"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	build           = kingpin.Command("build", "Build a container image from a buildpack and your code")
	buildConfig     = build.Flag("config", "config file path").Short('c').Required().String()
	buildSourceDir  = build.Flag("source", "path to your code").Short('s').Required().String()
	upload          = kingpin.Command("upload", "Build & Push your image to dockerhub.")
	uploadConfig    = upload.Flag("config", "config file path").Short('c').Required().String()
	uploadSourceDir = upload.Flag("source", "path to your code").Short('s').Required().String()
	deploy          = kingpin.Command("deploy", "Deploy your application container to kubernetes.")
	deployConfig    = deploy.Flag("config", "config file path").Short('c').Required().String()
	push            = kingpin.Command("push", "Build Push and Deploy your code")
	pushConfig      = push.Flag("config", "config file path").Short('c').Required().String()
	pushSourceDir   = push.Flag("source", "path to your code").Short('s').Required().String()
)

func main() {
	switch kingpin.Parse() {
	case build.FullCommand():
		_, err := buildDockerImage(*buildConfig, *buildSourceDir)
		if err != nil {
			log.Panicf("buildDockerImage failed: %v", err)
		}
	case upload.FullCommand():
		err := uploadDockerImage(*uploadConfig, *uploadSourceDir)
		if err != nil {
			log.Panicf("uploadDockerImage failed: %v", err)
		}
	case deploy.FullCommand():
		err := createDeployment(*deployConfig)
		if err != nil {
			log.Panicf("create deployment failed: %v", err)
		}
	case push.FullCommand():
		err := uploadDockerImage(*pushConfig, *pushSourceDir)
		if err != nil {
			log.Panicf("uploadDockerImage failed: %v", err)
		}
		err = createDeployment(*pushConfig)
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
	client, err := k8s.NewDeploymentsClient("")
	if err != nil {
		return fmt.Errorf("failed creating deployment: %v", err)
	}

	deployment := k8s.NewDeployment(
		cfg.Name,
		fmt.Sprintf("%s:%s", cfg.Image, cfg.Tag),
		filepath.Base(cfg.Buildpack),
		int32(cfg.Ports[0]),
	)
	return k8s.DeployApp(deployment, client)
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
