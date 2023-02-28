package main

import (
	"context"
	"flag"
	
	"github.com/Tal-or/nri-mixed-cpu-pools-plugin/pkg/deviceplugin"
	"github.com/Tal-or/nri-mixed-cpu-pools-plugin/pkg/mixedcpus"
	"github.com/golang/glog"
)

func main() {
	args := parseArgs()
	p, err := mixedcpus.New(args)
	if err != nil {
		glog.Fatalf("%v", err)
	}

	go func() {
		err = p.Stub.Run(context.Background())
		if err != nil {
			glog.Fatalf("plugin exited with error %v", err)
		}
	}()

	dp, err := deviceplugin.New(args.MutualCPUs)
	if err != nil {
		glog.Fatalf("%v", err)
	}
	dp.Run()
}

func parseArgs() *mixedcpus.Args {
	args := &mixedcpus.Args{}
	flag.StringVar(&args.PluginName, "name", "", "plugin name to register to NRI")
	flag.StringVar(&args.PluginIdx, "idx", "", "plugin index to register to NRI")
	flag.StringVar(&args.MutualCPUs, "mutual-cpus", "", "mutual cpus list")
	flag.Parse()
	return args
}
