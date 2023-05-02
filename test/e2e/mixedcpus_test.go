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

package e2e_test

import (
	"context"
	"fmt"
	"github.com/Tal-or/nri-mixed-cpu-pools-plugin/pkg/deviceplugin"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
	"time"

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
	BeforeEach(func() {
		Expect(createNamespace("mixedcpus-testing-")).ToNot(HaveOccurred())
		DeferCleanup(deleteNamespace, fixture.NS)
	})

	When("a pod gets request for shared cpu device", func() {
		var pod *corev1.Pod
		BeforeEach(func() {
			By("checking if minimal cpus are available for testing")
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
				Skipf("minimum of %d nodes with minimum of %d cpus are needed", minimalNodesForTesting, minimalCPUsForTesting)
			}

			By("creating a pod with shared-cpu device")
			pod = pods.Make("test", fixture.NS.Name, pods.WithLimits(corev1.ResourceList{
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
		})

		It("should contain the shared cpus under its cgroups", func() {
			cpus, err := pods.GetAllowedCPUs(fixture.K8SCli, pod)
			Expect(err).ToNot(HaveOccurred())

			sharedCpus := GetSharedCPUs()
			Expect(sharedCpus).ToNot(BeEmpty())
			sharedCpusSet := cpuset.MustParse(sharedCpus)

			By(fmt.Sprintf("checking if shared CPUs ids %s are presented under pod %s/%s", sharedCpus, pod.Namespace, pod.Name))
			intersect := cpus.Intersection(sharedCpusSet)
			Expect(intersect.Equals(sharedCpusSet)).To(BeTrue(), "shared cpu ids: %s, are not presented. pod: %v cpu ids are: %s", sharedCpusSet.String(), fmt.Sprintf("%s/%s", pod.Namespace, pod.Name), cpus.String())
		})

		It("can have more than one pod accessing shared cpus", func() {

		})

		It("should contain OPENSHIFT_MUTUAL_CPUS environment variable", func() {

		})
	})

	When("[Slow][Reboot] node goes into reboot", func() {
		It("should have all pods with shared cpus running after it goes back up", func() {

		})
	})
})
