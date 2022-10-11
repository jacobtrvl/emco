[//]: # "SPDX-License-Identifier: Apache-2.0"
[//]: # "Copyright (c) 2020-2022 Intel Corporation"

# Running EMCO testcases with emcoctl

This folder contains the following test cases to run with EMCO. These tests assumes one edge cluster to run all test cases. EMCO needs to be installed and listening, before running these tests.

1. Prometheus and collectd Helm charts
2. vFirewall
3. DTC (Create client/server images using examples/test-apps/README.md)

## Setup Test Environment to run test cases

* In the ``config`` file, set the following variables (which are not set by default):
  1. ``KUBE_PATH``: points to where the kubeconfig for the edge cluster is located
  2. ``HOST_IP``: IP address of the cluster (or machine) where EMCO is installed
  3. ``DNS=true``: EMCO services running on kubernetes in default namespace can talk to the controller services internally when value is true.

---
* Additionally, you can optionally modify other variables:
  1. ``LOGICAL_CLOUD_LEVEL``: specifies the kind of Logical Cloud to use (choose between ``admin`` (default), ``privileged`` and ``standard``)
  2. the ports where each of the services run

-----
* For ``HOST_IP``, ``KUBE_PATH`` and ``LOGICAL_CLOUD_LEVEL``, you can also choose to **export** those variables instead of setting them in the ``config`` file. Exporting them takes priority over what's defined in the ``config`` file. Example below:

        export HOST_IP=127.0.0.1
        export KUBE_PATH=/root/clusters/k23-1.conf
        export LOGICAL_CLOUD_LEVEL
        ./setup.sh create


*NOTE 1: For ``HOST_IP``, assuming here that nodeports are used to access all EMCO services both from outside and between the EMCO services. Otherwise, if EMCO is running directly on baremetal, this will simply be the publicly-reachable address of that machine, or localhost for a local baremetal deployment.*
*NOTE 2: Relative directories and expansion of certain symbols, such as the tilde (`~`) for the user's home directory, do not work within the config file. Please make sure to specify absolute paths.*

---
* The setup.sh script

    Creates artifacts needed to test EMCO on one cluster. The script will read from the ``config`` file to decide what EMCO resources to create.

    ```
    $ ./setup.sh create
    ```

    Output files of this command are:
    * ``values.yaml``: specifies useful variables for the creation of EMCO resources
    * ``emco_cfg.yaml``: defines the deployment details of EMCO (IP addresses and ports of each service)
    * ``prerequisites.yaml``: defines all non usecase-specific EMCO resources to create
    * Helm charts and profile tarballs for all the usecases.

    ```
    $ ./setup.sh cleanup
    ```

    Cleans up all artifacts previously generated.

---
* ``instantiate-lc.yaml``: defines the API call that instantiates a Logical Cloud (required for any usecase)

## Applying prerequisites to run tests
Apply prerequisites.yaml. This is required for all the tests. This creates controllers, one project, one cluster, a logical cloud. This step is required to be done only once for all usecases:

```
$ emcoctl --config emco-cfg.yaml apply -f prerequisites.yaml -v values.yaml
```


## Instantiating Logical Cloud over the cluster

```
$ emcoctl --config emco-cfg.yaml apply -f instantiate-lc.yaml -v values.yaml
```

## Running test cases

1. Prometheus and collectd usecase

    ```
    $ emcoctl --config emco-cfg.yaml apply -f test-prometheus-collectd-deployment.yaml -v values.yaml

    $ emcoctl --config emco-cfg.yaml apply -f test-prometheus-collectd-instantiate.yaml -v values.yaml
    ```

2. vFirewall usecase

    ```
    $ emcoctl --config emco-cfg.yaml apply -f test-vfw-deployment.yaml -v values.yaml

    $ emcoctl --config emco-cfg.yaml apply -f test-vfw-instantiate.yaml -v values.yaml
    ```
    #### NOTE: This usecase is only tested using kubernetes installation: https://github.com/onap/multicloud-k8s/tree/master/kud, which comes with the requisite packages installed.
    #### For running vFw use case, the Kubernetes cluster needs to have following packages installed:
     multus - https://github.com/k8snetworkplumbingwg/multus-cni

     ovn4nfv - https://github.com/akraino-edge-stack/icn-ovn4nfv-k8s-network-controller/tree/master

     virtlet - https://github.com/Mirantis/virtlet

3. DTC testcase

    ```
    $ emcoctl --config emco-cfg.yaml apply -f test-dtc-deployment.yaml -v values.yaml

    $ emcoctl --config emco-cfg.yaml apply -f test-dtc-instantiate.yaml -v values.yaml
    ```

4. Installing Monitor on edge cluster

    ```
    $ emcoctl --config emco-cfg.yaml apply -f monitor-deployment.yaml -v values.yaml

    $ emcoctl --config emco-cfg.yaml apply -f monitor-instantiate.yaml -v values.yaml
    ```

## Cleanup

1. Delete Prometheus and Collectd usecase

    ```
    $ emcoctl --config emco-cfg.yaml delete -f test-prometheus-collectd-instantiate.yaml -v values.yaml

    $ emcoctl --config emco-cfg.yaml delete -f test-prometheus-collectd-deployment.yaml -v values.yaml
    ```

2. Firewall testcase

    ```
    $ emcoctl --config emco-cfg.yaml delete -f test-vfw-instantiate.yaml -v values.yaml

    $ emcoctl --config emco-cfg.yaml delete -f test-vfw-deployment.yaml -v values.yaml
    ```

3. DTC testcase

    ```
    $ emcoctl --config emco-cfg.yaml delete -f test-dtc-instantiate.yaml -v values.yaml

    $ emcoctl --config emco-cfg.yaml delete -f test-dtc-deployment.yaml -v values.yaml
    ```

4. Terminate Logical Cloud

    ```
    $ emcoctl --config emco-cfg.yaml delete -f instantiate-lc.yaml -v values.yaml
    ```

5. Cleanup prerequisites

    ```
    $ emcoctl --config emco-cfg.yaml delete -f prerequisites.yaml -v values.yaml
    ```

6. Cleanup generated files

    ```
    $ ./setup.sh cleanup
    ```

*NOTE: Known issue with the test cases: deletion of the resources fails sometimes as some resources can't be deleted before others are deleted. This can happen due to timing issue. In that case try deleting again and the deletion should succeed.*

## Running EMCO testcases using test-aio.sh script

The test-aio.sh script can be used to simplify the process of running a test case. It makes use of the status query APIs to ensure that the logical cloud and deployment intent group are instantiated (During creation) or terminated (During deletion) before moving to the next step.

To run the test cases run the following commands after setting up the environment variables. There is no need to run the setup.sh script as it's taken care of by the test-aio.sh script.

1. For running the test case

    ```
    $ ./test-aio.sh <test case name> apply
    ```

2. For deleting the test case and cleaning up resources

    ```
    $ ./test-aio.sh <test case name> delete
    ```

For example to run the dtc test case

1. For running the dtc test case

    ```
    $ ./test-aio.sh test-dtc apply
    ```

2. For deleting the dtc test case and cleaning up resources

    ```
    $ ./test-aio.sh test-dtc delete
    ```
