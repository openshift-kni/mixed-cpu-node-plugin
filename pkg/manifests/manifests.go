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
	"embed"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	securityv1 "github.com/openshift/api/security/v1"
)

//go:embed yamls
var dir embed.FS

type Manifests struct {
	NS   corev1.Namespace
	SA   corev1.ServiceAccount
	DS   appsv1.DaemonSet
	Role rbacv1.Role
	RB   rbacv1.RoleBinding
	SSC  securityv1.SecurityContextConstraints
}

func Get() (*Manifests, error) {
	mf := Manifests{}
	var fileToObject = map[string]metav1.Object{
		"serviceaccount.yaml":            &mf.SA,
		"namespace.yaml":                 &mf.NS,
		"daemonset.yaml":                 &mf.DS,
		"role.yaml":                      &mf.Role,
		"rolebinding.yaml":               &mf.RB,
		"securitycontextconstraint.yaml": &mf.SSC,
	}

	files, err := dir.ReadDir("yamls")
	if err != nil {
		return nil, fmt.Errorf("failed to read yamls directory: %w", err)
	}

	for _, f := range files {
		fullPath := filepath.Join("yamls", f.Name())
		data, err := dir.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %q: %w", fullPath, err)
		}
		if _, ok := fileToObject[f.Name()]; !ok {
			return nil, fmt.Errorf("key %q does not exist", f.Name())
		}
		if err := yaml.Unmarshal(data, fileToObject[f.Name()]); err != nil {
			return nil, fmt.Errorf("failed to unmarshal file %q: %w", "bla", err)
		}
	}

	return &mf, nil
}
