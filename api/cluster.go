//Contains all cluster/ api endpoints handler functions
package api

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	clientapi "k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"net/http"
	"strconv"
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
	Cpu                ClusterFullStatusRequestLimits      `json:"cpu"`
	Memory             ClusterFullStatusRequestLimits      `json:"memory"`
	PercentOfAvailable ClusterFullStatusPercentOfAvailable `json:"percentOfAvailable"`
}

type ClusterFullStatusPercentOfAvailable struct {
	Cpu    ClusterFullStatusRequestLimits `json:"cpu"`
	Memory ClusterFullStatusRequestLimits `json:"memory"`
}

type ClusterFullStatusRequestLimits struct {
	Request string `json:"requests"`
	Limits  string `json:"limits"`
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
		if totalConditions <= 0 {
			continue
		}

		podList, err := clientset.CoreV1().Pods(node.GetNamespace()).List(v1.ListOptions{})
		if err != nil {
			err = fmt.Errorf("error when fetching the list of pods for the namespace %s, details %s ", node.GetNamespace(), err.Error())
			glog.Error(err)
			glog.Flush()
			continue
		}

		nodeResources, err := getNodeResource(podList, &node)

		statusNodes = append(statusNodes, ClusterFullStatusNode{
			node.Name,
			string(node.Status.Conditions[totalConditions-1].Type),
			ClusterFullStatusResources{
				ClusterFullStatusRequestLimits{
					nodeResources.cpuReqs,
					nodeResources.cpuLimits,
				},
				ClusterFullStatusRequestLimits{
					nodeResources.memoryReqs,
					nodeResources.memoryLimits,
				},
				ClusterFullStatusPercentOfAvailable{
					ClusterFullStatusRequestLimits{
						strconv.FormatInt(nodeResources.fractionCpuReqs, 10),
						strconv.FormatInt(nodeResources.fractionCpuLimits, 10),
					},
					ClusterFullStatusRequestLimits{
						strconv.FormatInt(nodeResources.fractionMemoryReqs, 10),
						strconv.FormatInt(nodeResources.fractionMemoryLimits, 10),
					},
				},
			},
		})
	}

	statusResponse := &ClusterFullStatusResponse{
		ClusterFullStatusResources{
			ClusterFullStatusRequestLimits{},
			ClusterFullStatusRequestLimits{},
			ClusterFullStatusPercentOfAvailable{
				ClusterFullStatusRequestLimits{},
				ClusterFullStatusRequestLimits{},
			},
		},
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

type NodeResource struct {
	cpuReqs              string
	fractionCpuReqs      int64
	cpuLimits            string
	fractionCpuLimits    int64
	memoryReqs           string
	fractionMemoryReqs   int64
	memoryLimits         string
	fractionMemoryLimits int64
}

func getNodeResource(nodeNonTerminatedPodsList *clientapi.PodList, node *clientapi.Node) (*NodeResource, error) {
	allocatable := node.Status.Capacity
	if len(node.Status.Allocatable) > 0 {
		allocatable = node.Status.Allocatable
	}

	reqs, limits, err := getPodsTotalRequestsAndLimits(nodeNonTerminatedPodsList)
	if err != nil {
		return nil, err
	}
	cpuReqs, cpuLimits, memoryReqs, memoryLimits := reqs[clientapi.ResourceCPU], limits[clientapi.ResourceCPU], reqs[clientapi.ResourceMemory], limits[clientapi.ResourceMemory]
	fractionCpuReqs := float64(cpuReqs.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
	fractionCpuLimits := float64(cpuLimits.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
	fractionMemoryReqs := float64(memoryReqs.Value()) / float64(allocatable.Memory().Value()) * 100
	fractionMemoryLimits := float64(memoryLimits.Value()) / float64(allocatable.Memory().Value()) * 100

	return &NodeResource{
		cpuReqs.String(),
		int64(fractionCpuReqs),
		cpuLimits.String(),
		int64(fractionCpuLimits),
		memoryReqs.String(),
		int64(fractionMemoryReqs),
		memoryLimits.String(),
		int64(fractionMemoryLimits)}, nil
}

func getPodsTotalRequestsAndLimits(podList *clientapi.PodList) (reqs map[clientapi.ResourceName]resource.Quantity, limits map[clientapi.ResourceName]resource.Quantity, err error) {
	reqs, limits = map[clientapi.ResourceName]resource.Quantity{}, map[clientapi.ResourceName]resource.Quantity{}
	for _, pod := range podList.Items {
		podReqs, podLimits, err := clientapi.PodRequestsAndLimits(&pod)
		if err != nil {
			return nil, nil, err
		}
		for podReqName, podReqValue := range podReqs {
			if value, ok := reqs[podReqName]; !ok {
				reqs[podReqName] = *podReqValue.Copy()
			} else {
				value.Add(podReqValue)
				reqs[podReqName] = value
			}
		}
		for podLimitName, podLimitValue := range podLimits {
			if value, ok := limits[podLimitName]; !ok {
				limits[podLimitName] = *podLimitValue.Copy()
			} else {
				value.Add(podLimitValue)
				limits[podLimitName] = value
			}
		}
	}
	return
}
