package clustersprovider

import (
	"os"
	"encoding/json"
	"fmt"
)

type InMemoryClusterList struct{
	ClusterList []Cluster
}

func NewInMemoryClusterList() *InMemoryClusterList {
	var clusters []Cluster

	clustersString := os.Getenv("CLUSTER_LIST")
	if err := json.Unmarshal([]byte(clustersString), &clusters); err != nil {
		panic(err)
	}

	return &InMemoryClusterList{
		ClusterList: clusters,
	}
}

func (c InMemoryClusterList) ByIdentifier(identifier string) (Cluster, error) {
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

func (c InMemoryClusterList) Clusters() ([]Cluster, error) {
	return c.ClusterList, nil
}
