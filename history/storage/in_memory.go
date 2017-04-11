package storage

import (
	"time"
	"github.com/continuouspipe/kube-status/datasnapshots"
	"github.com/satori/go.uuid"
	"container/ring"
	"fmt"
)

type InMemoryStatusHistory struct{
	Ring *ring.Ring
}

type ClusterStatusInRing struct {
	entry  ClusterStatusHistoryEntry
	status datasnapshots.ClusterFullStatusResponse
}

func NewInMemoryStatusHistory() *InMemoryStatusHistory {
	ring := ring.New(720)

	return &InMemoryStatusHistory{
		Ring: ring,
	}
}

func (h *InMemoryStatusHistory) Save(clusterIdentifier string, time time.Time, response datasnapshots.ClusterFullStatusResponse) (uuid.UUID, error) {
	uuid := uuid.NewV4()

	h.Ring.Value = ClusterStatusInRing{
		entry: ClusterStatusHistoryEntry{
			UUID: uuid.String(),
			ClusterIdentifier: clusterIdentifier,
			EntryTime: time,
		},
		status: response,
	}

	h.Ring = h.Ring.Next()

	return uuid, nil
}

func (h InMemoryStatusHistory) EntriesByCluster(clusterIdentifier string, left time.Time, right time.Time) ([]*ClusterStatusHistoryEntry, error) {
	entries := []*ClusterStatusHistoryEntry{}

	h.Ring.Do(func(entryInRing interface{}) {
		clusterStatusInRing, ok := entryInRing.(ClusterStatusInRing)
		if !ok {
			return
		}

		if clusterStatusInRing.entry.ClusterIdentifier == clusterIdentifier && left.Before(clusterStatusInRing.entry.EntryTime) && right.After(clusterStatusInRing.entry.EntryTime) {
			entries = append(entries, &clusterStatusInRing.entry)
		}
	})

	return entries, nil
}

func (h InMemoryStatusHistory) Fetch(identifier uuid.UUID) (datasnapshots.ClusterFullStatusResponse, error) {
	var status *datasnapshots.ClusterFullStatusResponse

	h.Ring.Do(func(entryInRing interface{}) {
		clusterStatusInRing, ok := entryInRing.(ClusterStatusInRing)
		if !ok {
			return
		}

		if identifier.String() == clusterStatusInRing.entry.UUID {
			status = &clusterStatusInRing.status
		}
	})

	if nil != status {
		return *status, nil
	}

	return datasnapshots.ClusterFullStatusResponse{}, fmt.Errorf("Not found")
}
