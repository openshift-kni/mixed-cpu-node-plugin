package mixedcpus

import (
	"fmt"
	containerresources "github.com/Tal-or/nri-mixed-cpu-pools-plugin/pkg/cntrresources"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"

	"github.com/containerd/nri/pkg/api"
	"github.com/containerd/nri/pkg/stub"
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/systemd"

	"github.com/Tal-or/nri-mixed-cpu-pools-plugin/pkg/annotations"
)

const (
	milliCPUToCPU    = 1000
	cgroupMountPoint = "/sys/fs/cgroup"
	crioPrefix       = "crio"
)

// Plugin mixedcpus plugin
type Plugin struct {
	Stub         stub.Stub
	ReservedCPUs *cpuset.CPUSet
	ctrStates    map[string]containerresources.State
}

type Args struct {
	PluginName   string
	PluginIdx    string
	ReservedCPUs string
}

func New(args *Args) (*Plugin, error) {
	p := &Plugin{}
	var opts []stub.Option
	p.ctrStates = make(map[string]containerresources.State)

	if args.PluginName != "" {
		opts = append(opts, stub.WithPluginName(args.PluginName))
	}
	if args.PluginIdx != "" {
		opts = append(opts, stub.WithPluginIdx(args.PluginIdx))
	}
	c, err := cpuset.Parse(args.ReservedCPUs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cpuset %q: %w", args.ReservedCPUs, err)
	}
	if c.Size() <= 4 {
		return p, fmt.Errorf("reserved CPUs must be more than 4")
	}
	klog.Infof("node %q reserved CPUs: %q", os.ExpandEnv("$NODE_NAME"), c.String())
	p.ReservedCPUs = &c

	if p.Stub, err = stub.New(p, opts...); err != nil {
		return nil, fmt.Errorf("failed to create plugin stub: %w", err)
	}
	return p, nil
}

