/*
 * Copyright 2023 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cgroups

import (
	"testing"
)

func TestAdapter_GetCFSQuotaPath(t *testing.T) {
	path := "foo-bar-test.slice"

	expectedOutputv1 := "/sys/fs/cgroup/cpu,cpuacct/foo.slice/foo-bar.slice/foo-bar-test.slice/cpu.cfs_quota_us"
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

	expectedOutputv1 := "/sys/fs/cgroup/cpu,cpuacct/bob.slice/bob-lucky.slice/bob-lucky-test.slice/crio-abcdefghijklmnopqrstvuwxyz.scope/cpu.cfs_quota_us"
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
