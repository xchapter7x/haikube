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
	buildConfig = build.Flag("config", "Build config file path").Short('c').String()
	sourceDir   = build.Flag("source", "path to your code").Short('s').String()
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
		sourcePath, err := filepath.Abs(*sourceDir)
		if err != nil {
			log.Panicf("absolute path to your code not found: %v", err)
		}

		err = os.Chdir(sourcePath)
		if err != nil {
			log.Panicf("chdir failed: %v", err)
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
			log.Panicf("create dockerfile failed: %v", err)
		}

		err = docker.BuildImage(
			dockerfileReader,
			fmt.Sprintf("%s:%s", cfg.Image, cfg.Tag),
		)
		if err != nil {
			log.Panic("build image fialed: ", err)
		}
	}
}
