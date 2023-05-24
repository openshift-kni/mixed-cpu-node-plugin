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
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift-kni/mixed-cpu-node-plugin/internal/nodes"
	"github.com/openshift-kni/mixed-cpu-node-plugin/internal/pods"
	"github.com/openshift-kni/mixed-cpu-node-plugin/internal/wait"
	"github.com/openshift-kni/mixed-cpu-node-plugin/pkg/deviceplugin"
	e2econfig "github.com/openshift-kni/mixed-cpu-node-plugin/test/e2e/config"
	e2ecpuset "github.com/openshift-kni/mixed-cpu-node-plugin/test/e2e/cpuset"
)

const (
	minimalCPUsForTesting  = 2
	minimalNodesForTesting = 1
)

var _ = Describe("Mixedcpus", func() {
	BeforeEach(func() {
		ns, err := createNamespace("mixedcpus-testing-")
		Expect(err).ToNot(HaveOccurred())
		fixture.TestingNS = ns
		DeferCleanup(deleteNamespace, fixture.TestingNS)
	})

	When("a pod gets request for shared cpu device", func() {
		var dp *appsv1.Deployment
		var pod *corev1.Pod
		BeforeEach(func() {
			checkMinimalCPUsForTesting()
			dp = createDeployment("dp-test")
			items, err := pods.OwnedByDeployment(fixture.Ctx, fixture.Cli, dp)
			Expect(err).ToNot(HaveOccurred())
			replicas := int(*dp.Spec.Replicas)
			Expect(len(items)).To(Equal(replicas), "expected to find %d pods for deployment=%q found %d; %v", replicas, dp.Name, len(items), items)
			pod = &items[0]
		})

		It("should contain the shared cpus under its cgroups", func() {
			cpus, err := pods.GetAllowedCPUs(fixture.K8SCli, pod)
			Expect(err).ToNot(HaveOccurred())

			sharedCpus := e2econfig.SharedCPUs()
			Expect(sharedCpus).ToNot(BeEmpty())
			sharedCpusSet := e2ecpuset.MustParse(sharedCpus)

			By(fmt.Sprintf("checking if shared CPUs ids %s are presented under pod %s/%s", sharedCpus, pod.Namespace, pod.Name))
			intersect := cpus.Intersection(sharedCpusSet)
			Expect(intersect.Equals(sharedCpusSet)).To(BeTrue(), "shared cpu ids: %s, are not presented. pod: %v cpu ids are: %s", sharedCpusSet.String(), fmt.Sprintf("%s/%s", pod.Namespace, pod.Name), cpus.String())
		})

		It("can have more than one pod accessing shared cpus", func() {
			dp2 := createDeployment("dp-test2")
			items, err := pods.OwnedByDeployment(fixture.Ctx, fixture.Cli, dp2)
			Expect(err).ToNot(HaveOccurred())
			replicas := int(*dp2.Spec.Replicas)
			Expect(len(items)).To(Equal(replicas), "expected to find %d pods for deployment=%q found %d; %v", replicas, dp2.Name, len(items), items)
			pod2 := &items[0]

			By("check the second pod successfully deployed with shared cpus")
			cpus, err := pods.GetAllowedCPUs(fixture.K8SCli, pod2)
			Expect(err).ToNot(HaveOccurred())

			sharedCpus := e2econfig.SharedCPUs()
			Expect(sharedCpus).ToNot(BeEmpty())
			sharedCpusSet := e2ecpuset.MustParse(sharedCpus)

			By(fmt.Sprintf("checking if shared CPUs ids %s are presented under pod %s/%s", sharedCpus, pod2.Namespace, pod2.Name))
			intersect := cpus.Intersection(sharedCpusSet)
			Expect(intersect.Equals(sharedCpusSet)).To(BeTrue(), "shared cpu ids: %s, are not presented. pod: %v cpu ids are: %s", sharedCpusSet.String(), fmt.Sprintf("%s/%s", pod2.Namespace, pod2.Name), cpus.String())
		})

		It("should contain OPENSHIFT_MUTUAL_CPUS environment variable", func() {
			sharedCpus := e2econfig.SharedCPUs()
			Expect(sharedCpus).ToNot(BeEmpty())

			sharedCpusSet := e2ecpuset.MustParse(sharedCpus)
			out, err := pods.Exec(fixture.K8SCli, pod, []string{"/bin/printenv", "OPENSHIFT_MUTUAL_CPUS"})
			Expect(err).ToNot(HaveOccurred())
			Expect(out).ToNot(BeEmpty(), "OPENSHIFT_MUTUAL_CPUS environment variable was not found")

			envVar := strings.Trim(string(out), "\r\n")
			sharedCpusFromEnv, err := cpuset.Parse(envVar)
			Expect(err).ToNot(HaveOccurred(), "failed parse %q to cpuset", sharedCpusFromEnv)
			Expect(sharedCpusSet.Equals(sharedCpusFromEnv)).To(BeTrue())
		})
	})

	When("[Slow][Reboot] node goes into reboot", func() {
		var dp *appsv1.Deployment
		var pod *corev1.Pod
		BeforeEach(func() {
			checkMinimalCPUsForTesting()
			dp = createDeployment("dp-test")
			items, err := pods.OwnedByDeployment(fixture.Ctx, fixture.Cli, dp)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(items)).To(Equal(int(*dp.Spec.Replicas)), "expected to find %d pods for deployment=%q found %d; %v", int(*dp.Spec.Replicas), dp.Name, len(items), items)
			pod = &items[0]
		})

		It("should have all pods with shared cpus running after it goes back up", func() {
			nodeName := pod.Spec.NodeName
			By(fmt.Sprintf("call reboot on node %q", nodeName))
			_, err := nodes.ExecCommand(fixture.Ctx, fixture.K8SCli, nodeName, []string{"chroot", "/rootfs", "systemctl", "reboot"})
			Expect(err).ToNot(HaveOccurred(), "failed to execute reboot on node %q", nodeName)
			By(fmt.Sprintf("wait for node %q to be ready", nodeName))
			Expect(wait.ForNodeReady(fixture.Ctx, fixture.Cli, client.ObjectKey{Name: nodeName})).ToNot(HaveOccurred())
			By(fmt.Sprintf("node %q is ready, moving on with testing", nodeName))
			Expect(wait.ForDeploymentReady(fixture.Ctx, fixture.Cli, client.ObjectKeyFromObject(dp))).ToNot(HaveOccurred())

			By("check pod successfully deployed with shared cpus")
			// new pod got created after reboot, we need to find it again
			items, err := pods.OwnedByDeployment(fixture.Ctx, fixture.Cli, dp)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(items)).To(Equal(int(*dp.Spec.Replicas)), "expected to find %d pods for deployment=%q found %d; %v", int(*dp.Spec.Replicas), dp.Name, len(items), items)
			pod = &items[0]

			cpus, err := pods.GetAllowedCPUs(fixture.K8SCli, pod)
			Expect(err).ToNot(HaveOccurred())

			sharedCpus := e2econfig.SharedCPUs()
			Expect(sharedCpus).ToNot(BeEmpty())
			sharedCpusSet := e2ecpuset.MustParse(sharedCpus)

			By(fmt.Sprintf("checking if shared CPUs ids %s are presented under pod %s/%s", sharedCpus, pod.Namespace, pod.Name))
			intersect := cpus.Intersection(sharedCpusSet)
			Expect(intersect.Equals(sharedCpusSet)).To(BeTrue(), "shared cpu ids: %s, are not presented. pod: %v cpu ids are: %s", sharedCpusSet.String(), fmt.Sprintf("%s/%s", pod.Namespace, pod.Name), cpus.String())
		})
	})
})

