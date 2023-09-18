package main

import (
	"bytes"
	"errors"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"

	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"

	"github.com/openshift-kni/mixed-cpu-node-plugin/pkg/deviceplugin"
)

func main() {
	klog.Infof("reading /sys/fs/cgroup/cpuset/cpuset.cpus to retrieve the accessible cpus")
	out, err := os.ReadFile("/sys/fs/cgroup/cpuset/cpuset.cpus")
	if err != nil {
		klog.Fatal(err)
	}
	cgroupCPUs := strings.TrimSuffix(string(out), "\n")
	completeSet, err := cpuset.Parse(cgroupCPUs)
	if err != nil {
		klog.Fatalf("failed to parse cpuset %q; %v", cgroupCPUs, err)
	}
	klog.Infof("/sys/fs/cgroup/cpuset/cpuset.cpus content: %s", cgroupCPUs)

	klog.Infof("reading environment variable:%q to retrieve the shared cpus", deviceplugin.EnvVarName)
	cpus, ok := os.LookupEnv(deviceplugin.EnvVarName)
	if !ok {
		klog.Fatalf("%q environment variable not set", deviceplugin.EnvVarName)
	}
	sharedSet, err := cpuset.Parse(cpus)
	if err != nil {
		klog.Fatalf("failed to parse cpuset %q; %v", cpus, err)
	}
	if sharedSet.IsEmpty() {
		klog.Warning("no shared cpus are configured for this process")
	}
	klog.Infof("%s=%s", deviceplugin.EnvVarName, cpus)
	// here we extract the shared cpus from the complete set and find the isolated set
	isolatedSet := completeSet.Difference(sharedSet)
	klog.Infof("container finalized cpuset layout:\ncomplete-set=%q\nisolated-set=%q\nshared-set=%q", completeSet.String(), isolatedSet.String(), sharedSet.String())
	var wg sync.WaitGroup
	spawnLightWeightTasks(&sharedSet, &wg, 2)
	spawnHeavyWeightTasks(&isolatedSet, &wg, 2)
	wg.Wait()
}

func spawnLightWeightTasks(set *cpuset.CPUSet, wg *sync.WaitGroup, tasks int) {
	for i := 0; i < tasks; i++ {
		spawnTask(set, wg, "light weight task")
	}
}

func spawnHeavyWeightTasks(set *cpuset.CPUSet, wg *sync.WaitGroup, tasks int) {
	for i := 0; i < tasks; i++ {
		spawnTask(set, wg, "heavy weight task")
	}
}

func spawnTask(set *cpuset.CPUSet, wg *sync.WaitGroup, desc string) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		unixSet := k8sCPUstoUnixCPUs(set)
		tid := syscall.Gettid()
		id, err := goid()
		if err != nil {
			klog.Fatal(err)
		}
		if err := unix.SchedSetaffinity(0, unixSet); err != nil {
			klog.Fatal(err)
		}
		for {
			klog.Infof("%s: thread id %d => goroutine id: %d set affinity to cores: %q", desc, tid, id, set.String())
			time.Sleep(60 * time.Second)
		}
	}()
}

func k8sCPUstoUnixCPUs(set *cpuset.CPUSet) *unix.CPUSet {
	unixSet := &unix.CPUSet{}
	for _, i := range set.List() {
		unixSet.Set(i)
	}
	return unixSet
}

var (
	goroutinePrefix = []byte("goroutine ")
	errBadStack     = errors.New("invalid runtime.Stack output")
)

// This is terrible, slow, and should never be used in production.
func goid() (int, error) {
	buf := make([]byte, 32)
	n := runtime.Stack(buf, false)
	buf = buf[:n]
	// goroutine 1 [running]: ...
	buf, ok := bytes.CutPrefix(buf, goroutinePrefix)
	if !ok {
		return 0, errBadStack
	}

	i := bytes.IndexByte(buf, ' ')
	if i < 0 {
		return 0, errBadStack
	}
	return strconv.Atoi(string(buf[:i]))
}
