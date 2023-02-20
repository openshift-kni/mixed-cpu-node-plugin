module github.com/Tal-or/nri-mixed-cpu-pools-plugin

go 1.19

require (
	github.com/containerd/nri v0.2.0
	github.com/containers/podman/v4 v4.4.1
	github.com/opencontainers/runc v1.1.4
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.26.1
)

require (
	github.com/cilium/ebpf v0.7.0 // indirect
	github.com/containerd/ttrpc v1.1.1-0.20220420014843-944ef4a40df3 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/cyphar/filepath-securejoin v0.2.3 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/godbus/dbus/v5 v5.1.1-0.20221029134443-4b691ce883d5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/moby/sys/mountinfo v0.6.2 // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20220825212826-86290f6a00fb // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	google.golang.org/genproto v0.0.0-20221227171554-f9683d7f8bef // indirect
	google.golang.org/grpc v1.51.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
)

// k8s deps
require (
	k8s.io/apimachinery v0.26.0
	k8s.io/cri-api v0.25.3 // indirect
	k8s.io/klog/v2 v2.90.0 // indirect

)
