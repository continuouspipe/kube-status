package history

import (
    "github.com/continuouspipe/kube-status/history/storage"
    "time"
    "fmt"
)

//GarbageCollector handles the data snapshot
type GarbageCollector struct {
    storage     storage.ClusterStatusHistory
    hoursToKeep int
}

//NewGarbageCollector create a GarbageCollector
func NewGarbageCollector(storage storage.ClusterStatusHistory, hoursToKeep int) *GarbageCollector {
    return &GarbageCollector{
        storage: storage,
        hoursToKeep: hoursToKeep,
    }
}

//Handle each 5 minutes while garbage collect the entries
func (gc *GarbageCollector) Handle() {
    gc.GarbageCollect()

    ticker := time.NewTicker(time.Minute * 5)
    for _ = range ticker.C {
        gc.GarbageCollect()
    }
}
func (gc *GarbageCollector) GarbageCollect() {
    removedEntries, err := gc.DoGarbageCollect()

    if err != nil {
        fmt.Printf("Error: could not garbage collect: %s\n", err.Error())
    } else {
        fmt.Printf("Garbage collected %d entries\n", removedEntries)
    }
}

func (gc *GarbageCollector) DoGarbageCollect() (int, error) {
    return gc.storage.RemoveEntriesBefore(
        time.Now().Add(-1 * (time.Duration(gc.hoursToKeep) * time.Hour)),
    )
}
