// Package main starts up the snapshot handler as a separate go routine and the http server after all handled routes are
// added.
//
// The application uses the following enviornments variables:
//
//  KUBE_STATUS_LISTEN_ADDRESS //e.g.: https://localhost:80
//  GOOGLE_CLOUD_PLATFORM_BUCKET_NAME //e.g.: kube-status-inviqa-bucket
//
package main

import (
	"flag"
	"fmt"
	"github.com/continuouspipe/kube-status/api"
	"github.com/continuouspipe/kube-status/clustersprovider"
	"github.com/continuouspipe/kube-status/datasnapshots"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	"net/http"
	"net/url"
	"os"
	"github.com/continuouspipe/kube-status/history"
	"github.com/continuouspipe/kube-status/history/storage"
	"github.com/prometheus/common/log"
	"strconv"
)

var envListenAddress, _ = os.LookupEnv("KUBE_STATUS_LISTEN_ADDRESS") //e.g.: https://localhost:80

func main() {
	listProviderType := flag.String("cluster-provider", "in-memory", "the cluster list provider")
	historyStorageType := flag.String("history-storage-backend", "in-memory", "the history storage provider")
	flag.Parse()

	arguments := flag.Args()
	var command string
	if len(arguments) == 0 {
		command = "run"
	} else {
		command = arguments[0]
	}

	var clusterList clustersprovider.ClusterListProvider
	if "in-memory" == *listProviderType {
		clusterList = clustersprovider.NewInMemoryClusterList()
	} else if "within-k8s" == *listProviderType {
		clusterList = clustersprovider.NewWithinKubernetesClusterList()
	} else if "continuous-pipe" == *listProviderType {
		clusterList = clustersprovider.NewCPClusterList()
	} else {
		log.Fatalf("Cluster list provider '%s' do not exists", *listProviderType)
	}

	var storageBackend storage.ClusterStatusHistory
	if "in-memory" == *historyStorageType {
		storageBackend = storage.NewInMemoryStatusHistory()
	} else if "google-cloud-datastore" == *historyStorageType {
		storageBackend = storage.NewGoogleCloudDatastoreStatusHistory()
	} else {
		log.Fatalf("History storage provider '%s' do not exists", *historyStorageType)
	}

	snapshooter := datasnapshots.NewClusterSnapshot()

	fmt.Printf("Run \"%s\"\n", command)
	if "run" == command {
		StartHistoryHandler(snapshooter, clusterList, storageBackend)
		StartApi(snapshooter, clusterList, storageBackend)
	} else if "history" == command {
		StartHistoryHandler(snapshooter, clusterList, storageBackend)
	} else if "api" == command {
		StartApi(snapshooter, clusterList, storageBackend)
	} else if "snapshot" == command {
		Snapshot(snapshooter, clusterList, storageBackend)
	} else {
		fmt.Printf("Command \"%s\"not found", command)
		os.Exit(1)
	}
}

func StartHistoryHandler(snapshooter datasnapshots.ClusterSnapshooter, clusterList clustersprovider.ClusterListProvider, storageBackend storage.ClusterStatusHistory) {
	hoursOfHistoryToKeep := os.Getenv("ONLY_KEEP_HOURS_OF_HISTORY")
	if "" != hoursOfHistoryToKeep {
		hours, err := strconv.ParseInt(hoursOfHistoryToKeep, 10, 64)
		if err != nil {
			log.Fatalf("Hours is not a valid integer: %s", hoursOfHistoryToKeep)
		}

		garbageCollector := history.NewGarbageCollector(storageBackend, int(hours))

		go garbageCollector.Handle()
	}

	storageHandler := NewHistoryHandler(snapshooter, clusterList, storageBackend)

	go storageHandler.Handle()
}

func Snapshot(snapshooter datasnapshots.ClusterSnapshooter, clusterList clustersprovider.ClusterListProvider, storage storage.ClusterStatusHistory) {
	storageHandler := NewHistoryHandler(snapshooter, clusterList, storage)
	storageHandler.Snapshot()
}

func NewHistoryHandler(snapshooter datasnapshots.ClusterSnapshooter, clusterList clustersprovider.ClusterListProvider, storage storage.ClusterStatusHistory) (history.DataSnapshotHandler) {
	var snapInterval int
	snapshotIntervalString := os.Getenv("SNAPSHOT_INTERVAL")
	if "" != snapshotIntervalString {
		parsedInternal, err := strconv.ParseInt(snapshotIntervalString, 10, 64)

		if err != nil {
			log.Fatalf("Hours is not a valid integer: %s", snapshotIntervalString)
		}

		snapInterval = int(parsedInternal)
	} else {
		snapInterval = 5
	}

	handler := history.NewDataSnapshotHandler(
		clusterList,
		snapshooter,
		storage,
		snapInterval,
	)

	return *handler
}

func StartApi(snapshooter datasnapshots.ClusterSnapshooter, clusterList clustersprovider.ClusterListProvider, storage storage.ClusterStatusHistory) {
	glog.Infof("Starting kube status api listening at address %s", envListenAddress)
	glog.Flush()

	listenURL, err := url.Parse(envListenAddress)
	if err != nil {
		glog.V(5).Infof("cannot parse URL: %v\n", err.Error())
		glog.Flush()
		fmt.Printf("Cannot parse URL: %v\n", err.Error())
		os.Exit(1)
	}

	clusterHandler := api.NewClusterApiHandler(snapshooter, clusterList)

	r := mux.NewRouter()

	// Clusters
	r.HandleFunc(api.BackwardCompatibleClusterFullStatusURLPath, clusterHandler.HandleBackwardCompatible).Methods(http.MethodPost)
	r.HandleFunc(api.ClusterFullStatusURLPath, clusterHandler.HandleStatus).Methods(http.MethodGet)
	r.HandleFunc(api.ClusterListURLPath, clusterHandler.HandleList).Methods(http.MethodGet)

	// History
	r.HandleFunc(api.ClusterHistoryURLPath, api.NewClusterHistoryH(storage).HandleList).Methods(http.MethodGet)
	r.HandleFunc(api.ClusterHistoryEntryURLPath, api.NewClusterHistoryH(storage).HandleEntry).Methods(http.MethodGet)

	// Static assets
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./var/static/")))

	headersOk := handlers.AllowedHeaders([]string{
		"Accept",
		"Accept-Encoding",
		"Accept-Language",
		"Content-Type",
		"Origin",
		"X-Requested-With",
	})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	handler := handlers.CORS(originsOk, headersOk, methodsOk)(r)
	http.ListenAndServe(listenURL.Host, handler)
	glog.Flush()
}
