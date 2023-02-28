package main

import (
	"context"
	"flag"

	"github.com/golang/glog"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"

	"github.com/Tal-or/nri-mixed-cpu-pools-plugin/pkg/deviceplugin"
	"github.com/Tal-or/nri-mixed-cpu-pools-plugin/pkg/mixedcpus"
)

func main() {
	args := parseArgs()
	p, err := mixedcpus.New(args)
	if err != nil {
		glog.Fatalf("%v", err)
	}

	dp, err := deviceplugin.New(args.MutualCPUs)
	if err != nil {
		glog.Fatalf("%v", err)
	}

	execute(p, dp)
}

func parseArgs() *mixedcpus.Args {
	args := &mixedcpus.Args{}
	flag.StringVar(&args.PluginName, "name", "", "plugin name to register to NRI")
	flag.StringVar(&args.PluginIdx, "idx", "", "plugin index to register to NRI")
	flag.StringVar(&args.MutualCPUs, "mutual-cpus", "", "mutual cpus list")
	flag.Parse()
	return args
}

func execute(p *mixedcpus.Plugin, dp *dpm.Manager) {
	go func() {
		err := p.Stub.Run(context.Background())
		if err != nil {
			glog.Fatalf("plugin exited with error %v", err)
		}
	}()

	dp.Run()
}
