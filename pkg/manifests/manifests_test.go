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
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	securityv1 "github.com/openshift/api/security/v1"
)

func TestGet(t *testing.T) {
	mf, err := Get()
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
	if reflect.DeepEqual(mf.SSC, securityv1.SecurityContextConstraints{}) {
		t.Errorf("%q object is empty", mf.SSC.Kind)
	}
}
