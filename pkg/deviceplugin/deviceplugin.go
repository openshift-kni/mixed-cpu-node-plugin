package deviceplugin

import (
	"strconv"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"

	"github.com/containerd/nri/pkg/api"
	"github.com/containers/podman/v4/pkg/env"
	"github.com/golang/glog"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
)

const (
	MutualCPUResourceNamespace = "openshift.io"
	MutualCPUResourceName      = "mutualcpu"
	MutualCPUDeviceName        = MutualCPUResourceNamespace + "/" + MutualCPUResourceName
	EnvVarName                 = "OPENSHIFT_MUTUAL_CPUS"
	// TODO is this number big enough?
	DefaultDevicesNumber = 99
)

type MutualCpu struct {
	cpus cpuset.CPUSet
}

func (mc *MutualCpu) GetResourceNamespace() string {
	return MutualCPUResourceNamespace
}

func (mc *MutualCpu) Discover(pnl chan dpm.PluginNameList) {
	pnl <- []string{MutualCPUResourceName}
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

func MakeMutualCpusDevices() []*pluginapi.Device {
	var devs []*pluginapi.Device

	for i := 0; i < DefaultDevicesNumber; i++ {
		dev := &pluginapi.Device{
			ID:     strconv.Itoa(i),
			Health: pluginapi.Healthy,
		}
		devs = append(devs, dev)
	}
	return devs
}

// Requested checks whether a given container is requesting the device
func Requested(ctr *api.Container) bool {
	if ctr.Env == nil {
		return false
	}

	envs, err := env.ParseSlice(ctr.Env)
	if err != nil {
		glog.Errorf("failed to parse environment variables for container: %q; err: %v", ctr.Name, err)
		return false
	}

	for k, v := range envs {
		if k == EnvVarName {
			glog.V(4).Infof("shared CPUs ids: %q allocated for container: %q", v, ctr.Name)
			return true
		}
	}
	return false
}
