package haikube_test

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/xchapter7x/haikube/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

			client, err := k8s.NewDeploymentsClient("")
			if err != nil {
				t.Fatalf("failed creating client: %v", err)
			}

			defer deploymentCleanup("unicornapp", client)
			session.Wait(600 * time.Second)
			if session.ExitCode() != 0 {
				t.Errorf("call failed: %v %v %v", session.ExitCode(), string(session.Out.Contents()), string(session.Err.Contents()))
			}

			found, err := checkForDeploymentName(controlDeploymentName, client)
			if err != nil {
				t.Fatalf("check for deployment failed: %v", err)
			}

			if !found {
				t.Errorf("didnt find deployment %s", controlDeploymentName)
			}
		})

		t.Run("hk push -c ./testdata/valid_config.yml -s pathtosource", func(t *testing.T) {
			controlDeploymentName := "unicornapp"
			command := exec.Command(pathToHKCLI, "push", "-c", "./testdata/valid_config.yml", "-s", "./testdata/fakerepo")
			session, err := gexec.Start(command, os.Stdout, os.Stderr)
			if err != nil {
				t.Fatalf("failed running command: %v", err)
			}

			client, err := k8s.NewDeploymentsClient("")
			if err != nil {
				t.Fatalf("failed creating client: %v", err)
			}

			defer deploymentCleanup("unicornapp", client)
			session.Wait(600 * time.Second)
			if session.ExitCode() != 0 {
				t.Errorf("call failed: %v %v %v", session.ExitCode(), string(session.Out.Contents()), string(session.Err.Contents()))
			}

			found, err := checkForDeploymentName(controlDeploymentName, client)
			if err != nil {
				t.Fatalf("check for deployment failed: %v", err)
			}

			if !found {
				t.Errorf("didnt find deployment %s", controlDeploymentName)
			}
		})
	})
}

func checkForDeploymentName(name string, client k8s.DeploymentInterface) (bool, error) {
	list, err := client.List(metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	found := false
	for _, d := range list.Items {
		if d.Name == name {
			found = true
		}
	}
	return found, nil
}

func deploymentCleanup(name string, client k8s.DeploymentInterface) {
	deletePolicy := metav1.DeletePropagationForeground
	if err := client.Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		log.Fatalf("couldnt cleanup from test: %v", err)
	}
}
