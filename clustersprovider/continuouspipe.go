package clustersprovider

type CPClusterList struct{}

func NewCPClusterList() *CPClusterList {
	return &CPClusterList{}
}

func (c CPClusterList) Clusters() []Cluster {
	// TODO
	return []Cluster{
	}
}
