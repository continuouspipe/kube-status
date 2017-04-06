//Package datasnapshots - writer requests the kubernetes clusters status every X minutes and stores it in a Google Cloud Storage bucket
package datasnapshots

import "github.com/continuouspipe/kube-status/clustersprovider"

//DataSnapshotHandler handles the data snapshot
type DataSnapshotHandler struct {
	clusterListProvider clustersprovider.ClusterListProvider
	clusterSnapshooter ClusterSnapshooter
}

func NewDataSnapshotHandler() *DataSnapshotHandler {
	return &DataSnapshotHandler{}
}

func (h DataSnapshotHandler) Handle() error {

	clusters := h.clusterListProvider.Clusters()

	for _, cluster := range clusters {
		h.clusterSnapshooter.Add(cluster)
	}

	//for each cluster use the ClusterSnapshooter to get the data

	//fetch the data and write it in the google cloud bucket

	return nil
}
