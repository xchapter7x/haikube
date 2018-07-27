package integration_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	uuid "github.com/satori/go.uuid"
	dclient "github.com/xchapter7x/haikube/pkg/docker"
	"k8s.io/client-go/util/homedir"
)

func TestRunDockerfileInTmpImage(t *testing.T) {
	t.Run("should run the docker file and cleanup without error", func(t *testing.T) {
		r := bytes.NewReader([]byte(`FROM busybox
RUN echo 'yo yo yo'`))
		err := dclient.RunDockerfileInTmpImage(r)
		if err != nil {
			t.Fatalf("build image failed: %v", err)
		}
	})

	t.Run("should run a helm flow and exit without error", func(t *testing.T) {

		if v := os.Getenv("K8S_CLUSTER"); strings.ToLower(v) == "false" {
			t.Skip(`skipping k8s deployment integration b/c you do not have a configured k8s. 
							please set env var K8S_CLUSTER=true if you have a configured environment`)
		}

		guid, err := uuid.NewV4()
		if err != nil {
			t.Fatalf("failed generating a guid: %v", err)
		}
		controlName := strings.ToLower("hktest-" + guid.String())

		err = dclient.HelmInstall(controlName, "nginx", "1.15.1", fmt.Sprint(80), nil)
		if err != nil {
			t.Errorf("helm install failed: %v", err)
		}

		err = deleteHelmInstall(controlName)
		if err != nil {
			t.Fatalf("helm cleanup failed: %v", err)
		}
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
	err = dclient.RunDockerfileInTmpImage(r)
	if err != nil {
		return fmt.Errorf("build image failed: %v", err)
	}
	return nil
}
