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
	for _, cluster := range c.Clusters() {
		if cluster.Identifier == identifier {
			return cluster, nil
		}
	}

	return Cluster{}, fmt.Errorf("Cluster not found")
}

func (c WithinKubernetesClusterList) Clusters() []Cluster {
	cluster, err := c.GetCluster()
	if err != nil {
		return []Cluster{}
	}

	return []Cluster{
		cluster,
	}
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
