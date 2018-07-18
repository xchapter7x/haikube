package haikube_test

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
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
}
