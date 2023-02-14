package main

import (
	"context"
	"flag"
	"log"
	"os"

	"k8s.io/klog"

	"github.com/containerd/nri/pkg/api"
	"github.com/containerd/nri/pkg/stub"
)

// our injector plugin
type plugin struct {
	stub stub.Stub
}

func main() {
	var (
		pluginName string
		pluginIdx  string
		verbose    bool
		opts       []stub.Option
		err        error
	)

	flag.StringVar(&pluginName, "name", "", "plugin name to register to NRI")
	flag.StringVar(&pluginIdx, "idx", "", "plugin index to register to NRI")
	flag.BoolVar(&verbose, "verbose", false, "enable (more) verbose logging")
	flag.Parse()

	if pluginName != "" {
		opts = append(opts, stub.WithPluginName(pluginName))
	}
	if pluginIdx != "" {
		opts = append(opts, stub.WithPluginIdx(pluginIdx))
	}

	p := &plugin{}
	if p.stub, err = stub.New(p, opts...); err != nil {
		log.Fatalf("failed to create plugin stub: %v", err)
	}

	err = p.stub.Run(context.Background())
	if err != nil {
		klog.Errorf("plugin exited with error %v", err)
		os.Exit(1)
	}

}

// CreateContainer handles container creation requests.
func (p *plugin) CreateContainer(pod *api.PodSandbox, ctr *api.Container) (*api.ContainerAdjustment, []*api.ContainerUpdate, error) {
	klog.Infof("Creating container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())
	res := api.ContainerAdjustment{}
	return &res, nil, nil
}

func (p *plugin) UpdateContainer(pod *api.PodSandbox, ctr *api.Container) ([]*api.ContainerUpdate, error) {
	klog.Infof("Updating container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())
	// do nothing but store the original resources
	// so CRIO code won't crash with nil pointer
	res := api.ContainerUpdate{
		//ContainerId: ctr.Id,
		Linux: &api.LinuxContainerUpdate{
			Resources: ctr.Linux.Resources,
		},
	}
	return []*api.ContainerUpdate{&res}, nil
}

func (p *plugin) Configure(config, runtime, version string) (api.EventMask, error) {

}
