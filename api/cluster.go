//Package api - cluster contains the code that gets the cluster status from the google cloud bucket
package api

import (
	"net/http"
	"github.com/continuouspipe/kube-status/clustersprovider"
	"github.com/continuouspipe/kube-status/errors"
	"encoding/json"
	"github.com/continuouspipe/kube-status/datasnapshots"
	"io/ioutil"
)

//ClusterFullStatusURLPath is the api endpoint for the retrieving the cluster status historic data
const ClusterFullStatusURLPath = "/cluster/status/full"
const ClusterListURLPath = "/clusters"

//ClusterFullStatusH handles the ClusterFullStatusURLPath api endpoint
type ClusterFullStatusH struct{
	Snapshooter datasnapshots.ClusterSnapshooter
}

//NewClusterFullStatusH is the ctor for ClusterFullStatusH
func NewClusterFullStatusH(snapshooter datasnapshots.ClusterSnapshooter) *ClusterFullStatusH {
	return &ClusterFullStatusH{
		Snapshooter: snapshooter,
	}
}

//Handle is the handler for the ClusterFullStatusURLPath api endpoint
//TODO: takes a uuid from the url query and write in the response the data stored on the google cloud bucket
func (h ClusterFullStatusH) Handle(w http.ResponseWriter, r *http.Request) {
	resBodyData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errors.LogAndRespondWithError(w, http.StatusBadRequest, "error when reading the request body %s, details %s ", r.Body, err.Error())
		return
	}

	requestedCluster := clustersprovider.Cluster{}
	err = json.Unmarshal(resBodyData, &requestedCluster)
	if err != nil {
		errors.LogAndRespondWithError(w, http.StatusBadRequest, "error when unmarshalling the request body json %s, details %s ", r.Body, err.Error())
		return
	}


	w.WriteHeader(http.StatusOK)

	status, err := h.Snapshooter.FetchCluster(requestedCluster)
	if err != nil {
		errors.LogAndRespondWithError(w, http.StatusBadRequest, "error when unmarshalling the request body json %s, details %s ", r.Body, err.Error())
		return
	}

	json, err := json.Marshal(*status)
	if err != nil {
		errors.LogAndRespondWithError(w, http.StatusBadRequest, "error when unmarshalling the request body json %s, details %s ", r.Body, err.Error())
		return
	}

	w.Write(json)
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
