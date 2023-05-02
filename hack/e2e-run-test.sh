#!/usr/bin/env bash

E2E_SHARED_CPUS=${1}

export E2E_SHARED_CPUS=${E2E_SHARED_CPUS}

echo "Running e2e test"
build/bin/e2e_test --ginkgo.v