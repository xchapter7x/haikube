package k8s_test

import (
	"testing"

	"github.com/xchapter7x/haikube/pkg/k8s"
)

func TestK8sClient(t *testing.T) {
	t.Run("create deployment", func(t *testing.T) {
		controlName := "my-fake-name"
		controlImage := "my-fake-image"
		controlBuildpack := "fake-buildpack"
		controlPort := int32(80)
		d := k8s.NewDeployment(controlName, controlImage, controlBuildpack, controlPort)
		for _, table := range []struct{ name, deploymentValue, controlValue string }{
			{"meta name", d.ObjectMeta.Name, controlName},
			{"app matchlabel", d.Spec.Selector.MatchLabels["app"], controlName},
			{"app label", d.Spec.Template.ObjectMeta.Labels["app"], controlName},
			{"buildpack matchlabel", d.Spec.Selector.MatchLabels["buildpack"], controlBuildpack},
			{"buildpack label", d.Spec.Template.ObjectMeta.Labels["buildpack"], controlBuildpack},
			{"deployment container name", d.Spec.Template.Spec.Containers[0].Name, controlName},
			{"container image", d.Spec.Template.Spec.Containers[0].Image, controlImage},
			{"container port", string(d.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort), string(controlPort)},
		} {
			t.Run(table.name, func(t *testing.T) {
				if table.deploymentValue != table.controlValue {
					t.Errorf("values dont match %s != %s", table.deploymentValue, table.controlValue)
				}
			})
		}
	})
}
