module app

go 1.20

replace (
	k8s.io/api => k8s.io/api v0.27.1
	k8s.io/apiserver => k8s.io/apiserver v0.27.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.27.1
	k8s.io/client-go => k8s.io/client-go v0.27.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.27.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.27.1
	k8s.io/code-generator => k8s.io/code-generator v0.27.1
	k8s.io/component-base => k8s.io/component-base v0.27.1
	k8s.io/component-helpers => k8s.io/component-helpers v0.27.1
	k8s.io/controller-manager => k8s.io/controller-manager v0.27.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.27.1
	k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.27.1
	k8s.io/kms => k8s.io/kms v0.27.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.27.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.27.1
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.27.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.27.1
	k8s.io/kubectl => k8s.io/kubectl v0.27.1
	k8s.io/kubelet => k8s.io/kubelet v0.27.1
	k8s.io/kubernetes => k8s.io/kubernetes v1.27.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.27.1
	k8s.io/metrics => k8s.io/metrics v0.27.1
	k8s.io/mount-utils => k8s.io/mount-utils v0.27.1
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.27.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.27.1
)

require (
	github.com/containerd/cgroups v1.1.0
	github.com/openshift-kni/mixed-cpu-node-plugin v0.0.0-20230814112512-f1a527d0a451
	k8s.io/klog/v2 v2.100.1
	k8s.io/kubernetes v1.25.4
)

require (
	github.com/containerd/nri v0.2.0 // indirect
	github.com/containerd/ttrpc v1.1.1-0.20220420014843-944ef4a40df3 // indirect
	github.com/containers/podman/v4 v4.4.2 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/godbus/dbus/v5 v5.1.1-0.20221029134443-4b691ce883d5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/kubevirt/device-plugin-manager v1.19.4 // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20220909204839-494a5a6aca78 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	google.golang.org/genproto v0.0.0-20221227171554-f9683d7f8bef // indirect
	google.golang.org/grpc v1.51.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	k8s.io/cri-api v0.25.3 // indirect
	k8s.io/kubelet v0.26.0 // indirect
)
