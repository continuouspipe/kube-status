package clustersprovider

import (
	"fmt"
	"k8s.io/kubernetes/pkg/client/restclient"
)

type WithinKubernetesClusterList struct{
}

func NewWithinKubernetesClusterList() *WithinKubernetesClusterList {
	return &WithinKubernetesClusterList{}
}

func (c WithinKubernetesClusterList) ByIdentifier(identifier string) (Cluster, error) {
	clusters, err := c.Clusters()
	if err != nil {
		return Cluster{}, err
	}

	for _, cluster := range clusters {
		if cluster.Identifier == identifier {
			return cluster, nil
		}
	}

	return Cluster{}, fmt.Errorf("Cluster not found")
}

func (c WithinKubernetesClusterList) Clusters() ([]Cluster, error) {
	cluster, err := c.GetCluster()
	if err != nil {
		return []Cluster{}, err
	}

	return []Cluster{
		cluster,
	}, nil
}

func (c WithinKubernetesClusterList) GetCluster() (Cluster, error) {
	inClusterConfig, err := restclient.InClusterConfig()
	if err != nil {
		return Cluster{}, err
	}

	return Cluster{
		Identifier: "k8s",
		Address: inClusterConfig.Host,
		Token: inClusterConfig.BearerToken,
	}, nil
}