// CreateContainer handles container creation requests.
func (p *Plugin) CreateContainer(pod *api.PodSandbox, ctr *api.Container) (*api.ContainerAdjustment, []*api.ContainerUpdate, error) {
	adjustment := &api.ContainerAdjustment{}
	updates := []*api.ContainerUpdate{}

	if !annotations.IsMutualCPUsEnabled(pod.Annotations) {
		return adjustment, updates, nil
	}
	klog.Infof("Append mutual cpus to container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())
	if err := setMutualCPUs(adjustment, ctr, p.ReservedCPUs); err != nil {
		return adjustment, updates, fmt.Errorf("setMutualCPUs failed: %w", err)
	}

	//Adding mutual cpus without increasing cpuQuota,
	//might result with throttling the processes' threads
	//if the threads that are running under the mutual cpus
	//oversteps their boundaries, or the threads that are running
	//under the reserved cpus consumes the cpuQuota (pretty common in dpdk/latency sensitive applications).
	//Since we can't determine the cpuQuota for the mutual cpus
	//and avoid throttling the process is critical, increasing the cpuQuota to the maximum is the best option.
	quota, err := calculateCFSQuota(ctr)
	if err != nil {
		return adjustment, updates, fmt.Errorf("calculateCFSQuota failed: %w", err)
	}

	cpuMountPoint, err := cgroups.FindCgroupMountpoint(cgroupMountPoint, "cpu")
	if err != nil {
		return adjustment, updates, fmt.Errorf("FindCgroupMountpoint failed: %w", err)
	}
	parentPath := pod.GetLinux().GetCgroupParent()
	var ctrPath string

	// systemd fs, otherwise cgroupfs
	if strings.HasSuffix(parentPath, ".slice") {
		parentPath, err = systemd.ExpandSlice(parentPath)
		if err != nil {
			return adjustment, updates, fmt.Errorf("FindCgroupMountpoint failed: %w", err)
		}
		// TODO this is for systemd. it needs to by dynamic (i.e for cgroupfs)
		ctrPath = filepath.Join(parentPath, crioPrefix+"-"+ctr.GetId()+".scope")
	}

	parentCfsQuotaPath := filepath.Join(cpuMountPoint, parentPath, "cpu.cfs_quota_us")
	ctrCfsQuotaPath := filepath.Join(cpuMountPoint, ctrPath, "cpu.cfs_quota_us")
	klog.Infof("Inject hook to modify container's cgroups %q quota to: %d", ctrCfsQuotaPath, quota)
	hook := &api.Hook{
		Path: "/bin/bash",
		Args: []string{
			"/bin/bash",
			"-c",
			fmt.Sprintf("echo %d > %s && echo %d > %s", quota, parentCfsQuotaPath, quota, ctrCfsQuotaPath),
		},
	}
	adjustment.Hooks = &api.Hooks{
		CreateRuntime: []*api.Hook{hook},
	}
	adjustment.Linux = &api.LinuxContainerAdjustment{
		Resources: ctr.Linux.GetResources(),
	}

	klog.Infof("adjustment: %+v", adjustment)
	return adjustment, updates, nil
}

func (p *Plugin) UpdateContainer(pod *api.PodSandbox, ctr *api.Container) ([]*api.ContainerUpdate, error) {
	updates := []*api.ContainerUpdate{}
	if !annotations.IsMutualCPUsEnabled(pod.Annotations) {
		return updates, nil
	}
	klog.Infof("Updating container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())
	// do nothing but store the original resources
	// so CRIO code won't crash with nil pointer
	//res := api.ContainerUpdate{
	//	//ContainerId: ctr.Id,
	//	Linux: &api.LinuxContainerUpdate{
	//		Resources: ctr.Linux.Resources,
	//	},
	//}
	return updates, nil
}

func setMutualCPUs(adjustment *api.ContainerAdjustment, ctr *api.Container, reservedCPUs *cpuset.CPUSet) error {
	lspec := ctr.GetLinux()
	if lspec == nil ||
		lspec.Resources == nil ||
		lspec.Resources.Cpu == nil ||
		lspec.Resources.Cpu.Cpus == "" {
		return fmt.Errorf("no cpus found")
	}
	ctrCpus := lspec.Resources.Cpu
	rcpus := reservedCPUs.ToSliceNoSort()
	var mutualCPUsSlice []int
	// 4 is the number of cpus that are needed
	// for housekeeping tasks, hence do not
	// use them as shared cpus
	for i := 4; i < reservedCPUs.Size(); i++ {
		mutualCPUsSlice = append(mutualCPUsSlice, rcpus[i])
	}
	if len(mutualCPUsSlice) == 0 {
		return fmt.Errorf("no mutual cpus found")
	}

	mutualCPUs := cpuset.NewCPUSet(mutualCPUsSlice...)
	curCpus, err := cpuset.Parse(ctrCpus.Cpus)
	klog.Infof("curCpus: %q", curCpus.String())
	if err != nil {
		return err
	}

	ctrCpus.Cpus = curCpus.Union(mutualCPUs).String()
	// set an environment variable to
	// reflect the mutual CPUs
	adjustment.Env = []*api.KeyValue{
		{
			Key:   "OPENSHIFT_MUTUAL_CPUS",
			Value: mutualCPUs.String(),
		},
	}
	return nil
}

func calculateCFSQuota(ctr *api.Container) (quota int64, err error) {
	lspec := ctr.Linux
	cpus, err := cpuset.Parse(lspec.Resources.Cpu.Cpus)
	if err != nil {
		return
	}
	quan, err := resource.ParseQuantity(strconv.Itoa(cpus.Size()))
	if err != nil {
		return
	}

	quota = (quan.MilliValue() * int64(lspec.Resources.Cpu.Period.Value)) / milliCPUToCPU
	//lspec.Resources.Cpu.Quota.Value = quota
	return
}
