//Package api - cluster contains the code that gets the cluster status from the google cloud bucket
package api

import (
	"net/http"
	"github.com/continuouspipe/kube-status/clustersprovider"
	"github.com/continuouspipe/kube-status/errors"
	"encoding/json"
	"github.com/continuouspipe/kube-status/datasnapshots"
	"io/ioutil"
	"github.com/gorilla/mux"
)

//ClusterFullStatusURLPath is the api endpoint for the retrieving the cluster status historic data
const BackwardCompatibleClusterFullStatusURLPath = "/cluster/full-status"
const ClusterListURLPath = "/clusters"
const ClusterFullStatusURLPath = "/clusters/{clusterIdentifier}/status"

//ClusterFullStatusH handles the ClusterFullStatusURLPath api endpoint
type ClusterApiHandler struct{
	Snapshooter datasnapshots.ClusterSnapshooter
	Provider clustersprovider.ClusterListProvider
}

//NewClusterFullStatusH is the ctor for ClusterFullStatusH
func NewClusterApiHandler(snapshooter datasnapshots.ClusterSnapshooter, clusterListProvider clustersprovider.ClusterListProvider) *ClusterApiHandler {
	return &ClusterApiHandler{
		Snapshooter: snapshooter,
		Provider: clusterListProvider,
	}
}

//Handle is the handler for the ClusterFullStatusURLPath api endpoint
func (h ClusterApiHandler) HandleBackwardCompatible(w http.ResponseWriter, r *http.Request) {
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

	h.ReturnClusterStatus(w, requestedCluster)
}

func (h ClusterApiHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	cluster, err := h.Provider.ByIdentifier(vars["clusterIdentifier"])
	if err != nil {
		errors.LogAndRespondWithError(w, http.StatusBadRequest, "%s", err.Error())
		return
	}

	h.ReturnClusterStatus(w, cluster)
}

func (h ClusterApiHandler) ReturnClusterStatus(w http.ResponseWriter, cluster clustersprovider.Cluster) {
	status, err := h.Snapshooter.FetchCluster(cluster)
	if err != nil {
		errors.LogAndRespondWithError(w, http.StatusBadRequest, "Unable to fetch cluster: %s", err.Error())
		return
	}

	jsonStr, err := json.Marshal(*status)
	if err != nil {
		errors.LogAndRespondWithError(w, http.StatusBadRequest, "Error while marshalling status: %s", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonStr)
}

func (h ClusterApiHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	clusters, err := h.Provider.Clusters()
	if err != nil {
		errors.LogAndRespondWithError(w, http.StatusInternalServerError, "Error when loading the clusters: %s", err.Error())
		return
	}

	obfuscatedClusters := make([]clustersprovider.Cluster, len(clusters))

	// Obfuscate the credentials
	for k, cluster := range clusters {
		obfuscatedClusters[k] = clustersprovider.Cluster{
			Identifier: cluster.Identifier,
			Address: cluster.Address,
			Username: "OBFUSCATED",
			Password: "OBFUSCATED",
		}
	}

	respBody, err := json.Marshal(obfuscatedClusters)
	if err != nil {
		errors.LogAndRespondWithError(w, http.StatusBadRequest, "error when marshalling the response body json %s, details %s ", respBody, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}
