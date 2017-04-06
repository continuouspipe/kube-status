//Package api - cluster contains the code that gets the cluster status from the google cloud bucket
package api

import "net/http"

//ClusterFullStatusURLPath is the api endpoint for the retrieving the cluster status historic data
const ClusterFullStatusURLPath = "/cluster/status/full"

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
