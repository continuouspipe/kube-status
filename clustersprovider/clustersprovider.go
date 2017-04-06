//Package clustersprovider contains the code that retrieves a list of cluster details: host address and authentication
package clustersprovider

//ClusterRequested contains the information required to fetch the cluster status
type Cluster struct {
	Address  string
	Username string
	Password string
}

//ClusterListProvider returns a list of clusters
type ClusterListProvider interface {
	Clusters() []Cluster
}

//CPClusterList returns a list of CP Clusters
type CPClusterList struct{}

func (c CPClusterList) Clusters() []Cluster {
	//TODO: return a list of Clusters
}
