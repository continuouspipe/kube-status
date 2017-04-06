//Package datasnapshots - gets the full cluster status
package datasnapshots

import (
	"encoding/json"
	"fmt"
	"github.com/continuouspipe/kube-status/clustersprovider"
	"github.com/continuouspipe/kube-status/errors"
	kubernetesapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	clientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	"k8s.io/kubernetes/pkg/fields"
	"strconv"
)

//ClusterFullStatusResponse top level of the individual cluster status response
type ClusterFullStatusResponse struct {
	Resources ClusterFullStatusResources        `json:"resources"`
	Nodes     []ClusterFullStatusNode           `json:"nodes"`
	Pods      map[string][]ClusterFullStatusPod `json:"pods"`
}

//ClusterFullStatusResources resource status information for each cluster node and for the children nodes
type ClusterFullStatusResources struct {
	CPU                ClusterFullStatusRequestLimits          `json:"cpu"`
	Memory             ClusterFullStatusRequestLimits          `json:"memory"`
	PercentOfAvailable ClusterFullStatusRequestLimitsCPUMemory `json:"percentOfAvailable"`
}

//ClusterFullStatusRequestLimitsCPUMemory limits for cpu and memory
type ClusterFullStatusRequestLimitsCPUMemory struct {
	CPU    ClusterFullStatusRequestLimits `json:"cpu"`
	Memory ClusterFullStatusRequestLimits `json:"memory"`
}

//ClusterFullStatusRequestLimits requests and limits
type ClusterFullStatusRequestLimits struct {
	Request string `json:"requests"`
	Limits  string `json:"limits"`
}

//ClusterFullStatusCPUMemory cpu and memory
type ClusterFullStatusCPUMemory struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

//ClusterFullStatusNode status for each children node
type ClusterFullStatusNode struct {
	Name              string                     `json:"name"`
	CreationTimestamp string                     `json:"creationTimestamp"`
	Status            string                     `json:"status"`
	Resources         ClusterFullStatusResources `json:"resources"`
	Capacity          ClusterFullStatusCPUMemory `json:"capacity"`
	Allocatable       ClusterFullStatusCPUMemory `json:"allocatable"`
	VolumesInUse      int                        `json:"volumesInUse"`
}

//ClusterFullStatusPod status for each pod
type ClusterFullStatusPod struct {
	Name              string                       `json:"name"`
	Status            string                       `json:"status"`
	CreationTimestamp string                       `json:"creationTimestamp"`
	IsReady           bool                         `json:"isReady"`
	Containers        []ClusterFullStatusContainer `json:"containers"`
	Events            []kubernetesapi.Event        `json:"events"`
}

//ClusterFullStatusContainer status for each container
type ClusterFullStatusContainer struct {
	Name         string                                  `json:"name"`
	State        string                                  `json:"state"`
	IsReady      bool                                    `json:"isReady"`
	RestartCount int32                                   `json:"restartCount"`
	Resources    ClusterFullStatusRequestLimitsCPUMemory `json:"resources"`
}

//ClusterSnapshooter takes the full status of one or more clusters and returns it as a json formatted string
type ClusterSnapshooter interface {
	Add(cluster clustersprovider.Cluster)
	Fetch() ([]byte, error)
}

//ClusterSnapshot takes the full status of one or more clusters and returns it as a json formatted string
type ClusterSnapshot struct {
	clusters        []clustersprovider.Cluster
	clusterStatuses []ClusterFullStatusResponse
}

//NewClusterSnapshot is the ctor for ClusterSnapshot
func NewClusterSnapshot() *ClusterSnapshot {
	return &ClusterSnapshot{}
}

//Add appends a new clustersprovider.Cluster struct to the stored list
func (s *ClusterSnapshot) Add(cluster clustersprovider.Cluster) {
	s.clusters = append(s.clusters, cluster)
}

//Fetch retrieves the cluster statuses for all the clusters
func (s *ClusterSnapshot) Fetch() ([]byte, error) {
	for _, requestedCluster := range s.clusters {
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
			return []byte{}, fmt.Errorf("error when creating the rest config, details %s ", err.Error())
		}

		clientset, err := internalclientset.NewForConfig(restConfig)
		if err != nil {
			return []byte{}, fmt.Errorf("error when creating the client api, details %s ", err.Error())
		}

		nodes, err := clientset.Core().Nodes().List(kubernetesapi.ListOptions{})
		if err != nil {
			return []byte{}, fmt.Errorf("error when getting the node list, details %s ", err.Error())
		}

		podLists, errList := getPodListByNode(clientset, nodes)
		if len(errList.Items()) > 0 {
			return []byte{}, errList
		}

		statusCluster := getStatusCluster()
		statusNodes, errList := getStatusNodes(podLists, nodes)
		if len(errList.Items()) > 0 {
			return []byte{}, errList
		}

		podsEvents, errList := getPodsEvents(clientset, podLists)
		if len(errList.Items()) > 0 {
			return []byte{}, errList
		}

		statusPods, err := getStatusPods(podLists, podsEvents)
		if err != nil {
			return []byte{}, fmt.Errorf("error when getting the node list, details %s ", err.Error())
		}

		//Build the full status response
		s.clusterStatuses = append(s.clusterStatuses, ClusterFullStatusResponse{
			statusCluster,
			statusNodes,
			*statusPods,
		})
	}
	statuses, err := json.Marshal(s.clusterStatuses)
	if err != nil {
		return []byte{}, fmt.Errorf("error when marshalling %#v, details %s ", s.clusterStatuses, err.Error())
	}

	return statuses, nil
}

