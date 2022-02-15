#### SPDX-License-Identifier: Apache-2.0
#### Copyright (c) 2022 Intel Corporation

Please see [generic-action-controller.md](../../docs/design/generic-action-controller.md) for more information about GAC implementation. 
This document explains the test cases for GAC.

#################################################################
# Running GAC test with emcoctl
#################################################################

## Setup Test Environment

1. Export environment variables 1) `KUBE_PATH` where the kubeconfig for edge cluster is located and 2) `HOST_IP`: IP address of the cluster where EMCO is installed

2. Setup script

    1. Run the setup script for creating artifacts needed to test GAC on a single cluster.

    `$ ./setup.sh create`

    The output of this command is
      1. `emco-cfg.yaml` for the current environment
      2. `prerequisites.yaml` for the test
      3. `values.yaml` for the current environment
      4. `helm chart and profiles tar.gz files` for all the use cases
    
    2. Cleanup artifacts generated above with the cleanup command

    `$ ./setup.sh cleanup`

## Running test
1. Create Prerequisites

    `$ $bin/emcoctl apply --config emco-cfg.yaml -v values.yaml -f prerequisites.yaml`

2. Create Resources

    `$ $bin/emcoctl apply --config emco-cfg.yaml -v values.yaml -f test-gac.yaml`

    Once the command runs successfully, we can see the following resources on the edge cluster.

    1. `networkpolicy`
        - sample-network-policy
        - test-network-policy

        ```shell
        $ kubectl get netpol
          NAME                    POD-SELECTOR    AGE
          sample-network-policy   role=database   5s
          test-network-policy     role=db         5s
        ```

    2. `configmap`
        - configmap-demo
        - test-configmap

        ```shell
        $ kubectl get cm
          NAME             DATA   AGE
          configmap-demo   5      5s
          test-configmap   4      5s
        ```

    3. `secret`
        - auth-secret
        - m3db-operator-token-bxt65
        - user-secret

        ```shell
        $ kubectl get secret
          NAME                        TYPE                                  DATA   AGE
          auth-secret                 kubernetes.io/service-account-token   7      5s
          m3db-operator-token-bxt65   kubernetes.io/service-account-token   3      5s
          user-secret                 kubernetes.io/service-account-token   7      5s
        ```

    4. `statefulset`
        - etcd
        - m3db-operator

        ```shell
        $ kubectl get statefulset 
          NAME            READY   AGE
          etcd            1/1     5s
          m3db-operator   1/1     5s
        ```

    <b> Note: We are using the operator app in this example. You can create resources for different apps based on the use cases. <b>

## Cleanup

1. Delete Resources 

    `$ $bin/emcoctl delete --config emco-cfg.yaml -v values.yaml -f test-gac.yaml`

    <b> Note: You cannot delete a record without deleting the dependent records (referential integrity). Please retry deleting the records if it fails. <b>

2. Delete Prerequisites

    `$ $bin/emcoctl delete --config emco-cfg.yaml -v values.yaml -f prerequisites.yaml`

    <b> Note: You cannot delete a record without deleting the dependent records (referential integrity). Please retry deleting the records if it fails. <b>

3. Cleanup generated files

    `$ ./setup.sh cleanup`
