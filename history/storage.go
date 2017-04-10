package history

import (
	"cloud.google.com/go/datastore"
	"google.golang.org/api/option"
	"fmt"
	"github.com/continuouspipe/kube-status/errors"
	"golang.org/x/net/context"
	"time"
	"github.com/continuouspipe/kube-status/datasnapshots"
	"github.com/satori/go.uuid"
	"k8s.io/kubernetes/pkg/util/json"
)

//var clusterProjectId, _ = os.LookupEnv("continuous-pipe-1042")
var clusterProjectId = "continuous-pipe-1042"
var UuidNamespace, _ = uuid.FromString("0bcaf5df-8117-440c-96f2-2f5499054299")

//BucketObjectWriter writes []bytes, if an error occurs it returns a list of errors
type BucketObjectWriter interface {
	Write([]byte) errors.ErrorListProvider
}

type ClusterStatusHistoryEntry struct {
	UUID 			  string
	ClusterIdentifier string
	JsonEncodedStatus []byte `datastore:",noindex"`
	EntryTime 		  time.Time
}

type ClusterStatusHistory interface {
	Save(clusterIdentifier string, time time.Time, response datasnapshots.ClusterFullStatusResponse) (uuid.UUID, error)
	EntriesByCluster(clusterIdentifier string, left time.Time, right time.Time) ([]*ClusterStatusHistoryEntry, error)
	Fetch(entry uuid.UUID) (ClusterStatusHistoryEntry, error)
}

//KubeStatusBucket allows to handle the kubernates status information stored on the google bucket
type GoogleCloudDatastoreStatusHistory struct{}

//NewKubeStatusBucket ctor for KubeStatusBucket
func NewGoogleCloudDatastoreStatusHistory() *GoogleCloudDatastoreStatusHistory {
	return &GoogleCloudDatastoreStatusHistory{}
}

func (gds *GoogleCloudDatastoreStatusHistory) Save(clusterIdentifier string, time time.Time, response datasnapshots.ClusterFullStatusResponse) (uuid.UUID, error) {
	client, err := gds.Client()
	if err != nil {
		return uuid.Nil, err
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return uuid.Nil, err
	}

	entryUuid := uuid.NewV5(UuidNamespace, clusterIdentifier+time.String())
	entry := &ClusterStatusHistoryEntry{
		UUID: entryUuid.String(),
		ClusterIdentifier: clusterIdentifier,
		JsonEncodedStatus: jsonBytes,
		EntryTime: time,
	}

	key := &datastore.Key{
		Kind: "HistoryEntry",
		Name: entry.UUID,
	}

	_, err = client.Put(gds.ClientContext(), key, entry)
	if err != nil {
		return uuid.Nil, err
	}

	return entryUuid, nil
}

func (gds *GoogleCloudDatastoreStatusHistory) EntriesByCluster(clusterIdentifier string, left time.Time, right time.Time) ([]*ClusterStatusHistoryEntry, error) {
	client, err := gds.Client()
	var entries []*ClusterStatusHistoryEntry

	if err != nil {
		return entries, err
	}

	// Create a query to fetch all Task entities, ordered by "created".
	query := datastore.NewQuery("HistoryEntry").Order("EntryTime").Filter("EntryTime > ", left).Filter("EntryTime < ", right)
	_, err = client.GetAll(gds.ClientContext(), query, &entries)
	if err != nil {
		return entries, err
	}

	// Set the id field on each Task from the corresponding key.
	//for i, key := range keys {
	//	entries[i].UUID = uuid.FromString(key.Name)
	//}

	return entries, nil
}

func (gds *GoogleCloudDatastoreStatusHistory) Fetch(entry uuid.UUID) (ClusterStatusHistoryEntry, error) {
	return ClusterStatusHistoryEntry{}, fmt.Errorf("Blah")
}

func (gds *GoogleCloudDatastoreStatusHistory) Client() (*datastore.Client, error) {
	return datastore.NewClient(gds.ClientContext(), clusterProjectId, option.WithServiceAccountFile("var/service-account.json"))
}

func (gds *GoogleCloudDatastoreStatusHistory) ClientContext() (context.Context) {
	return context.Background()
}
