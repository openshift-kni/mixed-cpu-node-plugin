module github.com/Tal-or/nri-mixed-cpu-pools-plugin

go 1.19

require (
	github.com/containerd/nri v0.2.0
	github.com/golang/glog v1.0.0
	github.com/google/go-cmp v0.5.9
	github.com/kubevirt/device-plugin-manager v1.19.4
	github.com/opencontainers/runc v1.1.4
	google.golang.org/grpc v1.51.0
	k8s.io/kubelet v0.26.0
	k8s.io/kubernetes v1.15.0-alpha.0
)

require (
	github.com/cilium/ebpf v0.7.0 // indirect
	github.com/containerd/ttrpc v1.1.1-0.20220420014843-944ef4a40df3 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/cyphar/filepath-securejoin v0.2.3 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/godbus/dbus/v5 v5.1.1-0.20221029134443-4b691ce883d5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/moby/sys/mountinfo v0.6.2 // indirect
	github.com/onsi/ginkgo/v2 v2.8.3 // indirect
	github.com/onsi/gomega v1.27.0 // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20220825212826-86290f6a00fb // indirect
	github.com/seccomp/libseccomp-golang v0.10.0 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/stretchr/testify v1.8.1 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/genproto v0.0.0-20221227171554-f9683d7f8bef // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/klog/v2 v2.80.1 // indirect
)

// k8s deps
require (
	k8s.io/apimachinery v0.26.0
	k8s.io/cri-api v0.25.3 // indirect

)

replace (
	k8s.io/api => k8s.io/api v0.26.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.26.0
	k8s.io/apiserver => k8s.io/apiserver v0.26.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.26.0
	k8s.io/client-go => k8s.io/client-go v0.26.0
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.26.0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.26.0
	k8s.io/code-generator => k8s.io/code-generator v0.26.0
	k8s.io/component-base => k8s.io/component-base v0.26.0
	k8s.io/component-helpers => k8s.io/component-helpers v0.26.0
	k8s.io/controller-manager => k8s.io/controller-manager v0.26.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.26.0
	k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.26.0
	k8s.io/kms => k8s.io/kms v0.26.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.26.0
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.26.0
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.26.0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.26.0
	k8s.io/kubectl => k8s.io/kubectl v0.26.0
	k8s.io/kubelet => k8s.io/kubelet v0.26.0
	k8s.io/kubernetes => k8s.io/kubernetes v1.26.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.26.0
	k8s.io/metrics => k8s.io/metrics v0.26.0
	k8s.io/mount-utils => k8s.io/mount-utils v0.26.0
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.26.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.26.0
)
