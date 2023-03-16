# nri-mixed-cpu-pools-plugin
nri:https://github.com/containerd/nri plugin implements both exclusive and mutual cpus request for containers
deployment on top of Kubernetes and OpenShift platforms.

In addition, it implements mutual-cpu as device via device plugin.
This way standardize the way of how container workloads should asks for mutual cpus.  

 - POC - DONE
 - Deployment - DONE
 - UpdateContainer flow - DONE
 - Support a case when plugin deployed after app container - DONE
 - Handle reboot/restart node/kubelet/crio/pod flow - DONE
 - Unit Tests - TODO
 - E2E Tests - TODO
 - Support cgroupfs - TODO

![](/home/titzhak/go/code/github.com/Tal-or/nri-mixed-cpu-pools-plugin/docs/MixedCPUSWorkloadsFlow.png)