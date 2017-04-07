//Package api - cluster contains the code that gets the cluster status from the google cloud bucket
package api

import (
	"net/http"
	"github.com/continuouspipe/kube-status/clustersprovider"
	"github.com/continuouspipe/kube-status/errors"
	"encoding/json"
)

//ClusterFullStatusURLPath is the api endpoint for the retrieving the cluster status historic data
const ClusterFullStatusURLPath = "/cluster/status/full"
const ClusterListURLPath = "/clusters"

//ClusterFullStatusH handles the ClusterFullStatusURLPath api endpoint
type ClusterFullStatusH struct{}

//NewClusterFullStatusH is the ctor for ClusterFullStatusH
func NewClusterFullStatusH() *ClusterFullStatusH {
	return &ClusterFullStatusH{}
}

//Handle is the handler for the ClusterFullStatusURLPath api endpoint
//TODO: takes a uuid from the url query and write in the response the data stored on the google cloud bucket
func (h ClusterFullStatusH) Handle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	//w.Write(resp)
}

type ClusterListHandler struct{
	Provider clustersprovider.ClusterListProvider
}

func NewClusterListHandler(clusterListProvider clustersprovider.ClusterListProvider) *ClusterListHandler {
	return &ClusterListHandler{
		Provider: clusterListProvider,
	}
}

func (h ClusterListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	respBody, err := json.Marshal(h.Provider.Clusters())
	if err != nil {
		errors.LogAndRespondWithError(w, http.StatusBadRequest, "error when marshalling the response body json %s, details %s ", respBody, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}
