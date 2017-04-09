//Package clustersprovider contains the code that retrieves a list of cluster details: host address and authentication
package clustersprovider

//ClusterRequested contains the information required to fetch the cluster status
type Cluster struct {
	Identifier string `json:"identifier"`
	Address  string   `json:"address"`
	Username string   `json:"username"`
	Password string   `json:"password"`
}

//ClusterListProvider returns a list of clusters
type ClusterListProvider interface {
	Clusters() []Cluster
}
