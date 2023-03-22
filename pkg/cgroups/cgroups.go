package cgroups

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/systemd"
)

const (
	crioPrefix       = "crio"
	cgroupMountPoint = "/sys/fs/cgroup"
)

type Mode string

const (
	cgroupv1            = "cgroupv1"
	cgroupv2UnifiedMode = "cgroupv2UnifiedMode"
)

var Adapter adapter

func init() {
	if cgroups.IsCgroup2UnifiedMode() {
		Adapter = adapter{
			ai:   &v2Adapter{},
			mode: cgroupv2UnifiedMode,
		}
	} else {
		Adapter = adapter{
			ai:   &v1Adapter{},
			mode: cgroupv1,
		}
	}
}

type adapter struct {
	mode Mode
	ai   adapterInterface
}

func (a *adapter) GetMode() Mode {
	return a.mode
}

func (a *adapter) GetCFSQuotaPath(processCgroupPath string) (string, error) {
	return a.ai.cfsQuotaPath(processCgroupPath)
}

func (a *adapter) GetCrioContainerControllerPath(parentPath, ctrId string) (string, error) {
	return a.ai.crioContainerCFSQuotaPath(parentPath, ctrId)
}

type adapterInterface interface {
	cfsQuotaPath(processCgroupPath string) (string, error)
	crioContainerCFSQuotaPath(parentPath, ctrId string) (string, error)
}

type v1Adapter struct{}

func (v1 *v1Adapter) absoluteCgroupPath(processCgroupPath string) (string, error) {
	cpuMountPoint, err := cgroups.FindCgroupMountpoint(cgroupMountPoint, "cpu")
	if err != nil {
		return "", fmt.Errorf("%q: failed to find cgroup mount point: %w", cgroupv1, err)
	}
	processCgroupPath, err = expandSlice(processCgroupPath)
	if err != nil {
		return "", fmt.Errorf("%q: systemd failed to expand slice: %w", cgroupv1, err)
	}
	return filepath.Join(cpuMountPoint, processCgroupPath), nil
}

func (v1 *v1Adapter) cfsQuotaPath(processCgroupPath string) (string, error) {
	absolutePath, err := v1.absoluteCgroupPath(processCgroupPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(absolutePath, "cpu.cfs_quota_us"), nil
}

func (v1 *v1Adapter) crioContainerCFSQuotaPath(parentPath, ctrId string) (string, error) {
	parentAbsolutePath, err := v1.absoluteCgroupPath(parentPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(parentAbsolutePath, crioPrefix+"-"+ctrId+".scope", "cpu.cfs_quota_us"), nil
}

type v2Adapter struct{}

func (v2 *v2Adapter) absoluteCgroupPath(processCgroupPath string) (string, error) {
	var err error
	processCgroupPath, err = expandSlice(processCgroupPath)
	if err != nil {
		return "", fmt.Errorf("%q: systemd failed to expand slice: %w", cgroupv2UnifiedMode, err)
	}
	return filepath.Join(cgroupMountPoint, processCgroupPath), nil
}

func (v2 *v2Adapter) cfsQuotaPath(processCgroupPath string) (string, error) {
	absolutePath, err := v2.absoluteCgroupPath(processCgroupPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(absolutePath, "cpu.max"), nil
}

func (v2 *v2Adapter) crioContainerCFSQuotaPath(parentPath, ctrId string) (string, error) {
	absoluteParentPath, err := v2.absoluteCgroupPath(parentPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(absoluteParentPath, crioPrefix+"-"+ctrId+".scope", "cpu.max"), nil
}

func expandSlice(path string) (string, error) {
	// systemd fs, otherwise cgroupfs
	if strings.HasSuffix(path, ".slice") {
		return systemd.ExpandSlice(path)
	}
	// TODO implement for cgroupfs)
	return "", fmt.Errorf("cgroupfs not implemented")
}
