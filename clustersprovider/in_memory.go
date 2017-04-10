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
	for _, cluster := range c.Clusters() {
		if cluster.Identifier == identifier {
			return cluster, nil
		}
	}

	return Cluster{}, fmt.Errorf("Cluster not found")
}

func (c InMemoryClusterList) Clusters() []Cluster {
	return c.ClusterList
}
