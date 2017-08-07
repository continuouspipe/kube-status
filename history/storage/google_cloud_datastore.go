package storage

import (
	"cloud.google.com/go/datastore"
	"google.golang.org/api/option"
	"golang.org/x/net/context"
	"time"
	"os"
	"github.com/continuouspipe/kube-status/datasnapshots"
	"github.com/satori/go.uuid"
	"k8s.io/kubernetes/pkg/util/json"
)

var googleCloudProjectId, _ = os.LookupEnv("GOOGLE_CLOUD_PROJECT_ID")
var UuidNamespace, _ = uuid.FromString("0bcaf5df-8117-440c-96f2-2f5499054299")

type GoogleCloudDatastoreStatusHistory struct{}

type ClusterStatusHistoryEntryInGoogleCloudDataStore struct {
	UUID 			  string
	ClusterIdentifier string
	JsonEncodedStatus []byte `datastore:",noindex"`
	EntryTime 		  time.Time
}

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
	entry := &ClusterStatusHistoryEntryInGoogleCloudDataStore{
		UUID: entryUuid.String(),
		ClusterIdentifier: clusterIdentifier,
		EntryTime: time,
		JsonEncodedStatus: jsonBytes,
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

	if err != nil {
		return []*ClusterStatusHistoryEntry{}, err
	}

	// Create a query to fetch all Task entities, ordered by "created".
	var entriesInDatastore []*ClusterStatusHistoryEntryInGoogleCloudDataStore
	query := datastore.NewQuery("HistoryEntry").Order("EntryTime").Filter("EntryTime > ", left).Filter("EntryTime < ", right).Filter("ClusterIdentifier = ", clusterIdentifier)
	_, err = client.GetAll(gds.ClientContext(), query, &entriesInDatastore)
	if err != nil {
		return []*ClusterStatusHistoryEntry{}, err
	}

	// Remove the content from the data
	entries := make([]*ClusterStatusHistoryEntry, len(entriesInDatastore))
	for i, entryInDatastore := range entriesInDatastore {
		entries[i] = &ClusterStatusHistoryEntry{
			UUID: entryInDatastore.UUID,
			ClusterIdentifier: entryInDatastore.ClusterIdentifier,
			EntryTime: entryInDatastore.EntryTime,
		}
	}

	return entries, nil
}

func (gds *GoogleCloudDatastoreStatusHistory) Fetch(identifier uuid.UUID) (datasnapshots.ClusterFullStatusResponse, error) {
	client, err := gds.Client()
	if err != nil {
		return datasnapshots.ClusterFullStatusResponse{}, err
	}

	var entry ClusterStatusHistoryEntryInGoogleCloudDataStore
	err = client.Get(gds.ClientContext(), &datastore.Key{Kind: "HistoryEntry", Name: identifier.String()}, &entry)
	if err != nil {
		return datasnapshots.ClusterFullStatusResponse{}, err
	}

	var snapshot datasnapshots.ClusterFullStatusResponse
	err = json.Unmarshal(entry.JsonEncodedStatus, &snapshot)
	if err != nil {
		return datasnapshots.ClusterFullStatusResponse{}, err
	}

	return snapshot, nil
}

func (gds *GoogleCloudDatastoreStatusHistory) Client() (*datastore.Client, error) {
	return datastore.NewClient(gds.ClientContext(), googleCloudProjectId, option.WithServiceAccountFile("var/service-account.json"))
}

func (gds *GoogleCloudDatastoreStatusHistory) ClientContext() (context.Context) {
	return context.Background()
}
