package infrastructure

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform/detect"
	"github.com/openshift-kni/mixed-cpu-node-plugin/pkg/manifests"
	e2econfig "github.com/openshift-kni/mixed-cpu-node-plugin/test/e2e/config"
	"github.com/openshift-kni/mixed-cpu-node-plugin/test/e2e/wait"
	machineconfigv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
)

const mcpName = "worker"

//go:embed yamls
var dir embed.FS

func Setup(ctx context.Context, c client.Client, ns string) error {
	v := os.Getenv("E2E_SETUP")
	if b, err := strconv.ParseBool(v); err != nil || !b {
		klog.Infof("no setup required E2E_SETUP=%q", v)
		return nil
	}
	return setup(ctx, c, ns)
}

func Teardown(ctx context.Context, c client.Client, ns string) error {
	v := os.Getenv("E2E_TEARDOWN")
	if b, err := strconv.ParseBool(v); err != nil || !b {
		klog.Infof("no teardown required E2E_TEARDOWN=%q", v)
		return nil
	}
	return teardown(ctx, c, ns)
}

func setup(ctx context.Context, c client.Client, ns string) error {
	fullPath := filepath.Join("yamls", "machineconfig.yaml")
	data, err := dir.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read machineconfig.yaml: %w", err)
	}

	mc := &machineconfigv1.MachineConfig{}
	if err = yaml.Unmarshal(data, mc); err != nil {
		return fmt.Errorf("failed to unmarshal machineconfig.yaml: %w", err)
	}

	fullPath = filepath.Join("yamls", "kubeletconfig.yaml")
	data, err = dir.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read kubeletconfig.yaml: %w", err)
	}

	kc := &machineconfigv1.KubeletConfig{}
	if err = yaml.Unmarshal(data, kc); err != nil {
		return fmt.Errorf("failed to unmarshal kubeletconfig.yaml: %w", err)
	}

	if err = c.Create(ctx, mc); err != nil {
		return fmt.Errorf("failed to create %s/%s; %w", mc.Kind, mc.Name, err)
	}

	if err = c.Create(ctx, kc); err != nil {
		return fmt.Errorf("failed to create %s/%s; %w", kc.Kind, kc.Name, err)
	}

	key := client.ObjectKey{
		Name: mcpName,
	}
	klog.Infof("waiting for MCP %q update", key.String())
	if err = wait.ForMCPUpdate(ctx, c, key); err != nil {
		return err
	}

	mf, err := manifests.Get(ns, e2econfig.SharedCPUs())
	if err != nil {
		return err
	}

	findPlatform, s, err := detect.FindPlatform(ctx, platform.Unknown)
	if err != nil {
		return fmt.Errorf("%q; %w", s, err)
	}

	if findPlatform.Discovered == platform.OpenShift {
		updateOpenshiftConfig(mf)
	}

	for _, obj := range mf.ToObjects() {
		if err = c.Create(ctx, obj); err != nil {
			return fmt.Errorf("failed to create %s/%s; %w", obj.GetObjectKind(), obj.GetName(), err)
		}
	}

	klog.Infof("waiting for daemonset %q to be ready", client.ObjectKeyFromObject(&mf.DS))
	err = wait.ForDSReady(ctx, c, client.ObjectKeyFromObject(&mf.DS))
	if err != nil {
		return err
	}
	return nil
}

func teardown(ctx context.Context, c client.Client, ns string) error {
	mf, err := manifests.Get(ns, e2econfig.SharedCPUs())
	if err != nil {
		return err
	}

	if err := c.Delete(ctx, &mf.NS); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete %s/%s", mf.NS.Kind, mf.NS.Name)
		}
	}

	klog.Infof("waiting for namespace %q deletion", mf.NS.Name)
	if err := wait.ForNSDeletion(ctx, c, client.ObjectKeyFromObject(&mf.NS)); err != nil {
		return err
	}

	if err := c.Delete(ctx, &mf.SCC); err != nil {
		return fmt.Errorf("failed to delete %s/%s", mf.SCC.Kind, mf.SCC.Name)
	}

	fullPath := filepath.Join("yamls", "machineconfig.yaml")
	data, err := dir.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read machineconfig.yaml: %w", err)
	}

	mc := &machineconfigv1.MachineConfig{}
	if err = yaml.Unmarshal(data, mc); err != nil {
		return fmt.Errorf("failed to unmarshal machineconfig.yaml: %w", err)
	}

	fullPath = filepath.Join("yamls", "kubeletconfig.yaml")
	data, err = dir.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read kubeletconfig.yaml: %w", err)
	}

	kc := &machineconfigv1.KubeletConfig{}
	if err = yaml.Unmarshal(data, kc); err != nil {
		return fmt.Errorf("failed to unmarshal kubeletconfig.yaml: %w", err)
	}

	if err = c.Delete(ctx, kc); err != nil {
		return fmt.Errorf("failed to delete %s/%s; %w", kc.Kind, kc.Name, err)
	}

	if err = c.Delete(ctx, mc); err != nil {
		return fmt.Errorf("failed to delete %s/%s; %w", mc.Kind, mc.Name, err)
	}

	key := client.ObjectKey{
		Name: mcpName,
	}
	klog.Infof("waiting for MCP %q update", key.String())
	if err = wait.ForMCPUpdate(ctx, c, key); err != nil {
		return err
	}
	return nil
}

func updateOpenshiftConfig(mf *manifests.Manifests) {
	podSpec := &mf.DS.Spec.Template.Spec
	// TODO detect by name
	podSpec.Containers[0].SecurityContext = &corev1.SecurityContext{
		Privileged: pointer.Bool(true),
	}
}
