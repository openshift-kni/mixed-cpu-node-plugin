package pods

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	PauseImage   = "quay.io/openshift-kni/pause:test-ci"
	PauseCommand = "/pause"
)

type Options func(*corev1.Pod)

func Make(name, namespace string, opts ...Options) *corev1.Pod {
	var zero int64
	p := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			TerminationGracePeriodSeconds: &zero,
			Containers: []corev1.Container{
				{
					Name:    name + "-cnt",
					Image:   PauseImage,
					Command: []string{PauseCommand},
				},
			},
		},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func WithRequests(list corev1.ResourceList, ids ...int) func(pod *corev1.Pod) {
	if ids == nil {
		ids = append(ids, 0)
	}
	return func(pod *corev1.Pod) {
		for _, id := range ids {
			pod.Spec.Containers[id].Resources.Requests = list
		}
	}
}

func WithLimits(list corev1.ResourceList, ids ...int) func(pod *corev1.Pod) {
	if ids == nil {
		ids = append(ids, 0)
	}
	return func(pod *corev1.Pod) {
		for _, id := range ids {
			pod.Spec.Containers[id].Resources.Limits = list
		}
	}
}

// GetAllowedCPUs returns a CPUSet of cpus that the pod's containers
// are allowed to access
func GetAllowedCPUs(c *kubernetes.Clientset, pod *corev1.Pod) (*cpuset.CPUSet, error) {
	// TODO is this reliable enough or we should check cgroups directly?
	cmd := []string{"/bin/sh", "-c", "grep Cpus_allowed_list /proc/self/status | cut -f2"}
	out, err := Exec(c, pod, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to run command: %s out: %s; %w", cmd, out, err)
	}

	cpus, err := cpuset.Parse(strings.Trim(string(out), "\r\n"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse cpuset when input is: %s; %w", out, err)
	}
	return &cpus, nil
}

func Exec(c *kubernetes.Clientset, pod *corev1.Pod, command []string) ([]byte, error) {
	var outputBuf bytes.Buffer
	var errorBuf bytes.Buffer

	req := c.CoreV1().RESTClient().
		Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: pod.Spec.Containers[0].Name,
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return nil, err
	}

	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: &outputBuf,
		Stderr: &errorBuf,
		Tty:    true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to run command %v: output %q; error %q; %w", command, outputBuf.String(), errorBuf.String(), err)
	}

	if errorBuf.Len() != 0 {
		return nil, fmt.Errorf("failed to run command %v: output %q; error %q", command, outputBuf.String(), errorBuf.String())
	}

	return outputBuf.Bytes(), nil
}
