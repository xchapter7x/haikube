package haikube_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/xchapter7x/haikube/pkg/haikube"
)

func TestConfig(t *testing.T) {
	t.Run("parsing Buildpack", func(t *testing.T) {
		t.Run("no base image defined", func(t *testing.T) {
			controlBase := "cloudfoundry/cflinuxfs2"
			c := new(haikube.Config)
			r := bytes.NewReader([]byte(`buildpack: go`))
			err := c.Parse(r)
			if err != nil {
				t.Errorf("file parse failed %v", err)
			}

			if c.BaseImage != controlBase {
				t.Errorf("base images should default to %s but has: %s", controlBase, c.BaseImage)
			}
		})

		t.Run("when value is a url", func(t *testing.T) {
			controlURI := "https://some.buildpack.org/file.zip"
			c := new(haikube.Config)
			r := bytes.NewReader([]byte(`buildpack: ` + controlURI))
			err := c.Parse(r)
			if err != nil {
				t.Errorf("file parse failed %v", err)
			}

			if c.Buildpack != controlURI {
				t.Errorf("buildpack url %s doesnt match %s", c.Buildpack, controlURI)
			}
		})

		t.Run("when value is a valid language name", func(t *testing.T) {
			controlURI := "https://github.com/cloudfoundry/go-buildpack/releases/download/v1.8.25/go-buildpack-v1.8.25.zip"
			c := new(haikube.Config)
			r := bytes.NewReader([]byte(`buildpack: go`))
			err := c.Parse(r)
			if err != nil {
				t.Errorf("file parse failed %v", err)
			}

			if c.Buildpack != controlURI {
				t.Errorf("buildpack url %s doesnt match %s", c.Buildpack, controlURI)
			}
		})

		t.Run("when value is a unknown language name", func(t *testing.T) {
			c := new(haikube.Config)
			r := bytes.NewReader([]byte(`buildpack: xxxxxx`))
			err := c.Parse(r)
			if err == nil {
				t.Errorf("we expected a buildpack error but got: %v", err)
			}
		})
	})

	t.Run("given a valid config file", func(t *testing.T) {
		c := new(haikube.Config)
		f, err := os.Open("testdata/valid_config.yml")
		if err != nil {
			t.Errorf("file read failed %v", err)
		}

		err = c.Parse(f)
		if err != nil {
			t.Errorf("file parse failed %v", err)
		}

		t.Run("it has its fields populated", func(t *testing.T) {
			for k, v := range map[string]string{
				"name":       c.Name,
				"cmd":        c.Cmd,
				"image":      c.Image,
				"tag":        c.Tag,
				"base_image": c.BaseImage,
				"buildpack":  c.Buildpack,
			} {
				if v == "" {
					t.Errorf("empty value set in object %v:%v", k, v)
				}
			}
		})
	})

	t.Run("given a invalid config file", func(t *testing.T) {
		c := new(haikube.Config)
		f, err := os.Open("testdata/invalid_config.yml")
		if err != nil {
			t.Errorf("file read failed %v", err)
		}

		err = c.Parse(f)
		if err == nil {
			t.Errorf("file was expected to fail, but didnt on %v", f)
		}
	})
}
