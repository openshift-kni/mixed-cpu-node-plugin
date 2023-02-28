package nriplugin

import (
	"fmt"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"

	"github.com/containerd/nri/pkg/api"
)

const (
	sampleCPUs = "0,5,7-10"
)

func TestCreateContainer(t *testing.T) {
	testCases := []struct {
		name       string
		mutualCPUs cpuset.CPUSet
		sb         *api.PodSandbox
		ctr        *api.Container
		lres       *api.LinuxResources
		quota      int64
		cpuset     string
	}{
		{
			name:       "pod without annotation",
			mutualCPUs: cpuset.MustParse(sampleCPUs),
			sb:         makePodSandbox("test-sb"),
			ctr:        makeContainer("test-ctr", withLinuxResources("1,2", 20000)),
			lres:       nil,
			quota:      20000,
			cpuset:     "1,2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &Plugin{
				Stub:       nil,
				MutualCPUs: &tc.mutualCPUs,
			}
			ca, _, err := p.CreateContainer(tc.sb, tc.ctr)
			if err != nil {
				t.Fatal(err)
			}
			if tc.lres != nil {
				lcpu := ca.Linux.Resources.Cpu
				if tc.cpuset != lcpu.Cpus {
					t.Fatalf("unexpected cpuset; want: %q, got: %q", tc.cpuset, lcpu.Cpus)
				}
				if tc.quota != lcpu.Quota.Value {
					t.Fatalf("unexpected quota; want: %q, got: %q", tc.quota, lcpu.Quota.Value)
				}
			} else {
				if ca.Linux != nil {
					t.Fatalf("expected api.LinuxContainerAdjustment response to be nil")
				}
			}
		})
	}
}

func makePodSandbox(name string, opts ...func(sb *api.PodSandbox)) *api.PodSandbox {
	uid := string(uuid.NewUUID())
	sb := &api.PodSandbox{
		Name: name,
		Id:   uid,
		Linux: &api.LinuxPodSandbox{
			CgroupParent: generateCgroupParent(uid),
		},
	}
	for _, opt := range opts {
		opt(sb)
	}
	return sb

}

func makeContainer(name string, opts ...func(ctr *api.Container)) *api.Container {
	ctr := &api.Container{
		Name:  name,
		Linux: &api.LinuxContainer{},
	}
	for _, opt := range opts {
		opt(ctr)
	}
	return ctr
}

func withLinuxResources(cpus string, quota int64) func(ctr *api.Container) {
	lres := &api.LinuxResources{
		Cpu: &api.LinuxCPU{
			Quota: &api.OptionalInt64{Value: quota},
			Cpus:  cpus,
		},
	}
	return func(ctr *api.Container) {
		ctr.Linux.Resources = lres
	}
}

func generateCgroupParent(uid string) string {
	return fmt.Sprintf("kubepods.slice/kubepods-pod%s.slice", strings.Replace(uid, "-", "_", -1))
}
