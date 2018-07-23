package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	docker "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	dclient "github.com/xchapter7x/haikube/pkg/docker"
)

func TestBuildImage(t *testing.T) {
	t.Run("Build dockerimage from reader", func(t *testing.T) {
		cli, err := docker.NewEnvClient()
		if err != nil {
			t.Fatalf("couldnt create client: %v", err)
		}

		testImageName := "myimage:1.2.3"
		r := bytes.NewReader([]byte(`FROM busybox
RUN echo 'hi there'`))
		err = dclient.BuildImage(r, testImageName)
		if err != nil {
			t.Fatalf("build image failed: %v", err)
		}

		var imageID string
		defer func() {
			if imageID != "" {
				fmt.Println("cleaning up", imageID)
				_, err := cli.ImageRemove(context.Background(), imageID, types.ImageRemoveOptions{Force: true})
				if err != nil {
					t.Errorf("failed cleaning up test: %v", err)
				}
			}
		}()

		images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
		if err != nil {
			t.Fatalf("unable to list images: %v", err)
		}

		found := false
		for _, image := range images {
			for _, tag := range image.RepoTags {
				if tag == testImageName {
					found = true
					imageID = image.ID
				}
			}
		}
		if !found {
			t.Errorf("couldnt find image %s", testImageName)
		}
	})
}