func checkMinimalCPUsForTesting() {
	nodeList := &corev1.NodeList{}
	ExpectWithOffset(1, fixture.Cli.List(context.TODO(), nodeList)).ToNot(HaveOccurred())
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
}

func createDeployment(name string) *appsv1.Deployment {
	pod := pods.Make("pod-test", fixture.TestingNS.Name, pods.WithLimits(corev1.ResourceList{
		corev1.ResourceCPU:               resource.MustParse("1"),
		corev1.ResourceMemory:            resource.MustParse("100M"),
		deviceplugin.MutualCPUDeviceName: resource.MustParse("1"),
	}))
	labelsMap := map[string]string{"name": name}
	dp := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: fixture.TestingNS.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsMap,
			},
			Replicas: pointer.Int32(1),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelsMap,
				},
				Spec: pod.Spec,
			},
		},
	}
	klog.Infof("create deployment %q with a pod requesting for shared cpus", client.ObjectKeyFromObject(dp).String())
	ExpectWithOffset(1, fixture.Cli.Create(context.TODO(), dp)).ToNot(HaveOccurred(), "failed to create deployment %q", client.ObjectKeyFromObject(dp).String())
	ExpectWithOffset(1, wait.ForDeploymentReady(fixture.Ctx, fixture.Cli, client.ObjectKeyFromObject(dp))).ToNot(HaveOccurred())
	return dp
}
