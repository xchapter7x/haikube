package docker_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"testing"

	docker "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	uuid "github.com/satori/go.uuid"
	dclient "github.com/xchapter7x/haikube/pkg/docker"
)

func TestDockerClient(t *testing.T) {
	t.Run("Generate dockerfile", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			rdr, cleanup, err := dclient.CreateDockerfile("abc", "ubuntu", "80", ".", fakeSuccessDownloader)
			defer cleanup()
			if err != nil {
				t.Fatalf("creating dockerfile failed: %v", err)
			}
			b, _ := ioutil.ReadAll(rdr)
			validDockerfileFormat := regexp.MustCompile(`
FROM ubuntu
RUN mkdir /app /cache /deps || true
WORKDIR /app
COPY /.*/buildpack.* /buildpack
COPY . /app`)
			if !validDockerfileFormat.MatchString(string(b)) {
				t.Errorf("invalid dockerfile created: %s", string(b))
			}
		})

		t.Run("failure", func(t *testing.T) {
			_, cleanup, err := dclient.CreateDockerfile("abc", "ubuntu", "80", ".", fakeFailureDownloader)
			defer cleanup()
			if err == nil {
				t.Errorf("download failed but the creation didnt send error: %v", err)
			}
		})
	})

	t.Run("Build dockerimage from reader", func(t *testing.T) {
		cli, err := docker.NewEnvClient()
		if err != nil {
			t.Fatalf("couldnt create client: %v", err)
		}

		testImageName := "myimage:1.2.3"
		r := bytes.NewReader([]byte(`FROM ubuntu`))
		err = dclient.BuildImage(r, testImageName)
		if err != nil {
			t.Fatalf("build image failed: %v", err)
		}

		var imageID string
		defer func() {
			if imageID != "" {
				fmt.Println("cleaning up", imageID)
				cli.ImageRemove(context.Background(), imageID, types.ImageRemoveOptions{Force: true})
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

	t.Run("Upload image to docker repo", func(t *testing.T) {

	})
}

func fakeFailureDownloader(downloadURI string) (string, error) {
	return "", fmt.Errorf("we failed fakely")
}

func fakeSuccessDownloader(downloadURI string) (string, error) {
	tmpDir, _ := ioutil.TempDir("", "buildpack")
	u, _ := uuid.NewV4()
	newFilePath := tmpDir + u.String()
	copyFile("./testdata/go-buildpack-v1.8.22.zip", newFilePath)
	return newFilePath, nil
}

func copyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
