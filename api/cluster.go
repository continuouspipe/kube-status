//Contains all cluster/ api endpoints handler functions
package api

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const ClusterFullStatusUrlPath = "/cluster/full-status"

type clusterRequested struct {
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ClusterFullStatusHandler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type ClusterFullStatusResponse struct {
	Resources ClusterFullStatusResources `json:"resources"`
	Nodes     []ClusterFullStatusNode    `json:"nodes"`
}

type ClusterFullStatusResources struct {
}

type ClusterFullStatusRequests struct {
	Cpu                int    `json:"int"`
	Memory             string `json:"memory"`
	PercentOfAvailable string `json:"percentOfAvailable"`
}

type ClusterFullStatusNode struct {
	Name      string                     `json:"name"`
	Status    string                     `json:"status"`
	Resources ClusterFullStatusResources `json:"resources"`
}

type ClusterFullStatusH struct{}

func NewClusterFullStatusH() *ClusterFullStatusH {
	return &ClusterFullStatusH{}
}

/*
`
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
`
*/

func (h ClusterFullStatusH) Handle(w http.ResponseWriter, r *http.Request) {
	resBodyData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := fmt.Errorf("error when reading the request body %s, details %s ", r.Body, err.Error())
		glog.Error(err.Error())
		glog.Flush()
		respondWithError(w, err, http.StatusBadRequest)
		return
	}

	requestedCluster := clusterRequested{}
	err = json.Unmarshal(resBodyData, &requestedCluster)
	if err != nil {
		err = fmt.Errorf("error when unmarshalling the request body json %s, details %s ", r.Body, err.Error())
		glog.Error(err)
		glog.Flush()
		respondWithError(w, err, http.StatusBadRequest)
		return
	}

	ctx := clientcmdapi.NewContext()
	cfg := clientcmdapi.NewConfig()
	authInfo := clientcmdapi.NewAuthInfo()

	authInfo.Username = requestedCluster.Username
	authInfo.Password = requestedCluster.Password

	cluster := clientcmdapi.NewCluster()
	cluster.Server = requestedCluster.Address
	cluster.InsecureSkipTLSVerify = true

	cfg.Contexts = map[string]*clientcmdapi.Context{"default": ctx}
	cfg.CurrentContext = "default"
	overrides := clientcmd.ConfigOverrides{
		ClusterInfo: *cluster,
		AuthInfo:    *authInfo,
	}

	clientConfig := clientcmd.NewNonInteractiveClientConfig(*cfg, "default", &overrides, nil)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		err = fmt.Errorf("error when creating the rest config, details %s ", err.Error())
		glog.Error(err)
		glog.Flush()
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		err = fmt.Errorf("error when creating the client api, details %s ", err.Error())
		glog.Error(err)
		glog.Flush()
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}

	nodes, err := clientset.CoreV1().Nodes().List(v1.ListOptions{})
	if err != nil {
		err = fmt.Errorf("error when getting the node list, details %s ", err.Error())
		glog.Error(err)
		glog.Flush()
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}

	statusNodes := []ClusterFullStatusNode{}

	for _, node := range nodes.Items {
		totalConditions := len(node.Status.Conditions)
		statusNodes = append(statusNodes, ClusterFullStatusNode{
			node.Name,
			string(node.Status.Conditions[totalConditions-1].Type),
			ClusterFullStatusResources{},
		})
	}

	statusResponse := &ClusterFullStatusResponse{
		ClusterFullStatusResources{},
		statusNodes,
	}

	respBody, err := json.Marshal(statusResponse)
	if err != nil {
		err = fmt.Errorf("error when marshalling the response body json %s, details %s ", respBody, err.Error())
		glog.Error(err)
		glog.Flush()
		respondWithError(w, err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}
