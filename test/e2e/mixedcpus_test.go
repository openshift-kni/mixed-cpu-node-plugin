package e2e_test

import (
	"context"
	"fmt"
	"github.com/Tal-or/nri-mixed-cpu-pools-plugin/pkg/deviceplugin"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	k8sutils "k8s.io/kubernetes/test/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tal-or/nri-mixed-cpu-pools-plugin/test/e2e/pods"
)

const (
	minimalCPUsForTesting  = 2
	minimalNodesForTesting = 1
)

var _ = Describe("Mixedcpus", func() {
	When("a pod gets request for shared cpu device", func() {
		It("should contains the shared cpus under its cgroups", func() {
			nodeList := &corev1.NodeList{}
			Expect(fixture.Cli.List(context.TODO(), nodeList)).ToNot(HaveOccurred())

			var nodes []*corev1.Node
			for i := 0; i < len(nodeList.Items); i++ {
				node := &nodeList.Items[i]
				if node.Status.Allocatable.Cpu().Value() >= minimalCPUsForTesting {
					nodes = append(nodes, node)
				}
			}
			if len(nodes) < minimalNodesForTesting {
				Skip(fmt.Sprintf("minimum of %d nodes with minimum of %d cpus are needed", minimalNodesForTesting, minimalCPUsForTesting))
			}

			By("creating a pod with shared-cpu device")
			pod := pods.Make("test", fixture.NS.Name, pods.WithLimits(corev1.ResourceList{
				corev1.ResourceCPU:               resource.MustParse("1"),
				corev1.ResourceMemory:            resource.MustParse("100M"),
				deviceplugin.MutualCPUDeviceName: resource.MustParse("1"),
			}))
			Expect(fixture.Cli.Create(context.TODO(), pod)).ToNot(HaveOccurred())

			Eventually(func() bool {
				Expect(fixture.Cli.Get(context.TODO(), client.ObjectKeyFromObject(pod), pod)).ToNot(HaveOccurred())
				ready, err := k8sutils.PodRunningReady(pod)
				if !ready {
					klog.Warning(err)
					return false
				}
				return true
			}).WithPolling(time.Second * 5).WithTimeout(time.Minute * 3).WithContext(context.TODO()).Should(BeTrue())

			cpus, err := pods.GetAllowedCPUs(fixture.K8SCli, pod)
			Expect(err).ToNot(HaveOccurred())

			sharedCpus := GetSharedCPUs()
			Expect(sharedCpus).ToNot(BeEmpty())
			sharedCpusSet := cpuset.MustParse(sharedCpus)

			By(fmt.Sprintf("checking if shared CPUs ids %s are presented under pod %s/%s", sharedCpus, pod.Namespace, pod.Name))
			intersect := cpus.Intersection(sharedCpusSet)
			Expect(intersect.Equals(sharedCpusSet)).To(BeTrue(), "shared cpu ids: %s, are not presented. pod: %v cpu ids are: %s", sharedCpusSet.String(), fmt.Sprintf("%s/%s", pod.Namespace, pod.Name), cpus.String())
		})
	})
})
