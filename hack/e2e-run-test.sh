#!/usr/bin/env bash

E2E_SHARED_CPUS=${1}
E2E_SETUP=${2}
E2E_TEARDOWN=${3}

echo "Running e2e test"
E2E_SETUP=${E2E_SETUP} E2E_SHARED_CPUS=${E2E_SHARED_CPUS} E2E_TEARDOWN=${E2E_TEARDOWN} \
build/bin/e2e_test --ginkgo.v
