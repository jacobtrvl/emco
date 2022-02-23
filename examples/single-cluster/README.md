[//]: # "SPDX-License-Identifier: Apache-2.0"
[//]: # "Copyright (c) 2020-2022 Intel Corporation"

# Running EMCO testcases with emcoctl

This folder contains the following test cases to run with EMCO. These tests assumes one edge cluster to run all test cases. EMCO needs to be installed and listening, before running these tests.

1. Prometheus and collectd Helm charts
2. vFirewall
3. collectd Helm chart and adding configmap during instantiation (using Generic Action Controller)
4. DTC (Create client/server images using examples/test-apps/README.md)

## Setup Test Environment to run test cases

* In the ``config`` file, set the following variables (which are not set by default):
  1. ``KUBE_PATH``: points to where the kubeconfig for the edge cluster is located
  2. ``HOST_IP``: IP address of the cluster (or machine) where EMCO is installed

* Additionally, you can optionally modify other variables:
  1. ``LOGICAL_CLOUD_LEVEL``: specifies the kind of Logical Cloud to use (choose between ``admin`` (default), ``privileged`` and ``standard``)
  2. the ports where each of the services run

* For ``HOST_IP``, ``KUBE_PATH`` and ``LOGICAL_CLOUD_LEVEL``, you can also choose to **export** those variables instead of setting them in the ``config`` file. Exporting them takes priority over what's defined in the ``config`` file. Example below:

        export HOST_IP=127.0.0.1
        export KUBE_PATH=/root/clusters/k23-1.conf
        export LOGICAL_CLOUD_LEVEL
        ./setup.sh create


*NOTE 1: For ``HOST_IP``, assuming here that nodeports are used to access all EMCO services both from outside and between the EMCO services. Otherwise, if EMCO is running directly on baremetal, this will simply be the publicly-reachable address of that machine, or localhost for a local baremetal deployment.*
*NOTE 2: Relative directories and expansion of certain symbols, such as the tilde (`~`) for the user's home directory, do not work within the config file. Please make sure to specify absolute paths.*

* The setup.sh script

    Creates artifacts needed to test EMCO on one cluster. The script will read from the ``config`` file to decide what EMCO resources to create.

    `$ ./setup.sh create`

    Output files of this command are:
    * ``values.yaml``: specifies useful variables for the creation of EMCO resources
    * ``emco_cfg.yaml``: defines the deployment details of EMCO (IP addresses and ports of each service)
    * ``prerequisites.yaml``: defines all non usecase-specific EMCO resources to create
    * Helm charts and profile tarballs for all the usecases.

    `$ ./setup.sh cleanup`

    Cleans up all artifacts previously generated.


* ``instantiate-lc.yaml``: defines the API call that instantiates a Logical Cloud (required for any usecase)

## Applying prerequisites to run tests
Apply prerequisites.yaml. This is required for all the tests. This creates controllers, one project, one cluster, a logical cloud. This step is required to be done only once for all usecases:

``$ $bin/emcoctl --config emco-cfg.yaml apply -f prerequisites.yaml -v values.yaml``

*NOTE: ``$bin`` refers to the path where the emcoctl binary is installed. This variable is not defined by default, so you need to adapt the command to your environment.*

## Instantiating Logical Cloud over the cluster

``$ $bin/emcoctl --config emco-cfg.yaml apply -f instantiate-lc.yaml -v values.yaml``

## Running test cases

1. Prometheus and collectd usecase

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f test-prometheus-collectd.yaml -v values.yaml`

2. Generic action controller usecase

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f test-gac.yaml -v values.yaml`

3. vFirewall usecase

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f test-vfw.yaml -v values.yaml`
    #### NOTE: This usecase is only tested using kubernetes installation: https://github.com/onap/multicloud-k8s/tree/master/kud, which comes with the requisite packages installed.
    #### For running vFw use case, the Kubernetes cluster needs to have following packages installed:
     multus - https://github.com/k8snetworkplumbingwg/multus-cni

     ovn4nfv - https://github.com/akraino-edge-stack/icn-ovn4nfv-k8s-network-controller/tree/master

     virtlet - https://github.com/Mirantis/virtlet

4. DTC testcase

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f test-dtc.yaml -v values.yaml`

5. Installing Monitor on edge cluster

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f monitor.yaml -v values.yaml`

## Cleanup

1. Delete Prometheus and Collectd usecase

    `$ $bin/emcoctl --config emco-cfg.yaml delete -f test-prometheus-collectd.yaml -v values.yaml`

2. Delete Generic action controller testcase

    `$ $bin/emcoctl --config emco-cfg.yaml delete -f test-gac.yaml -v values.yaml`

3. Firewall testcase

    `$ $bin/emcoctl --config emco-cfg.yaml delete -f test-vfw.yaml -v values.yaml`

4. DTC testcase

    `$ $bin/emcoctl --config emco-cfg.yaml delete -f test-dtc.yaml -v values.yaml`

5. Terminate Logical Cloud

    `$ $bin/emcoctl --config emco-cfg.yaml delete -f instantiate-lc.yaml -v values.yaml`

6. Cleanup prerequisites

    `$ $bin/emcoctl --config emco-cfg.yaml delete -f prerequisites.yaml -v values.yaml`

7. Cleanup generated files

    `$ ./setup.sh cleanup`

*NOTE: Known issue with the test cases: deletion of the resources fails sometimes as some resources can't be deleted before others are deleted. This can happen due to timing issue. In that case try deleting again and the deletion should succeed.*