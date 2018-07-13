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
	build       = kingpin.Command("build", "Build a container image from a buildpack and your code")
	buildConfig = build.Flag("file", "Build config file path").Short('f').String()
	push        = kingpin.Command("push", "Push your image to dockerhub.")
	deploy      = kingpin.Command("deploy", "Deploy your application container to kubernetes.")
	make        = kingpin.Command("make", "Build Push and Deploy your code")
)

func main() {
	switch kingpin.Parse() {
	case build.FullCommand():
		fmt.Println("Creating image from your code")
		cfg := new(haikube.Config)
		yamlFilePath, err := filepath.Abs(*buildConfig)
		if err != nil {
			log.Panicf("absolute path to config not found: %v", err)
		}

		f, err := os.Open(yamlFilePath)
		if err != nil {
			log.Panicf("file read failed %v", err)
		}
		cfg.Parse(f)
		dockerfileReader, cleanup, err := docker.CreateDockerfile(cfg.Buildpack, cfg.BaseImage, fmt.Sprintf("%v", cfg.Ports[0]), ".", docker.URIDownloader)
		defer cleanup()
		if err != nil {
			log.Panicf("create dockerfile failed: %v", err)
		}

		if err := docker.BuildImage(dockerfileReader, fmt.Sprintf("%s:%s", cfg.Image, cfg.Tag)); err != nil {
			log.Panic(err)
		}
	}
}
