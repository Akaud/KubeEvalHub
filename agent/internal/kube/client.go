package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Clients struct {
	Core    kubernetes.Interface
	Metrics metricsclient.Interface
}

func NewInClusterClients() (*Clients, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	coreClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	metricsClient, err := metricsclient.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &Clients{
		Core:    coreClient,
		Metrics: metricsClient,
	}, nil
}
