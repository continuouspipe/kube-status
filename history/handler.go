package history

import (
	"github.com/continuouspipe/kube-status/clustersprovider"
	"github.com/continuouspipe/kube-status/datasnapshots"
	"github.com/continuouspipe/kube-status/history/storage"
	"github.com/satori/go.uuid"
	"time"
	"fmt"
	"sync"
)

//DataSnapshotHandler handles the data snapshot
type DataSnapshotHandler struct {
	clusterListProvider clustersprovider.ClusterListProvider
	clusterSnapshooter  datasnapshots.ClusterSnapshooter
	storage				storage.ClusterStatusHistory
}

//NewDataSnapshotHandler ctor for DataSnapshotHandler
func NewDataSnapshotHandler(clp clustersprovider.ClusterListProvider, cs datasnapshots.ClusterSnapshooter, storage storage.ClusterStatusHistory) *DataSnapshotHandler {
	return &DataSnapshotHandler{
		clusterListProvider: clp,
		clusterSnapshooter: cs,
		storage: storage,
	}
}

//Handle each 2 minutes takes the list of cluster, fetch the data and stores it into a google bucket
func (h DataSnapshotHandler) Handle() {
	ticker := time.NewTicker(time.Minute * 2)
	for _ = range ticker.C {
		h.Snapshot()
	}
}

func (h DataSnapshotHandler) Snapshot() {
	t := time.Now()
	fmt.Println("Snapshotting clusters status", t)

	var wg sync.WaitGroup

	clusters := h.clusterListProvider.Clusters()

	for _, cluster := range clusters {
		fmt.Printf("Snapshotting cluster '%s'\n", cluster.Identifier)
		wg.Add(1)

		go func (cluster clustersprovider.Cluster, time time.Time) {
			defer wg.Done()

			snapshotUuid := h.SnapshotCluster(cluster, time)

			fmt.Printf("Stored snapshot with UUID '%s' for cluster '%s'\n", snapshotUuid.String(), cluster.Identifier)
		}(cluster, t)
	}

	wg.Wait()
	fmt.Println("Finished snapshots")
}

func (h DataSnapshotHandler) SnapshotCluster(cluster clustersprovider.Cluster, time time.Time) uuid.UUID {
	status, err := h.clusterSnapshooter.FetchCluster(cluster)

	if err != nil {
		fmt.Printf("Something wrong happened while snapshotting the cluster '%s': %s", cluster.Identifier, err)

		return uuid.Nil
	}

	savedUuid, err := h.storage.Save(cluster.Identifier, time, *status)
	if err != nil {
		fmt.Printf("Something wrong happened while storing the snapshot of cluster '%s': %s", cluster.Identifier, err)

		return uuid.Nil
	}

	return savedUuid
}
