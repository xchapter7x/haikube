package integration_test

import (
	"os"
	"testing"

	dclient "github.com/xchapter7x/haikube/pkg/docker"
)

func TestURIDownloader(t *testing.T) {
	t.Run("downloader should grab a file and put it in a tmp path", func(t *testing.T) {
		testpath, err := dclient.URIDownloader("https://github.com/cloudfoundry/go-buildpack/releases/download/v1.8.22/go-buildpack-v1.8.22.zip")
		defer os.Remove(testpath)
		if err != nil {
			t.Fatal("we didnt expect an error but got: ", err)
		}

		fi, err := os.Stat(testpath)
		if err != nil {
			t.Fatal("we should not recieve an error when stating the file: ", err)
		}
		if fi == nil {
			t.Fatal("file info is nil for: ", testpath)
		}

		if fi.Size() <= 0 {
			t.Fatal("the file should have contents, but is of size: ", fi.Size())
		}
	})
}
