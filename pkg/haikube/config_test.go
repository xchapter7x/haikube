package haikube_test

import (
	"os"
	"testing"

	"github.com/xchapter7x/haikube/pkg/haikube"
)

func TestConfig(t *testing.T) {
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
