package k8s

import (
	"fmt"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Deployment struct {
	*appsv1.Deployment
}

type DeploymentInterface interface {
	Create(*appsv1.Deployment) (*appsv1.Deployment, error)
	Delete(name string, options *metav1.DeleteOptions) error
	Get(name string, options metav1.GetOptions) (*appsv1.Deployment, error)
	List(opts metav1.ListOptions) (*appsv1.DeploymentList, error)
}

func DeployApp(deployment Deployment, client DeploymentInterface) error {
	fmt.Println("Creating deployment...")
	result, err := client.Create(deployment.Deployment)
	if err != nil {
		return fmt.Errorf("create deployment failed: %v", err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
	return nil
}

func NewDeploymentsClient(kubeconfig string) (DeploymentInterface, error) {
	kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	return deploymentsClient, nil
}

func NewDeployment(name, image, buildpack string, port int32) Deployment {
	deployment := Deployment{
		Deployment: &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(2),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":       name,
						"buildpack": buildpack,
					},
				},
				Template: apiv1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app":       name,
							"buildpack": buildpack,
						},
					},
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{
							{
								Name:  name,
								Image: image,
								Ports: []apiv1.ContainerPort{
									{
										Name:          "http",
										Protocol:      apiv1.ProtocolTCP,
										ContainerPort: port,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment
}

func int32Ptr(i int32) *int32 { return &i }
