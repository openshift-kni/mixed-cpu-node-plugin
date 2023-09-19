package config

import "os"

const defaultImageName = "registry.ci.openshift.org/ocp-kni/mixed-cpu-node-plugin:mixed-cpu-node-plugin"

func SharedCPUs() string {
	cpus, ok := os.LookupEnv("E2E_SHARED_CPUS")
	if !ok {
		return ""
	}
	return cpus
}

func Image() string {
	image, ok := os.LookupEnv("E2E_IMAGE_NAME")
	if !ok {
		return defaultImageName
	}
	return image
}
