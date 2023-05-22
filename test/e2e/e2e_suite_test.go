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
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/openshift-kni/mixed-cpu-node-plugin/test/e2e/infrastructure"
	securityv1 "github.com/openshift/api/security/v1"
	machineconfigv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
)

const defaultNamespaceName = "e2e-mixed-cpu-node-plugin"

type TestFixture struct {
	Ctx    context.Context
	Cli    client.Client
	K8SCli *kubernetes.Clientset
	// TestingNS randomly generated for each test and gets deleted when the test ends
	TestingNS *corev1.Namespace
}

var fixture TestFixture

func TestE2e(t *testing.T) {
	BeforeSuite(func() {
		fixture.Ctx = context.Background()
		Expect(initClient()).ToNot(HaveOccurred())
		Expect(initK8SClient()).ToNot(HaveOccurred())
		Expect(infrastructure.Setup(fixture.Ctx, fixture.Cli, GetNamespaceName())).ToNot(HaveOccurred(), "failed setup test infrastructure")
	})

	AfterSuite(func() {
		Expect(infrastructure.Teardown(fixture.Ctx, fixture.Cli, GetNamespaceName())).ToNot(HaveOccurred())
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")

}
// GetNamespaceName returns the namespace provided by E2E_NAMESPACE environment variable.
// When E2E_SETUP=true, all infrastructure resources get deployed under this namespace.
// This has nothing to do with the createNamespace() function
func GetNamespaceName() string {
	cpus, ok := os.LookupEnv("E2E_NAMESPACE")
	if !ok {
		return defaultNamespaceName
	}
	return cpus
}

func Skipf(format string, a ...any) {
	Skip(fmt.Sprintf(format, a...))
}

func initClient() error {
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	if err = machineconfigv1.AddToScheme(scheme.Scheme); err != nil {
		return err
	}

	if err = securityv1.AddToScheme(scheme.Scheme); err != nil {
		return err
	}

	fixture.Cli, err = client.New(cfg, client.Options{})
	return err
}

func initK8SClient() error {
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}
	fixture.K8SCli, err = kubernetes.NewForConfig(cfg)
	return err
}

func createNamespace(prefix string) (*corev1.Namespace, error) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: prefix,
			Labels: map[string]string{
				"security.openshift.io/scc.podSecurityLabelSync": "false",
				"pod-security.kubernetes.io/audit":               "privileged",
				"pod-security.kubernetes.io/enforce":             "privileged",
				"pod-security.kubernetes.io/warn":                "privileged",
			},
		},
	}
	err := fixture.Cli.Create(context.TODO(), ns)
	if err != nil {
		return ns, fmt.Errorf("failed to create namespace %s; %w", ns.Name, err)
	}
	return ns, nil
}

func deleteNamespace(ns *corev1.Namespace) error {
	err := fixture.Cli.Delete(context.TODO(), ns)
	if err != nil {
		return fmt.Errorf("failed deleting namespace %q; %w", ns.Name, err)
	}

	EventuallyWithOffset(1, func() (bool, error) {
		err = fixture.Cli.Get(fixture.Ctx, client.ObjectKeyFromObject(ns), ns)
		if err != nil {
			if !errors.IsNotFound(err) {
				return false, err
			}
			return true, nil
		}
		return false, nil
	}).WithPolling(time.Second*5).WithTimeout(time.Minute*5).Should(BeTrue(), "namespace %q has not been terminated", ns.Name)
	return nil
}
