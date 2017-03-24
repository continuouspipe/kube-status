//Contains all cluster/ api endpoints handler functions
package api

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	kubernetesapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	clientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	"k8s.io/kubernetes/pkg/fields"
	"net/http"
	"strconv"
)

const ClusterFullStatusUrlPath = "/cluster/full-status"

type ClusterRequested struct {
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ClusterFullStatusHandler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type ClusterFullStatusResponse struct {
	Resources ClusterFullStatusResources        `json:"resources"`
	Nodes     []ClusterFullStatusNode           `json:"nodes"`
	Pods      map[string][]ClusterFullStatusPod `json:"pods"`
}

type ClusterFullStatusResources struct {
	Cpu                ClusterFullStatusRequestLimits          `json:"cpu"`
	Memory             ClusterFullStatusRequestLimits          `json:"memory"`
	PercentOfAvailable ClusterFullStatusRequestLimitsCpuMemory `json:"percentOfAvailable"`
}

type ClusterFullStatusRequestLimitsCpuMemory struct {
	Cpu    ClusterFullStatusRequestLimits `json:"cpu"`
	Memory ClusterFullStatusRequestLimits `json:"memory"`
}

type ClusterFullStatusRequestLimits struct {
	Request string `json:"requests"`
	Limits  string `json:"limits"`
}

type ClusterFullStatusCpuMemory struct {
	Cpu    string `json:"cpu"`
	Memory string `json:"memory"`
}

type ClusterFullStatusNode struct {
	Name              string                     `json:"name"`
	CreationTimestamp string                     `json:"creationTimestamp"`
	Status            string                     `json:"status"`
	Resources         ClusterFullStatusResources `json:"resources"`
	Capacity          ClusterFullStatusCpuMemory `json:"capacity"`
	Allocatable       ClusterFullStatusCpuMemory `json:"allocatable"`
	VolumesInUse      int                        `json:"volumesInUse"`
}

type ClusterFullStatusPod struct {
	Name              string                       `json:"name"`
	Status            string                       `json:"status"`
	CreationTimestamp string                       `json:"creationTimestamp"`
	IsReady           bool                         `json:"isReady"`
	Containers        []ClusterFullStatusContainer `json:"containers"`
	Events            []kubernetesapi.Event        `json:"events"`
}

type ClusterFullStatusContainer struct {
	Name         string                                  `json:"name"`
	State        string                                  `json:"state"`
	IsReady      bool                                    `json:"isReady"`
	RestartCount int32                                   `json:"restartCount"`
	Resources    ClusterFullStatusRequestLimitsCpuMemory `json:"resources"`
}

type ClusterFullStatusH struct{}

func NewClusterFullStatusH() *ClusterFullStatusH {
	return &ClusterFullStatusH{}
}

func (h ClusterFullStatusH) Handle(w http.ResponseWriter, r *http.Request) {
	resBodyData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logAndRespondWithError(w, http.StatusBadRequest, "error when reading the request body %s, details %s ", r.Body, err.Error())
		return
	}

	requestedCluster := ClusterRequested{}
	err = json.Unmarshal(resBodyData, &requestedCluster)
	if err != nil {
		logAndRespondWithError(w, http.StatusBadRequest, "error when unmarshalling the request body json %s, details %s ", r.Body, err.Error())
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
		logAndRespondWithError(w, http.StatusInternalServerError, "error when creating the rest config, details %s ", err.Error())
		return
	}

	clientset, err := internalclientset.NewForConfig(restConfig)
	if err != nil {
		logAndRespondWithError(w, http.StatusInternalServerError, "error when creating the client api, details %s ", err.Error())
		return
	}

	nodes, err := clientset.Core().Nodes().List(kubernetesapi.ListOptions{})
	if err != nil {
		logAndRespondWithError(w, http.StatusInternalServerError, "error when getting the node list, details %s ", err.Error())
		return
	}

	//TODO: do not pass the html writer to this sub-function but let them return an error
	podLists := getPodListByNode(w, clientset, nodes)
	statusCluster := getStatusCluster()
	statusNodes := getStatusNodes(w, podLists, nodes)
	podsEvents := getPodsEvents(clientset, podLists)
	statusPods, err := getStatusPods(podLists, podsEvents)
	if err != nil {
		logAndRespondWithError(w, http.StatusInternalServerError, "error when getting the node list, details %s ", err.Error())
		return
	}

	//Build the full status response
	statusResponse := &ClusterFullStatusResponse{
		statusCluster,
		statusNodes,
		*statusPods,
	}

	respBody, err := json.Marshal(statusResponse)
	if err != nil {
		logAndRespondWithError(w, http.StatusBadRequest, "error when marshalling the response body json %s, details %s ", respBody, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}

func getStatusCluster() ClusterFullStatusResources {
	return ClusterFullStatusResources{
		ClusterFullStatusRequestLimits{},
		ClusterFullStatusRequestLimits{},
		ClusterFullStatusRequestLimitsCpuMemory{
			ClusterFullStatusRequestLimits{},
			ClusterFullStatusRequestLimits{},
		},
	}
}

func getPodListByNode(w http.ResponseWriter, clientset *internalclientset.Clientset, nodes *kubernetesapi.NodeList) map[string]*kubernetesapi.PodList {
	podLists := make(map[string]*kubernetesapi.PodList)

	//get the pod list
	for _, node := range nodes.Items {
		fieldSelector, err := fields.ParseSelector("spec.nodeName=" + node.GetName() + ",status.phase!=" + string(kubernetesapi.PodSucceeded) + ",status.phase!=" + string(kubernetesapi.PodFailed))
		if err != nil {
			logAndRespondWithError(w, http.StatusInternalServerError, "error when parsing the list option fields, details %s ", err.Error())
			return nil
		}

		nodeNonTerminatedPodsList, err := clientset.Core().Pods("").List(kubernetesapi.ListOptions{FieldSelector: fieldSelector})
		if err != nil {
			logAndRespondWithError(w, http.StatusInternalServerError, "error when fetching the list of pods for the namespace %s, details %s ", node.GetNamespace(), err.Error())
			continue
		}
		podLists[node.GetName()] = nodeNonTerminatedPodsList
	}

	return podLists
}

func getStatusNodes(w http.ResponseWriter, podLists map[string]*kubernetesapi.PodList, nodes *kubernetesapi.NodeList) []ClusterFullStatusNode {
	statusNodes := []ClusterFullStatusNode{}
	for _, node := range nodes.Items {
		totalConditions := len(node.Status.Conditions)
		if totalConditions <= 0 {
			continue
		}

		nodeResources, err := getNodeResource(podLists[node.GetName()], &node)
		if err != nil {
			logAndRespondWithError(w, http.StatusInternalServerError, "error when getting the node resources for namespace %s, details %s ", node.GetNamespace(), err.Error())
			continue
		}

		statusNodes = append(statusNodes, ClusterFullStatusNode{
			node.Name,
			node.GetCreationTimestamp().String(),
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
				ClusterFullStatusRequestLimitsCpuMemory{
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
			ClusterFullStatusCpuMemory{
				node.Status.Capacity.Cpu().String(),
				node.Status.Capacity.Memory().String(),
			},
			ClusterFullStatusCpuMemory{
				node.Status.Allocatable.Cpu().String(),
				node.Status.Allocatable.Memory().String(),
			},
			len(node.Status.VolumesInUse),
		})
	}
	return statusNodes
}

type podEventWrapper struct {
	name   string
	err    error
	events []kubernetesapi.Event
}

func getPodsEvents(clientset *internalclientset.Clientset, podLists map[string]*kubernetesapi.PodList) map[string][]kubernetesapi.Event {
	eventItemsRequestedCount := 0
	eventsChan := make(chan podEventWrapper)
	defer close(eventsChan)

	for _, podList := range podLists {
		for _, pod := range podList.Items {
			isPodReady := kubernetesapi.IsPodReady(&pod)
			//only fetch the events for the pods that are not ready
			if isPodReady == false {
				eventItemsRequestedCount = eventItemsRequestedCount + 1
				go getEvents(clientset, pod, eventsChan)
			}
		}
	}
	podEventsMap := make(map[string][]kubernetesapi.Event)
	var podEvent podEventWrapper
	for i := 0; i < eventItemsRequestedCount; i++ {
		podEvent = <-eventsChan
		if podEvent.err != nil {
			glog.V(4).Infof("it was not possible to fetch the events for the pod %s, error %s", podEvent.name, podEvent.err.Error())
			glog.Flush()
			continue
		}
		podEventsMap[podEvent.name] = podEvent.events
	}
	return podEventsMap
}

func getEvents(clientset *internalclientset.Clientset, pod kubernetesapi.Pod, eventsChan chan<- podEventWrapper) {
	ref, err := kubernetesapi.GetReference(&pod)
	if err != nil {
		eventsChan <- podEventWrapper{err: fmt.Errorf("Unable to construct reference to '%#v': %v", pod, err)}

	}

	ref.Kind = ""
	e, err := clientset.Core().Events(pod.GetNamespace()).Search(ref)
	if err != nil {
		eventsChan <- podEventWrapper{err: fmt.Errorf("Unable to get events for pod %s", pod.GetName())}
	}

	eventsChan <- podEventWrapper{
		name:   pod.GetName(),
		events: e.Items,
	}
}

func getStatusPods(podLists map[string]*kubernetesapi.PodList, podsEvents map[string][]kubernetesapi.Event) (*map[string][]ClusterFullStatusPod, error) {
	statusPods := make(map[string][]ClusterFullStatusPod)
	for _, podList := range podLists {
		for _, pod := range podList.Items {
			statuses := map[string]kubernetesapi.ContainerStatus{}
			for _, status := range pod.Status.ContainerStatuses {
				statuses[status.Name] = status
			}

			statusContainers := []ClusterFullStatusContainer{}
			for _, container := range pod.Spec.Containers {

				status, ok := statuses[container.Name]

				containerStatus := ClusterFullStatusContainer{}
				containerStatus.Name = container.Name
				if ok {
					containerStatus.State = describeContainerState(status.State)
					containerStatus.IsReady = status.Ready
					containerStatus.RestartCount = status.RestartCount
				}

				containerStatus.Resources = ClusterFullStatusRequestLimitsCpuMemory{
					Cpu: ClusterFullStatusRequestLimits{
						Request: container.Resources.Requests.Cpu().String(),
						Limits:  container.Resources.Limits.Cpu().String(),
					},
					Memory: ClusterFullStatusRequestLimits{
						Request: container.Resources.Requests.Memory().String(),
						Limits:  container.Resources.Limits.Memory().String(),
					},
				}

				statusContainers = append(statusContainers, containerStatus)
			}
			isPodReady := kubernetesapi.IsPodReady(&pod)

			statusPods[pod.GetNamespace()] = append(statusPods[pod.GetNamespace()], ClusterFullStatusPod{
				pod.GetName(),
				string(pod.Status.Phase),
				pod.GetCreationTimestamp().String(),
				isPodReady,
				statusContainers,
				podsEvents[pod.GetName()],
			})
		}
	}
	return &statusPods, nil
}

func describeContainerState(state kubernetesapi.ContainerState) string {
	switch {
	case state.Running != nil:
		return "Running"
	case state.Terminated != nil:
		return "Terminated"
	default:
		return "Waiting"
	}
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

func getNodeResource(nodeNonTerminatedPodsList *kubernetesapi.PodList, node *kubernetesapi.Node) (*NodeResource, error) {
	allocatable := node.Status.Capacity
	if len(node.Status.Allocatable) > 0 {
		allocatable = node.Status.Allocatable
	}

	reqs, limits, err := getPodsTotalRequestsAndLimits(nodeNonTerminatedPodsList)
	if err != nil {
		return nil, err
	}
	cpuReqs, cpuLimits, memoryReqs, memoryLimits := reqs[kubernetesapi.ResourceCPU], limits[kubernetesapi.ResourceCPU], reqs[kubernetesapi.ResourceMemory], limits[kubernetesapi.ResourceMemory]
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

func getPodsTotalRequestsAndLimits(podList *kubernetesapi.PodList) (reqs map[kubernetesapi.ResourceName]resource.Quantity, limits map[kubernetesapi.ResourceName]resource.Quantity, err error) {
	reqs, limits = map[kubernetesapi.ResourceName]resource.Quantity{}, map[kubernetesapi.ResourceName]resource.Quantity{}
	for _, pod := range podList.Items {
		podReqs, podLimits, err := kubernetesapi.PodRequestsAndLimits(&pod)
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
