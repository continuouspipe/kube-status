//Package api - history contains the code that returns the list of the google cloud storage entries for the given interval
package api

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/continuouspipe/kube-status/history"
	"time"
	"github.com/continuouspipe/kube-status/errors"
	"encoding/json"
	"github.com/satori/go.uuid"
)

//ClusterHistoryURLPath is the api endpoint for the retrieving the cluster status historic data
const ClusterHistoryURLPath = "/clusters/{identifier}/history"
const ClusterHistoryEntryURLPath = "/clusters/{clusterIdentifier}/history/{entryUuid}"

//ClusterHistoryH handles the ClusterHistoryURLPath api endpoint
type ClusterHistoryH struct{
	ClusterStatusHistory history.ClusterStatusHistory
}

//NewClusterHistoryH returns an instance of ClusterHistoryH
func NewClusterHistoryH(history history.ClusterStatusHistory) *ClusterHistoryH {
	return &ClusterHistoryH{
		ClusterStatusHistory: history,
	}
}

//Handle is the handler for the ClusterHistoryURLPath api endpoint
//accepts query parameters ?cluster=&from=&to= using basic auth
func (h ClusterHistoryH) HandleList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	entries, err := h.ClusterStatusHistory.EntriesByCluster(
		vars["clusterIdentifier"],
		ParseDate(r, "from", time.Now().AddDate(0, 0, -1)),
		ParseDate(r, "to", time.Now()),
	)

	if err != nil {
		errors.LogAndRespondWithError(w, http.StatusInternalServerError, "Cannot find history: %s", err.Error())
		return
	}

	json, err := json.Marshal(entries)
	if err != nil {
		errors.LogAndRespondWithError(w, http.StatusInternalServerError, "error when marshalling the response body json: %s ", err.Error())
		return
	}

	w.Write(json)
}

func (h ClusterHistoryH) HandleEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

 	uuid, err := uuid.FromString(vars["entryUuid"])
	if err != nil {
		errors.LogAndRespondWithError(w, http.StatusInternalServerError, "Entry UUID seams invalid: %s", err.Error())
		return
	}

	entry, err := h.ClusterStatusHistory.Fetch(uuid)
	if err != nil {
		errors.LogAndRespondWithError(w, http.StatusInternalServerError, "error when fetching the entry: %s ", err.Error())
		return
	}

	json, err := json.Marshal(entry)
	if err != nil {
		errors.LogAndRespondWithError(w, http.StatusInternalServerError, "error when marshalling the response body json: %s ", err.Error())
		return
	}

	w.Write(json)
}

func ParseDate(request *http.Request, field string, defaultTime time.Time) time.Time {
	dateString := request.FormValue("from")
	if "" == dateString {
		return defaultTime
	}

	time, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		return defaultTime
	}

	return time
}
