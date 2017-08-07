package clustersprovider

import (
	"fmt"
	"os"
	"net/http"
	"github.com/golang/glog"
	"io/ioutil"
	"encoding/json"
	"strings"
)

type CPClusterList struct{
	ApiKey string
	HttpClient *http.Client
}

type ContinuousPipeTeam struct {
	Slug       string `json:"slug"`
	BucketUuid string `json:"bucket_uuid"`
}

type ContinuousPipeCluster struct {
	Identifier string `json:"identifier"`
	Address    string `json:"address"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

func NewCPClusterList() *CPClusterList {
	apiKey := os.Getenv("CONTINUOUS_PIPE_API_KEY")
	if "" == apiKey {
		panic(fmt.Errorf("API key in environment variable %s required", "CONTINUOUS_PIPE_API_KEY"))
	}

	return &CPClusterList{
		ApiKey: apiKey,
		HttpClient: &http.Client{},
	}
}

func (c CPClusterList) Clusters() ([]Cluster, error) {
	responseBody, err := c.request("GET", "https://authenticator.continuouspipe.io/api/teams")
	if err != nil {
		return []Cluster{}, err
	}

	teams := make([]ContinuousPipeTeam, 0)
	err = json.Unmarshal(responseBody, &teams)
	if err != nil {
		return []Cluster{}, err
	}

	clusters := make([]Cluster, 0)
	for _, team := range teams {
		teamClusters, err := c.getTeamCluster(team)

		if err != nil {
			return clusters, err
		}

		clusters = append(clusters, teamClusters...)
	}

	return clusters, nil
}

func (c CPClusterList) getTeamCluster(team ContinuousPipeTeam) ([]Cluster, error) {
	responseBody, err := c.request("GET", "https://authenticator.continuouspipe.io/api/bucket/"+team.BucketUuid+"/clusters")
	if err != nil {
		return []Cluster{}, err
	}

	continuousPipeClusters := make([]ContinuousPipeCluster, 0)
	err = json.Unmarshal(responseBody, &continuousPipeClusters)
	if err != nil {
		return []Cluster{}, err
	}

	clusters := make([]Cluster, len(continuousPipeClusters))
	for key, cluster := range continuousPipeClusters {
		clusters[key] = Cluster{
			Identifier: team.Slug+"+"+cluster.Identifier,
			Address: cluster.Address,
			Username: cluster.Username,
			Password: cluster.Password,
		}
	}

	return clusters, nil
}

func (c CPClusterList) ByIdentifier(identifier string) (Cluster, error) {
	identifierParts := strings.Split(identifier, "+")

	if len(identifierParts) < 2 {
		return Cluster{}, fmt.Errorf("Cluster identifier do not have the expected syntax: %s", identifier)
	}

	teamName := identifierParts[0]
	clusterIdentifier := strings.Join(identifierParts[1:], "+")

	responseBody, err := c.request("GET", "https://authenticator.continuouspipe.io/api/teams/"+teamName)
	if err != nil {
		return Cluster{}, err
	}

	team := ContinuousPipeTeam{}
	err = json.Unmarshal(responseBody, &team)
	if err != nil {
		return Cluster{}, err
	}

	clusters, err := c.getTeamCluster(team)
	if err != nil {
		return Cluster{}, err
	}

	for _, cluster := range clusters {
		if cluster.Identifier == clusterIdentifier {
			return cluster, nil
		}
	}

	return Cluster{}, fmt.Errorf("Cluster %s was not found in team %s", clusterIdentifier, team.Slug)
}

func (c CPClusterList) request(method string, urlStr string) ([]byte, error) {
	request, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return []byte{}, err
	}

	request.Header.Add("X-Api-Key", c.ApiKey)

	return c.getResponseBody(request)
}

func (c CPClusterList) getResponseBody(req *http.Request) ([]byte, error) {
	res, err := c.HttpClient.Do(req)
	if err != nil {
		glog.V(4).Infoln("Error when creating client for request")
		return nil, err
	}
	if res.Body == nil {
		return nil, fmt.Errorf("Error requesting user information, response body empty, request status: %d", res.StatusCode)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error requesting user information, request status: %d", res.StatusCode)
	}
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return resBody, nil
}
