package docker_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"

	uuid "github.com/satori/go.uuid"
	dclient "github.com/xchapter7x/haikube/pkg/docker"
)

func TestDockerClient(t *testing.T) {
	t.Run("Generate dockerfile", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			rdr, cleanup, err := dclient.CreateDockerfile("abc", "ubuntu", "80", ".", make(map[string]string), fakeSuccessDownloader)
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
			_, cleanup, err := dclient.CreateDockerfile("abc", "ubuntu", "80", ".", make(map[string]string), fakeFailureDownloader)
			defer cleanup()
			if err == nil {
				t.Errorf("download failed but the creation didnt send error: %v", err)
			}
		})
	})

	t.Run("Upload image to docker repo", func(t *testing.T) {
		t.Run("should fail if username or password is not set", func(t *testing.T) {
			restoreEnv := clearEnvironment()
			defer restoreEnv()
			err := dclient.PushImage("alpine")
			if err == nil {
				t.Errorf("we expect this to fail if docker user/pass are not set: %v", err)
			}
		})
	})
}

func clearEnvironment() func() {
	environ := os.Environ()
	os.Clearenv()
	return func() {
		os.Clearenv()
		for _, e := range environ {
			arr := strings.Split(e, "=")
			os.Setenv(arr[0], arr[1])
		}
	}
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
