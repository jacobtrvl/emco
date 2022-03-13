#### SPDX-License-Identifier: Apache-2.0
#### Copyright (c) 2022 Intel Corporation

Please see [generic-action-controller.md](../../docs/design/generic-action-controller.md) for more information about GAC implementation. This document explains the test cases for GAC.

#################################################################
# Running GAC test with emcoctl
#################################################################

## Setup Test Environment

1. Export environment variables 1) `KUBE_PATH` where the kubeconfig for edge cluster is located and 2) `HOST_IP`: IP address of the cluster where EMCO is installed

2. Setup script

    1. Run the setup script for creating artifacts needed to test GAC on a single cluster

          ```shell
            $ sudo chmod +x setup.sh

            $ ./setup.sh create
          ```

      The output of this command is
        1. `emco-cfg.yaml` for the current environment
        2. `prerequisites.yaml` for the test
        3. `values.yaml` for the current environment
        4. `instantiate.yaml` to instantiate the deployment intent group
        5. `update.yaml` to update the deployment intent group
        6. `rollback.yaml` to rollback to a previous version
        7. `helm chart and profiles tar.gz files` for all the use cases
      
    2. Cleanup artifacts generated above with the cleanup command

          ```shell
            $ ./setup.sh cleanup
          ```

## Test Kubernetes resource creation

1. Create Prerequisites

    ```shell
      $ $bin/emcoctl apply --config emco-cfg.yaml -v values.yaml -f prerequisites.yaml
    ```

2. Create Resources and Customization in the GAC

    ```shell
      $ $bin/emcoctl apply --config emco-cfg.yaml -v values.yaml -f test-gac.yaml
    ```

3. Instantiate the GAC compositeApp deploymentIntentGroup

    ```shell
      $ $bin/emcoctl apply --config emco-cfg.yaml -v values.yaml -f instantiate.yaml
    ```

Once the deploymentIntentGroup is instantiated successfully, we can see the following resources on the edge cluster.

  1. `networkpolicy`

        - netpol-db
        - netpol-web

      ```shell
        $ kubectl get netpol
        NAME                    POD-SELECTOR    AGE
        netpol-db               role=database   5s
        netpol-web              app=emco        5s
      ```

  2. `configmap`

        - cm-game
        - cm-team

      ```shell
        $ kubectl get cm
        NAME             DATA   AGE
        cm-game          5      5s
        cm-team          4      5s
      ```

  3. `secret`

        - m3db-operator-token-bxt65
        - secret-auth
        - secret-user

      ```shell
          $ kubectl get secret
          NAME                        TYPE                                  DATA   AGE
          m3db-operator-token-bxt65   kubernetes.io/service-account-token   3      5s
          secret-auth                 kubernetes.io/service-account-token   7      5s
          secret-user                 kubernetes.io/service-account-token   7      5s
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

## Test Kubernetes resource update/rollback

  You can create/update Kubernetes objects after instantiating a deployment intent. For example, if you want to update the replica count of the `etcd` statefulset, follow the below steps.

  1. Update customization

      ```shell
        $ $bin/emcoctl update --config emco-cfg.yaml -v values.yaml -f test-gac-update.yaml
      ```

  2. Update GAC compositeApp deploymentIntentGroup

      ```shell
        $ $bin/emcoctl apply --config emco-cfg.yaml -v values.yaml -f update.yaml
      ```

  Once the deployment is updated successfully, we can see that the replica count of `etcd` statefulset is now two.

  ```shell
    $ kubectl get statefulset
    NAME            READY   AGE
    etcd            2/2     2m
  ```

  You can also rollback the statefulset state to a previous version. For example, if you want to roll back the above changes,

  ```shell
    $ $bin/emcoctl apply --config emco-cfg.yaml -v values.yaml -f rollback.yaml
  ```

  Once the deployment is updated successfully, we can see that the replica count of `etcd` statefulset is now back to one.

  ```shell
    $ kubectl get statefulset
    NAME            READY   AGE
    etcd            1/1     3m
  ```

### Cleanup

1. Delete Resources 

    ```shell
      $ $bin/emcoctl delete --config emco-cfg.yaml -v values.yaml -f instantiate.yaml

      $ $bin/emcoctl delete --config emco-cfg.yaml -v values.yaml -f test-gac.yaml
    ```

    <b> Note: You cannot delete a record without deleting the dependent records (referential integrity). Please retry deleting the records if it fails. <b>

2. Delete Prerequisites

    ```shell
      $ $bin/emcoctl delete --config emco-cfg.yaml -v values.yaml -f prerequisites.yaml
    ```

    <b> Note: You cannot delete a record without deleting the dependent records (referential integrity). Please retry deleting the records if it fails. <b>

3. Cleanup generated files

    ```shell
      $ ./setup.sh cleanup
    ```

## Test JSON patch with an external lookup URL
1. Create Prerequisites

    ```shell
      $ $bin/emcoctl apply --config emco-cfg.yaml -v values.yaml -f prerequisites.yaml
    ```

2. Create Resources and Customization in the GAC

    ```shell
      $ $bin/emcoctl apply --config emco-cfg.yaml -v values.yaml -f test-gac-patch-with-external-url.yaml
    ```

3. Instantiate the GAC compositeApp deploymentIntentGroup

    ```shell
      $ $bin/emcoctl apply --config emco-cfg.yaml -v values.yaml -f instantiate.yaml
    ```

Once the deployment intent group is instantiated successfully, we can see the following resources on the edge cluster.

  1. `configmap`

        - cm-istio

      ```shell
        $ kubectl get cm
        NAME             DATA   AGE
        cm-istio         4      5s
      ```

  2. `secret`
  
        - m3db-operator-token-bxt65
        - secret-db

      ```shell
        $ kubectl get secret
        NAME                        TYPE                                  DATA   AGE
        m3db-operator-token-bxt65   kubernetes.io/service-account-token   3      5s
        secret-db                   kubernetes.io/service-account-token   6      5s
      ```

<b> Note: We are using the operator app in this example. You can create resources for different apps based on the use cases. <b>

### Cleanup

1. Delete Resources 

    ```shell
      $ $bin/emcoctl delete --config emco-cfg.yaml -v values.yaml -f instantiate.yaml

      $ $bin/emcoctl delete --config emco-cfg.yaml -v values.yaml -f test-gac-patch-with-external-url.yaml
    ```

    <b> Note: You cannot delete a record without deleting the dependent records (referential integrity). Please retry deleting the records if it fails. <b>

2. Delete Prerequisites

    ```shell
      $ $bin/emcoctl delete --config emco-cfg.yaml -v values.yaml -f prerequisites.yaml
    ```

    <b> Note: You cannot delete a record without deleting the dependent records (referential integrity). Please retry deleting the records if it fails. <b>

3. Cleanup generated files

    ```shell
      $ ./setup.sh cleanup
    ```
