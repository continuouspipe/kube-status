package clustersprovider
import "fmt"

type CPClusterList struct{}

func NewCPClusterList() *CPClusterList {
	return &CPClusterList{}
}

func (c CPClusterList) Clusters() []Cluster {
	// TODO
	return []Cluster{
	}
}

func (c CPClusterList) ByIdentifier(identifier string) (Cluster, error) {
	return Cluster{}, fmt.Errorf("Not implemented yet")
}
