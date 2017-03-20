//Contains all cluster/ api endpoints handler functions
package api

import (
	"fmt"
	"github.com/golang/glog"
	"net/http"
)

const ClusterFullStatusUrlPath = "/cluster/full-status"

type ClusterFullStatusHandler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type ClusterFullStatusH struct{}

func NewClusterFullStatusH() *ClusterFullStatusH {
	return &ClusterFullStatusH{}
}

func (h ClusterFullStatusH) Handle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
	{

        “resources”: {
             “requests”: {
                  “cpu”: “12”,
                  “memory”: “120G”,
                  “percentOfAvailable”: “87%”,
             }
        },
        “nodes”: [
            {
                "name": "...",
                "status": "...",
                “resources”: {
                    “requests”: {
                        “cpu”: “2”,
                        “memory”: “12G”,
                        “percentOfAvailable”: “34%”,
                    }
                }
            },
            ...
        ]

}
`)
	glog.Info("returning cluster full status")
	glog.Flush()
}
