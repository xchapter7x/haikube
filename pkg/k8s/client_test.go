package k8s_test

import (
	"testing"

	"github.com/xchapter7x/haikube/pkg/k8s"
)

func TestK8sClient(t *testing.T) {
	t.Run("create DeploymentsClient", func(t *testing.T) {
		t.Run("should create a client from default kubeconfig", func(t *testing.T) {
			_, err := k8s.NewDeploymentsClient("")
			if err != nil {
				t.Errorf("we expect a client but got error: %v", err)
			}
		})
	})

	t.Run("create deployment", func(t *testing.T) {
		controlName := "my-fake-name"
		controlImage := "my-fake-image"
		controlPort := int32(80)
		d := k8s.NewDeployment(controlName, controlImage, controlPort)
		t.Run("should have proper names and labels", func(t *testing.T) {
			if d.ObjectMeta.Name != controlName {
				t.Errorf("name doesnt match %s != %s", d.ObjectMeta.Name, controlName)
			}
			if d.Spec.Selector.MatchLabels["app"] != controlName {
				t.Errorf("label doesnt match %s != %s", d.Spec.Selector.MatchLabels["app"], controlName)
			}
			if d.Spec.Template.ObjectMeta.Labels["app"] != controlName {
				t.Errorf("label doesnt match %s != %s", d.Spec.Template.ObjectMeta.Labels["app"], controlName)
			}
			if d.Spec.Template.Spec.Containers[0].Name != controlName {
				t.Errorf("image doesnt match %s != %s", d.Spec.Template.Spec.Containers[0].Name, controlName)
			}
		})

		t.Run("should have proper image", func(t *testing.T) {
			if d.Spec.Template.Spec.Containers[0].Image != controlImage {
				t.Errorf("image doesnt match %s != %s", d.Spec.Template.Spec.Containers[0].Image, controlImage)
			}
		})

		t.Run("should have proper port info", func(t *testing.T) {
			if d.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort != controlPort {
				t.Errorf("port doesnt match %v != %v", d.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort, controlPort)
			}
		})
	})
}
