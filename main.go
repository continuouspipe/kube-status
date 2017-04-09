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
)

var envListenAddress, _ = os.LookupEnv("KUBE_STATUS_LISTEN_ADDRESS") //e.g.: https://localhost:80

func main() {
	listProviderType := flag.String("cluster-list", "in-memory", "the cluster list provider")
	flag.Parse()

	glog.Infof("starting kube status api listening at address %s", envListenAddress)
	glog.Flush()

	listenURL, err := url.Parse(envListenAddress)
	if err != nil {
		glog.V(5).Infof("cannot parse URL: %v\n", err.Error())
		glog.Flush()
		fmt.Printf("Cannot parse URL: %v\n", err.Error())
		os.Exit(1)
	}

	var clusterList clustersprovider.ClusterListProvider
	if "in-memory" == *listProviderType {
		clusterList = clustersprovider.NewInMemoryClusterList()
	} else {
		clusterList = clustersprovider.NewCPClusterList()
	}

	snapshooter := datasnapshots.NewClusterSnapshot()
	snapshotHandler := datasnapshots.NewDataSnapshotHandler(
		clusterList,
		snapshooter,
	)
	go snapshotHandler.Handle()

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
