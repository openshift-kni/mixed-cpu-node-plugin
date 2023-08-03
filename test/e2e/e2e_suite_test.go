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
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openshift-kni/mixed-cpu-node-plugin/test/e2e/fixture"
	"github.com/openshift-kni/mixed-cpu-node-plugin/test/e2e/infrastructure"
	_ "github.com/openshift-kni/mixed-cpu-node-plugin/test/e2e/mixedcpus"
)

const defaultNamespaceName = "e2e-mixed-cpu-node-plugin"

func TestE2E(t *testing.T) {
	f := fixture.New()
	BeforeSuite(func() {
		Expect(infrastructure.Setup(f.Ctx, f.Cli, GetNamespaceName())).ToNot(HaveOccurred(), "failed setup test infrastructure")
	})

	AfterSuite(func() {
		Expect(infrastructure.Teardown(f.Ctx, f.Cli, GetNamespaceName())).ToNot(HaveOccurred())
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")

}

// GetNamespaceName returns the namespace provided by E2E_NAMESPACE environment variable.
// When E2E_SETUP=true, all infrastructure resources get deployed under this namespace.
func GetNamespaceName() string {
	cpus, ok := os.LookupEnv("E2E_NAMESPACE")
	if !ok {
		return defaultNamespaceName
	}
	return cpus
}
