package clustersprovider

import (
	"os"
	"encoding/json"
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

func (c InMemoryClusterList) Clusters() []Cluster {
	return c.ClusterList
}
