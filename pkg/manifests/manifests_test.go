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

package manifests

import (
	"reflect"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"

	securityv1 "github.com/openshift/api/security/v1"
)

func TestSetSharedCPUs(t *testing.T) {
	tcs := []struct {
		cpus    string
		IsError bool
	}{
		{
			cpus:    "0-3,5",
			IsError: false,
		},
		{
			cpus:    "1,5,4",
			IsError: false,
		},
		{
			cpus:    "badformat",
			IsError: true,
		},
		{
			cpus:    "1,6,b",
			IsError: true,
		},
	}
	for _, tc := range tcs {
		mf, err := Get(tc.cpus)
		if !tc.IsError && err != nil {
			t.Errorf("failed to get manifests %v", err)
		}
		if tc.IsError {
			if err == nil {
				t.Errorf("a bad cpu format was given %q, expected error to have occured", tc.cpus)
			}
			continue
		}

		var gotCPUset cpuset.CPUSet
		cnt := &mf.DS.Spec.Template.Spec.Containers[0]
		for _, arg := range cnt.Args {
			keyAndValue := strings.Split(arg, "=")
			if keyAndValue[0] == "--shared-cpus" {
				// we know the format is correct, otherwise Get() would return with an error
				gotCPUset, _ = cpuset.Parse(keyAndValue[1])
				break
			}
		}

		wantCPUset, err := cpuset.Parse(tc.cpus)
		if err != nil {
			t.Error(err)
		}
		if !gotCPUset.Equals(wantCPUset) {
			t.Errorf("shared CPUs were not set correctly; want: %q, got: %q", wantCPUset.String(), gotCPUset.String())
		}
	}
}

func TestGet(t *testing.T) {
	mf, err := Get("0-3,5", WithNewNamespace("unit-test-ns"))
	if err != nil {
		t.Errorf("failed to get manifests %v", err)
	}
	if reflect.DeepEqual(mf.NS, corev1.Namespace{}) {
		t.Errorf("%q object is empty", mf.NS.Kind)
	}
	if reflect.DeepEqual(mf.SA, corev1.ServiceAccount{}) {
		t.Errorf("%q object is empty", mf.SA.Kind)
	}
	if reflect.DeepEqual(mf.DS, appsv1.DaemonSet{}) {
		t.Errorf("%q object is empty", mf.DS.Kind)
	}
	if reflect.DeepEqual(mf.Role, rbacv1.Role{}) {
		t.Errorf("%q object is empty", mf.Role.Kind)
	}
	if reflect.DeepEqual(mf.RB, rbacv1.RoleBinding{}) {
		t.Errorf("%q object is empty", mf.RB.Kind)
	}
	if reflect.DeepEqual(mf.SCC, securityv1.SecurityContextConstraints{}) {
		t.Errorf("%q object is empty", mf.SCC.Kind)
	}
	if mf.NS.Name != "unit-test-ns" {
		t.Errorf("wrong namespace name %q", mf.NS.Name)
	}

	mf, err = Get("1,2,4", WithNamespace("unit-test-12"))
	if err != nil {
		t.Errorf("failed to get manifests %v", err)
	}
	if !reflect.DeepEqual(mf.NS, corev1.Namespace{}) {
		t.Errorf("should have an empty namespace when WithNamespace called")
	}
	if mf.DS.Namespace != "unit-test-12" {
		t.Errorf("%q object namespace not set", mf.DS.Kind)
	}

	mf, err = Get("1,2,4")
	if err != nil {
		t.Errorf("failed to get manifests %v", err)
	}
	if !reflect.DeepEqual(mf.NS, corev1.Namespace{}) {
		t.Errorf("should have an empty namespace when WithNamespace called")
	}
	if mf.DS.Namespace != "" {
		t.Errorf("%q object namespace should be set to default", mf.DS.Kind)
	}

	mf, err = Get("1,2,4", WithName("foo"))
	if err != nil {
		t.Errorf("failed to get manifests %v", err)
	}
	if mf.DS.Name != "foo" {
		t.Errorf("%q object name should be equal to foo", mf.DS.Kind)
	}
	if mf.RB.Subjects[0].Name != mf.SA.Name {
		t.Errorf("%q -> subject[0] -> name should be equal to %s", mf.RB.Kind, mf.SA.Name)
	}
	if mf.RB.Subjects[0].Namespace != mf.SA.Namespace {
		t.Errorf("%q -> subject[0] -> namespace should be equal to %s", mf.RB.Kind, mf.SA.Namespace)
	}
}
