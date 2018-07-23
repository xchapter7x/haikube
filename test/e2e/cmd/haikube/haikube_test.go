package haikube_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/xchapter7x/haikube/pkg/docker"
	"k8s.io/client-go/util/homedir"
)

func TestHaikube(t *testing.T) {
	gomega.RegisterTestingT(t)
	pathToHKCLI, err := gexec.Build("github.com/xchapter7x/haikube/cmd/haikube")
	defer gexec.CleanupBuildArtifacts()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	t.Run("hk build -c haikube.yml -s pathtosource", func(t *testing.T) {
		command := exec.Command(pathToHKCLI, "build", "-c", "./testdata/valid_config.yml", "-s", "./testdata/fakerepo")
		session, err := gexec.Start(command, os.Stdout, os.Stderr)
		if err != nil {
			t.Fatalf("failed running command: %v", err)
		}
		session.Wait(120 * time.Second)
		if session.ExitCode() != 0 {
			t.Errorf("call failed: %v %v %v", session.ExitCode(), string(session.Out.Contents()), string(session.Err.Contents()))
		}
	})

	t.Run("hk upload -c haikube.yml -s pathtosource", func(t *testing.T) {
		command := exec.Command(pathToHKCLI, "upload", "-c", "./testdata/valid_config.yml", "-s", "./testdata/fakerepo")
		session, err := gexec.Start(command, os.Stdout, os.Stderr)
		if err != nil {
			t.Fatalf("failed running command: %v", err)
		}
		session.Wait(600 * time.Second)
		if session.ExitCode() != 0 {
			t.Errorf("call failed: %v %v %v", session.ExitCode(), string(session.Out.Contents()), string(session.Err.Contents()))
		}
	})

	t.Run("create kubernetes assets", func(t *testing.T) {
		if v := os.Getenv("K8S_CLUSTER"); strings.ToLower(v) == "false" {
			t.Skip(`skipping k8s deployment integration b/c you do not have a configured k8s. 
						please set env var K8S_CLUSTER=true if you have a configured environment`)
		}

		t.Run("hk deploy -c ./testdata/valid_config.yml", func(t *testing.T) {
			controlDeploymentName := "unicornapp"
			command := exec.Command(pathToHKCLI, "deploy", "-c", "./testdata/valid_config.yml")
			session, err := gexec.Start(command, os.Stdout, os.Stderr)
			if err != nil {
				t.Fatalf("failed running command: %v", err)
			}

			session.Wait(600 * time.Second)
			if session.ExitCode() != 0 {
				t.Errorf("call failed: %v %v %v", session.ExitCode(), string(session.Out.Contents()), string(session.Err.Contents()))
			}

			err = deleteHelmInstall(controlDeploymentName)
			if err != nil {
				t.Fatalf("check for deployment failed: %v", err)
			}
		})

		t.Run("hk push -c ./testdata/valid_config.yml -s pathtosource", func(t *testing.T) {
			controlDeploymentName := "unicornapp"
			command := exec.Command(pathToHKCLI, "push", "-c", "./testdata/valid_config.yml", "-s", "./testdata/fakerepo")
			session, err := gexec.Start(command, os.Stdout, os.Stderr)
			if err != nil {
				t.Fatalf("failed running command: %v", err)
			}

			session.Wait(600 * time.Second)
			if session.ExitCode() != 0 {
				t.Errorf("call failed: %v %v %v", session.ExitCode(), string(session.Out.Contents()), string(session.Err.Contents()))
			}

			err = deleteHelmInstall(controlDeploymentName)
			if err != nil {
				t.Fatalf("check for deployment failed: %v", err)
			}
		})
	})
}

func deleteHelmInstall(name string) error {
	kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	kubeConfigReader, err := os.Open(kubeConfigPath)
	if err != nil {
		return fmt.Errorf("couldnt read file: %v", err)
	}

	config, err := ioutil.TempFile(".", "config")
	if err != nil {
		return fmt.Errorf("create tmp file: %v", err)
	}

	io.Copy(config, kubeConfigReader)
	defer os.Remove(config.Name())
	r := bytes.NewReader([]byte(`
FROM dtzar/helm-kubectl
WORKDIR /root
COPY ` + config.Name() + ` /root/.kube/config 
ENV KUBECONFIG /root/.kube/config
RUN helm init 
RUN helm ls | grep "` + name + `"
RUN helm del ` + name + ` --purge 
`))
	err = docker.RunDockerfileInTmpImage(r)
	if err != nil {
		return fmt.Errorf("build image failed: %v", err)
	}
	return nil
}
