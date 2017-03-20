package main

import (
	"flag"
	"fmt"
	"github.com/continuouspipe/kube-status/api"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"os"
)

var envListenAddress, _ = os.LookupEnv("KUBE_STATUS_LISTEN_ADDRESS") //e.g.: https://localhost:80

func main() {
	//parse the flags before glog start using them
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

	r := mux.NewRouter()
	r.HandleFunc(api.ClusterFullStatusUrlPath, api.NewClusterFullStatusH().Handle).Methods(http.MethodPost)
	http.ListenAndServe(listenURL.Host, r)
	glog.Flush()
}
