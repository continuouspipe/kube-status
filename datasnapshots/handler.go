//Package datasnapshots - writer requests the kubernetes clusters status every X minutes and stores it in a Google Cloud Storage bucket
package datasnapshots

//DataSnapshotHandler handles the data snapshot
type DataSnapshotHandler struct{}

func NewDataSnapshotHandler() *DataSnapshotHandler {
	return &DataSnapshotHandler{}
}

func (h DataSnapshotHandler) Handle() error {

	//get the list of clusters


	//for each cluster use the ClusterSnapshooter to get the data

	//fetch the data and write it in the google cloud bucket

	return nil
}
