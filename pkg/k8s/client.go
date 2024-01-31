package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	dv_client "kubevirt.io/client-go/generated/containerized-data-importer/clientset/versioned"
	dv_scheme "kubevirt.io/client-go/generated/containerized-data-importer/clientset/versioned/scheme"
)

type KubernetesClient struct {
	Client   *kubernetes.Clientset
	KubeVirt *dv_client.Clientset
	Config   *rest.Config
}

func NewClient(kubeconfig string) (*KubernetesClient, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)

	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	dvClient, err := dv_client.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	err = dv_scheme.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, err
	}

	return &KubernetesClient{
		Client:   client,
		KubeVirt: dvClient,
		Config:   config,
	}, nil

}
