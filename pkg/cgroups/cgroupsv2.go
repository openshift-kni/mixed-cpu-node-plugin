package cgroups

import (
	"fmt"
	"path/filepath"
)

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
