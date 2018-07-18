package integration_test

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xchapter7x/haikube/pkg/k8s"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func TestDeployApp(t *testing.T) {
	if v := os.Getenv("K8S_CLUSTER"); strings.ToLower(v) == "false" {
		t.Skip(`skipping k8s deployment integration b/c you do not have a configured k8s. 
							please set env var K8S_CLUSTER=true if you have a configured environment`)
	}

	t.Run("create DeploymentsClient", func(t *testing.T) {
		t.Run("should create a client from default kubeconfig", func(t *testing.T) {
			_, err := k8s.NewDeploymentsClient("")
			if err != nil {
				t.Errorf("we expect a client but got error: %v", err)
			}
		})
	})

	t.Run("should create a kubernetes deployment", func(t *testing.T) {
		controlDeploymentName := "demo-deployment"
		config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
		if err != nil {
			t.Fatal(err)
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			t.Fatal(err)
		}

		deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
		deployment := fakeDeployment(controlDeploymentName)
		err = k8s.DeployApp(deployment, deploymentsClient)
		if err != nil {
			t.Errorf("failed deployapp call: %s", err)
		}
		defer deploymentCleanup(controlDeploymentName, deploymentsClient)

		list, err := deploymentsClient.List(metav1.ListOptions{})
		if err != nil {
			panic(err)
		}
		found := false
		for _, d := range list.Items {
			if d.Name == controlDeploymentName {
				found = true
			}
		}
		if !found {
			t.Errorf("didnt find deployment %s in:\n %v", controlDeploymentName, list.Items)
		}
	})
}

func int32Ptr(i int32) *int32 { return &i }
func deploymentCleanup(name string, client k8s.DeploymentInterface) {
	deletePolicy := metav1.DeletePropagationForeground
	if err := client.Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		log.Fatalf("couldnt cleanup from test: %v", err)
	}
}
func fakeDeployment(name string) k8s.Deployment {
	return k8s.Deployment{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(2),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "demo",
					},
				},
				Template: apiv1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "demo",
						},
					},
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{
							{
								Name:  "web",
								Image: "nginx:1.12",
								Ports: []apiv1.ContainerPort{
									{
										Name:          "http",
										Protocol:      apiv1.ProtocolTCP,
										ContainerPort: 80,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