func getStatusCluster() ClusterFullStatusResources {
	return ClusterFullStatusResources{
		ClusterFullStatusRequestLimits{},
		ClusterFullStatusRequestLimits{},
		ClusterFullStatusRequestLimitsCPUMemory{
			ClusterFullStatusRequestLimits{},
			ClusterFullStatusRequestLimits{},
		},
	}
}

func getPodListByNode(clientset *internalclientset.Clientset, nodes *kubernetesapi.NodeList) (*map[string]*kubernetesapi.PodList, *errors.ErrorList) {
	podLists := make(map[string]*kubernetesapi.PodList)
	el := errors.ErrorList{}

	//get the pod list
	for _, node := range nodes.Items {
		fieldSelector, err := fields.ParseSelector("spec.nodeName=" + node.GetName() + ",status.phase!=" + string(kubernetesapi.PodSucceeded) + ",status.phase!=" + string(kubernetesapi.PodFailed))
		if err != nil {
			el.Add(err)
			el.Add(fmt.Errorf("error when parsing the list option fields"))
			return nil, &el
		}

		nodeNonTerminatedPodsList, err := clientset.Core().Pods("").List(kubernetesapi.ListOptions{FieldSelector: fieldSelector})
		if err != nil {
			el.Add(err)
			el.Add(fmt.Errorf("error when fetching the list of pods for the namespace %s, details %s ", node.GetNamespace(), err.Error()))
			continue
		}
		podLists[node.GetName()] = nodeNonTerminatedPodsList
	}

	return &podLists, &el
}

func getStatusNodes(podLists *map[string]*kubernetesapi.PodList, nodes *kubernetesapi.NodeList) ([]ClusterFullStatusNode, *errors.ErrorList) {
	statusNodes := []ClusterFullStatusNode{}
	el := errors.ErrorList{}

	for _, node := range nodes.Items {
		totalConditions := len(node.Status.Conditions)
		if totalConditions <= 0 {
			continue
		}

		nodeResources, err := getNodeResource((*podLists)[node.GetName()], &node)
		if err != nil {
			el.Add(err)
			el.Add(fmt.Errorf("error when getting the node resources for namespace %s", node.GetNamespace()))
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
				ClusterFullStatusRequestLimitsCPUMemory{
					ClusterFullStatusRequestLimits{
						strconv.FormatInt(nodeResources.fractionCPUReqs, 10),
						strconv.FormatInt(nodeResources.fractionCPULimits, 10),
					},
					ClusterFullStatusRequestLimits{
						strconv.FormatInt(nodeResources.fractionMemoryReqs, 10),
						strconv.FormatInt(nodeResources.fractionMemoryLimits, 10),
					},
				},
			},
			ClusterFullStatusCPUMemory{
				node.Status.Capacity.Cpu().String(),
				node.Status.Capacity.Memory().String(),
			},
			ClusterFullStatusCPUMemory{
				node.Status.Allocatable.Cpu().String(),
				node.Status.Allocatable.Memory().String(),
			},
			len(node.Status.VolumesInUse),
		})
	}
	return statusNodes, &el
}

type podEventWrapper struct {
	name   string
	err    error
	events []kubernetesapi.Event
}

func getPodsEvents(clientset *internalclientset.Clientset, podLists *map[string]*kubernetesapi.PodList) (map[string][]kubernetesapi.Event, *errors.ErrorList) {
	el := &errors.ErrorList{}
	eventItemsRequestedCount := 0
	eventsChan := make(chan podEventWrapper)
	defer close(eventsChan)

	for _, podList := range *podLists {
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
			el.Add(podEvent.err)
			el.Add(fmt.Errorf("it was not possible to fetch the events for the pod %s", podEvent.name))
			continue
		}
		podEventsMap[podEvent.name] = podEvent.events
	}
	return podEventsMap, el
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

func getStatusPods(podLists *map[string]*kubernetesapi.PodList, podsEvents map[string][]kubernetesapi.Event) (*map[string][]ClusterFullStatusPod, error) {
	statusPods := make(map[string][]ClusterFullStatusPod)
	for _, podList := range *podLists {
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

				containerStatus.Resources = ClusterFullStatusRequestLimitsCPUMemory{
					CPU: ClusterFullStatusRequestLimits{
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

//NodeResource contains the resources for each node
type NodeResource struct {
	cpuReqs              string
	fractionCPUReqs      int64
	cpuLimits            string
	fractionCPULimits    int64
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
	fractionCPUReqs := float64(cpuReqs.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
	fractionCPULimits := float64(cpuLimits.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
	fractionMemoryReqs := float64(memoryReqs.Value()) / float64(allocatable.Memory().Value()) * 100
	fractionMemoryLimits := float64(memoryLimits.Value()) / float64(allocatable.Memory().Value()) * 100

	return &NodeResource{
		cpuReqs.String(),
		int64(fractionCPUReqs),
		cpuLimits.String(),
		int64(fractionCPULimits),
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
