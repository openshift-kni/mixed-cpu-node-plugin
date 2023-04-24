package deviceplugin

import (
	"context"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/glog"
)

type pluginImp struct {
	mutualCpus *cpuset.CPUSet
}

func (p pluginImp) ListAndWatch(empty *pluginapi.Empty, server pluginapi.DevicePlugin_ListAndWatchServer) error {
	devs := MakeMutualCpusDevices()
	glog.V(4).Infof("ListAndWatch respond with: %+v", devs)
	if err := server.Send(&pluginapi.ListAndWatchResponse{Devices: devs}); err != nil {
		return err
	}
	// do not return, we need to keep the connection open
	select {}
}

func (p pluginImp) Allocate(ctx context.Context, request *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	return &pluginapi.AllocateResponse{
		ContainerResponses: []*pluginapi.ContainerAllocateResponse{
			{
				Envs: map[string]string{"OPENSHIFT_MUTUAL_CPUS": p.mutualCpus.String()},
			},
		},
	}, nil
}

func (p pluginImp) GetDevicePluginOptions(ctx context.Context, empty *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

func (p pluginImp) GetPreferredAllocation(ctx context.Context, request *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PreStartContainer not implemented")
}

func (p pluginImp) PreStartContainer(ctx context.Context, request *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PreStartContainer not implemented")
}
