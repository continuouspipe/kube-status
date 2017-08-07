package storage

import (
	"time"
	"github.com/continuouspipe/kube-status/datasnapshots"
	"github.com/satori/go.uuid"
)

type ClusterStatusHistoryEntry struct {
	UUID 			  string
	ClusterIdentifier string
	EntryTime 		  time.Time
}

type ClusterStatusHistory interface {
	Save(clusterIdentifier string, time time.Time, response datasnapshots.ClusterFullStatusResponse) (uuid.UUID, error)
	EntriesByCluster(clusterIdentifier string, left time.Time, right time.Time) ([]*ClusterStatusHistoryEntry, error)
	Fetch(identifier uuid.UUID) (datasnapshots.ClusterFullStatusResponse, error)
	RemoveEntriesBefore(datetime time.Time) (int, error)
}
