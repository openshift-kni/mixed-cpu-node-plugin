package cgroups

import (
	"fmt"
	"path/filepath"

	"github.com/opencontainers/runc/libcontainer/cgroups"
)

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
