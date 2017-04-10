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
)

var envListenAddress, _ = os.LookupEnv("KUBE_STATUS_LISTEN_ADDRESS") //e.g.: https://localhost:80

func main() {
	listProviderType := flag.String("cluster-list", "in-memory", "the cluster list provider")
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
	} else {
		clusterList = clustersprovider.NewCPClusterList()
	}

	snapshooter := datasnapshots.NewClusterSnapshot()

	fmt.Printf("Run \"%s\"\n", command)
	if "run" == command {
		StartHistoryHandler(snapshooter, clusterList)
		StartApi(snapshooter, clusterList)
	} else if "snapshot" == command {
		Snapshot(snapshooter, clusterList)
	} else {
		fmt.Printf("Command \"%s\"not found", command)
		os.Exit(1)
	}
}

func StartHistoryHandler(snapshooter datasnapshots.ClusterSnapshooter, clusterList clustersprovider.ClusterListProvider) {
	storageHandler := NewHistoryHandler(snapshooter, clusterList)

	go storageHandler.Handle()
}

func Snapshot(snapshooter datasnapshots.ClusterSnapshooter, clusterList clustersprovider.ClusterListProvider) {
	storageHandler := NewHistoryHandler(snapshooter, clusterList)
	storageHandler.Snapshot()
}

func NewHistoryHandler(snapshooter datasnapshots.ClusterSnapshooter, clusterList clustersprovider.ClusterListProvider) (history.DataSnapshotHandler) {
	storage := history.NewGoogleCloudDatastoreStatusHistory()
	handler := history.NewDataSnapshotHandler(
		clusterList,
		snapshooter,
		storage,
	)

	return *handler
}

func StartApi(snapshooter datasnapshots.ClusterSnapshooter, clusterList clustersprovider.ClusterListProvider) {
	glog.Infof("Starting kube status api listening at address %s", envListenAddress)
	glog.Flush()

	listenURL, err := url.Parse(envListenAddress)
	if err != nil {
		glog.V(5).Infof("cannot parse URL: %v\n", err.Error())
		glog.Flush()
		fmt.Printf("Cannot parse URL: %v\n", err.Error())
		os.Exit(1)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", rootHandle)
	r.HandleFunc(api.ClusterFullStatusURLPath, api.NewClusterFullStatusH(snapshooter).Handle).Methods(http.MethodPost)
	r.HandleFunc(api.ClusterHistoryURLPath, api.NewClusterHistoryH().Handle).Methods(http.MethodPost)
	r.HandleFunc(api.ClusterListURLPath, api.NewClusterListHandler(clusterList).Handle).Methods(http.MethodGet)

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

func rootHandle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
