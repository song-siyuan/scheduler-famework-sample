package sample

import (
	"context"

	v1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// Name is plugin name
	Name = "sample"
)

var _ framework.FilterPlugin = &Sample{}
var _ framework.PreBindPlugin = &Sample{}

type Sample struct {
	handle framework.Handle
}

func New(_ runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &Sample{
		handle: handle,
	}, nil
}

func (s *Sample) Name() string {
	return Name
}

type DominantResourceMap struct {
	mmap map[string]map[string]int
}

func (m *DominantResourceMap) Clone() framework.StateData {
	c := &DominantResourceMap{
		mmap: m.mmap,
	}
	return c
}

// podResource:map[cpu:1 memory:1048576000 myway5.com/device1:1 myway5.com/device2:1...]

func (s *Sample) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, node *framework.NodeInfo) *framework.Status {
	nodeName := node.Node().Name
	klog.V(2).Infof("nodeName: %v", nodeName)
	podReource := GetPodResource(pod)
	nodeResource := s.GetNodResource(nodeName)
	ResourceMap := DominantResourceMap{}
	ResourceMap.mmap = make(map[string]map[string]int)
	ResourceMap.mmap["podReource"] = podReource
	ResourceMap.mmap["nodeReource"] = nodeResource
	state.Write("ResourceMap", &ResourceMap)
	return framework.NewStatus(framework.Success, "")
}

func (s *Sample) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	podReource, err := state.Read("podReource")
	if err != nil {
		return 0, framework.NewStatus(framework.Error, err.Error())
	}
	nodeResource, err := state.Read("podReource")
	if err != nil {
		return 0, framework.NewStatus(framework.Error, err.Error())
	}

	/*
		计算集群整体的负载均衡得分
		1.计算放置后整个集群各个指标的负载均值
		2.计算节点的各个资源得分
		3.计算节点综合得分
		4.计算负载均衡度
	*/
	klog.V(2).Infof("podReource: %v", podReource)
	klog.V(2).Infof("nodeResource: %v", nodeResource)
	/*
		计算单个节点的资源平衡度值
	*/

	/*
		计算成本
	*/
	return 0, framework.NewStatus(framework.Success, "")
}

func (s *Sample) PreBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	//nodeInfo, err := s.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	// if err != nil {
	// 	return framework.NewStatus(framework.Error, err.Error())
	// }
	//klog.V(2).Infof("prebind node info: %+v", nodeInfo.Node())
	return framework.NewStatus(framework.Success, "")
}

// map[cpu:1 memory:1048576000 myway5.com/device1:1 myway5.com/device2:1]
func GetPodResource(pod *v1.Pod) map[string]int {
	containerNum := len(pod.Spec.Containers)
	podResource := map[string]int{}
	//获取所有container总共申请的资源种类和数量
	for i := 0; i < containerNum; i++ {
		//pod.Spec.Containers[i].Resources.Requests

		for res := range pod.Spec.Containers[i].Resources.Requests {
			var resourceName string = string(res)
			var r resource.Quantity = pod.Spec.Containers[i].Resources.Requests[res]
			podResource[resourceName] = podResource[resourceName] + int(r.Value())
		}
		//return resourceName
	}
	return podResource
}
func (s *Sample) GetNodResource(nodeName string) map[string]int {
	NodeInfos, err := s.handle.ClientSet().CoreV1().Nodes().Get(context.TODO(), nodeName, metaV1.GetOptions{})
	if err != nil {
		return map[string]int{}
	}
	nodeResource := map[string]int{}
	for res := range NodeInfos.Status.Capacity {
		var resourceName string = string(res)
		var r resource.Quantity = NodeInfos.Status.Capacity[res]
		nodeResource[resourceName] = int(r.Value())
	}
	return nodeResource
}
