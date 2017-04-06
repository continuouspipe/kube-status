//Package api - history contains the code that returns the list of the google cloud storage entries for the given interval
package api

import "net/http"

//ClusterHistoryURLPath is the api endpoint for the retrieving the cluster status historic data
const ClusterHistoryURLPath = "/history"

//ClusterHistoryH handles the ClusterHistoryURLPath api endpoint
type ClusterHistoryH struct{}

//NewClusterHistoryH returns an instance of ClusterHistoryH
func NewClusterHistoryH() *ClusterHistoryH {
	return &ClusterHistoryH{}
}

//Handle is the handler for the ClusterHistoryURLPath api endpoint
//accepts query parameters ?cluster=&from=&to= using basic auth
func (h ClusterHistoryH) Handle(w http.ResponseWriter, r *http.Request) {

}

// ClusterHistoryWriter requests the kubernetes clusters status every X minutes and stores it in a Google Cloud Storage bucket
type ClusterHistoryWriter struct {
}
