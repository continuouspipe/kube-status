//Package datasnapshots - writer requests the kubernetes clusters status every X minutes and stores it in a Google Cloud Storage bucket
package datasnapshots

import (
	"github.com/continuouspipe/kube-status/clustersprovider"
	"github.com/continuouspipe/kube-status/errors"
//	"github.com/golang/glog"
	"time"
)

//DataSnapshotHandler handles the data snapshot
type DataSnapshotHandler struct {
	clusterListProvider clustersprovider.ClusterListProvider
	clusterSnapshooter  ClusterSnapshooter
}

//NewDataSnapshotHandler ctor for DataSnapshotHandler
func NewDataSnapshotHandler(clp clustersprovider.ClusterListProvider, cs ClusterSnapshooter) *DataSnapshotHandler {
	h := &DataSnapshotHandler{}
	h.clusterListProvider = clp
	h.clusterSnapshooter = cs
	return h
}

//Handle each 2 minutes takes the list of cluster, fetch the data and stores it into a google bucket
func (h DataSnapshotHandler) Handle() {
	ticker := time.NewTicker(time.Minute * 2)
	for t := range ticker.C {
		go func(clp clustersprovider.ClusterListProvider, cs ClusterSnapshooter) {
			el := errors.NewErrorList()
			el.AddErrorf("error occured when getting the kubernetes status snapshot at %s", t)

			clusters := h.clusterListProvider.Clusters()

			for _, cluster := range clusters {
				h.clusterSnapshooter.Add(cluster)
			}

//			_, elr := h.clusterSnapshooter.Fetch()
//			if len(elr.Items()) > 0 {
//				el.AddErrorf("error occured when fetching the kubernetes status snapshot")
//				el.Add(elr.Items()...)
//				glog.Error(el)
//				glog.Flush()
//				return
//			}

			//store res in bucket

		}(h.clusterListProvider, h.clusterSnapshooter)
	}
}
