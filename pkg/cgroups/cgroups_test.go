package cgroups

import (
	"testing"
)

func TestAdapter_GetCFSQuotaPath(t *testing.T) {
	path := "foo-bar-test.slice"

	expectedOutputv1 := "/sys/fs/cgroup/foo.slice/foo-bar.slice/foo-bar-test.slice/cpu.cfs_quota_us"
	expectedOutputv2 := "/sys/fs/cgroup/foo.slice/foo-bar.slice/foo-bar-test.slice/cpu.max"

	quotaPath, err := Adapter.GetCFSQuotaPath(path)
	if err != nil {
		t.Error(err)
	}

	if Adapter.GetMode() == cgroupv1 {
		if quotaPath != expectedOutputv1 {
			t.Errorf("cgroup version %q; want: %q got: %q", Adapter.GetMode(), expectedOutputv1, quotaPath)
		}
	}
	if Adapter.GetMode() == cgroupv2UnifiedMode {
		if quotaPath != expectedOutputv2 {
			t.Errorf("cgroup version %q; want: %q got: %q", Adapter.GetMode(), expectedOutputv2, quotaPath)
		}
	}
}

func TestAdapter_GetCrioContainerCFSQuotaPath(t *testing.T) {
	parentPath := "bob-lucky-test.slice"
	ctrId := "abcdefghijklmnopqrstvuwxyz"

	expectedOutputv1 := "/sys/fs/cgroup/bob.slice/bob-lucky.slice/bob-lucky-test.slice/crio-abcdefghijklmnopqrstvuwxyz.scope/cpu.cfs_quota_us"
	expectedOutputv2 := "/sys/fs/cgroup/bob.slice/bob-lucky.slice/bob-lucky-test.slice/crio-abcdefghijklmnopqrstvuwxyz.scope/cpu.max"

	quotaPath, err := Adapter.GetCrioContainerCFSQuotaPath(parentPath, ctrId)
	if err != nil {
		t.Error(err)
	}

	if Adapter.GetMode() == cgroupv1 {
		if quotaPath != expectedOutputv1 {
			t.Errorf("cgroup version %q; want: %q got: %q", Adapter.GetMode(), expectedOutputv1, quotaPath)
		}
	}
	if Adapter.GetMode() == cgroupv2UnifiedMode {
		if quotaPath != expectedOutputv2 {
			t.Errorf("cgroup version %q; want: %q got: %q", Adapter.GetMode(), expectedOutputv2, quotaPath)
		}
	}
}
