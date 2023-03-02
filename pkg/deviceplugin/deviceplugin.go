package deviceplugin

import (
	"strconv"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"

	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
)

const (
	mutualcpuResourceNamespace = "openshift.io"
	mutualcpuResourceName      = "mutualcpu"
)

type MutualCpu struct {
	cpus cpuset.CPUSet
}

func (mc *MutualCpu) GetResourceNamespace() string {
	return mutualcpuResourceNamespace
}

func (mc *MutualCpu) Discover(pnl chan dpm.PluginNameList) {
	pnl <- []string{mutualcpuResourceName}
	return
}

func (mc *MutualCpu) NewPlugin(s string) dpm.PluginInterface {
	return pluginImp{mutualCpus: &mc.cpus}
}

func New(cpus string) (*dpm.Manager, error) {
	mutualCpus, err := cpuset.Parse(cpus)
	if err != nil {
		return nil, err
	}
	mc := &MutualCpu{cpus: mutualCpus}
	return dpm.NewManager(mc), nil
}

func MakeMutualCpusDevices(cpus *cpuset.CPUSet) []*pluginapi.Device {
	var devs []*pluginapi.Device
	cpuSlice := cpus.ToSlice()

	for i := 0; i < cpus.Size(); i++ {
		dev := &pluginapi.Device{
			ID:     strconv.Itoa(cpuSlice[i]),
			Health: pluginapi.Healthy,
		}
		devs = append(devs, dev)
	}
	return devs
}
