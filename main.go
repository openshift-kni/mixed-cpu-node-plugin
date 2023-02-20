package main

import (
	"context"
	"flag"
	"os"

	"k8s.io/klog"

	"github.com/Tal-or/nri-mixed-cpu-pools-plugin/pkg/mixedcpus"
)

const ProgramName = "mixed-cpu-pool-plugin"

func main() {
	flags := flag.NewFlagSet(ProgramName, flag.ExitOnError)
	args, err := parseArgs(flags, os.Args[1:]...)
	if err != nil {
		klog.Fatalf("failed to parse arguments %v", err)
	}

	p, err := mixedcpus.New(args)
	if err != nil {
		klog.Fatalf("%v", err)
	}

	err = p.Stub.Run(context.Background())
	if err != nil {
		klog.Fatalf("plugin exited with error %v", err)
	}
}

func parseArgs(flags *flag.FlagSet, osArgs ...string) (*mixedcpus.Args, error) {
	args := &mixedcpus.Args{}
	flags.StringVar(&args.PluginName, "name", "", "plugin name to register to NRI")
	flags.StringVar(&args.PluginIdx, "idx", "", "plugin index to register to NRI")
	flags.StringVar(&args.ReservedCPUs, "reserved-cpus", "", "kubelet reserved cpus list ")
	klog.InitFlags(flags)

	return args, flags.Parse(osArgs)
}
