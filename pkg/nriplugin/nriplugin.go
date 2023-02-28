package nriplugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"

	"github.com/containerd/nri/pkg/api"
	"github.com/containerd/nri/pkg/stub"
	"github.com/golang/glog"
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/systemd"

	"github.com/Tal-or/nri-mixed-cpu-pools-plugin/pkg/annotations"
)

const (
	milliCPUToCPU    = 1000
	cgroupMountPoint = "/sys/fs/cgroup"
	crioPrefix       = "crio"
)

// Plugin nriplugin for mixed cpus
type Plugin struct {
	Stub       stub.Stub
	MutualCPUs *cpuset.CPUSet
}

type Args struct {
	PluginName string
	PluginIdx  string
	MutualCPUs string
}

func New(args *Args) (*Plugin, error) {
	p := &Plugin{}
	var opts []stub.Option

	if args.PluginName != "" {
		opts = append(opts, stub.WithPluginName(args.PluginName))
	}
	if args.PluginIdx != "" {
		opts = append(opts, stub.WithPluginIdx(args.PluginIdx))
	}
	c, err := cpuset.Parse(args.MutualCPUs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cpuset %q: %w", args.MutualCPUs, err)
	}
	if c.Size() == 0 {
		return p, fmt.Errorf("there has to be at least one mutual CPU")
	}
	glog.Infof("node %q mutual CPUs: %q", os.ExpandEnv("$NODE_NAME"), c.String())
	p.MutualCPUs = &c

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
	glog.Infof("append mutual cpus to container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())
	err := setMutualCPUs(ctr, p.MutualCPUs)
	if err != nil {
		return adjustment, updates, fmt.Errorf("CreateContainer: setMutualCPUs failed: %w", err)
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
		return adjustment, updates, fmt.Errorf("failed to calculate CFS quota: %w", err)
	}

	cpuMountPoint, err := cgroups.FindCgroupMountpoint(cgroupMountPoint, "cpu")
	if err != nil {
		return adjustment, updates, fmt.Errorf("failed to find cgroup mount point: %w", err)
	}
	parentPath := pod.GetLinux().GetCgroupParent()
	var ctrPath string

	// systemd fs, otherwise cgroupfs
	if strings.HasSuffix(parentPath, ".slice") {
		parentPath, err = systemd.ExpandSlice(parentPath)
		if err != nil {
			return adjustment, updates, err
		}
		// TODO this is for systemd. it needs to by dynamic (i.e for cgroupfs)
		ctrPath = filepath.Join(parentPath, crioPrefix+"-"+ctr.GetId()+".scope")
	}

	parentCfsQuotaPath := filepath.Join(cpuMountPoint, parentPath, "cpu.cfs_quota_us")
	ctrCfsQuotaPath := filepath.Join(cpuMountPoint, ctrPath, "cpu.cfs_quota_us")
	glog.Infof("inject hook to modify container's cgroups %q quota to: %d", ctrCfsQuotaPath, quota)
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

	glog.V(4).Infof("sending adjustment to runtime: %+v", adjustment)
	return adjustment, updates, nil
}

func (p *Plugin) UpdateContainer(pod *api.PodSandbox, ctr *api.Container) ([]*api.ContainerUpdate, error) {
	updates := []*api.ContainerUpdate{}
	if !annotations.IsMutualCPUsEnabled(pod.Annotations) {
		return updates, nil
	}

	glog.Infof("updating container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())
	curCpus, err := cpuset.Parse(ctr.Linux.Resources.Cpu.Cpus)
	if err != nil {
		return nil, fmt.Errorf("failed to parse container %q cpuset %w", ctr.Id, err)
	}
	// bypass updates coming from CPUManager
	ctr.Linux.Resources.Cpu.Cpus = curCpus.Union(*p.MutualCPUs).String()
	quota, err := calculateCFSQuota(ctr)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate CFS quota: %w", err)
	}
	ctr.Linux.Resources.Cpu.Quota.Value = quota

	res := &api.ContainerUpdate{
		ContainerId: ctr.Id,
		Linux: &api.LinuxContainerUpdate{
			Resources: ctr.Linux.Resources,
		},
	}
	updates = append(updates, res)
	glog.V(4).Infof("sending update to runtime: %+v", updates)
	return updates, nil
}

func setMutualCPUs(ctr *api.Container, mutualCPUs *cpuset.CPUSet) error {
	lspec := ctr.GetLinux()
	if lspec == nil ||
		lspec.Resources == nil ||
		lspec.Resources.Cpu == nil ||
		lspec.Resources.Cpu.Cpus == "" {
		return fmt.Errorf("no cpus found")
	}
	ctrCpus := lspec.Resources.Cpu
	curCpus, err := cpuset.Parse(ctrCpus.Cpus)
	glog.Infof("curCpus: %q", curCpus.String())
	if err != nil {
		return err
	}

	ctrCpus.Cpus = curCpus.Union(*mutualCPUs).String()
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
	return
}
