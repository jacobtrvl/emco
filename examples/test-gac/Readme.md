#### SPDX-License-Identifier: Apache-2.0
#### Copyright (c) 2022 Intel Corporation

Please see [generic-action-controller.md](../../docs/design/generic-action-controller.md) for more information about GAC implementation. This document explains the test cases for GAC.

#################################################################
# Running GAC test with emcoctl
#################################################################

## Setup Test Environment

1. Export environment variables 1) `KUBE_PATH` where the kubeconfig for edge cluster is located, 2) `CLUSTER_ISTIO_INGRESS_GATEWAY_ADDRESS` to the Istio ingress address of the edge cluster, and 3) `HOST_IP`: IP address of the cluster where EMCO is installed

## Test Kubernetes resource creation

Create prerequisites, create resources and customization in the GAC, and instantiate the GAC compositeApp deploymentIntentGroup

    ```shell
      $ ./apply.sh test-gac.yaml
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

Delete resources, prerequisites, and generated files

    ```shell
	  $ ./delete.sh test-gac.yaml
    ```

## Test JSON patch with an external lookup URL
Create prerequisites, create resources and customization in the GAC, and instantiate the GAC compositeApp deploymentIntentGroup

    ```shell
      $ ./apply.sh test-gac-patch-with-external-url.yaml
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

Delete resources, prerequisites, and generated files

    ```shell
      $ ./delete.sh test-gac-patch-with-external-url.yaml
    ```

## Test Strategic Merge

1. Create prerequisites, create resources and customization in the GAC, and instantiate the GAC compositeApp deploymentIntentGroup

    ```shell
      $ ./apply.sh test-strategic-merge.yaml
    ```

Once the deployment intent group is instantiated successfully, we can see the following resources on the edge cluster.

  1. `deployment`

        - deploy-web

      ```shell
        $ kubectl get deploy
        NAME              READY   UP-TO-DATE   AVAILABLE   AGE
        deploy-web        1/1     1            1           91s
      ```

      The deployment will have the new container, redis, on the list

      ```shell
        $ kubectl edit deploy deploy-web

          spec:
            template:
              spec:
                containers:
                - image: redis
                  imagePullPolicy: Always
                  name: redis-ctr
                  resources: {}
                  terminationMessagePath: /dev/termination-log
                  terminationMessagePolicy: File
      ```

  2. `statefulset`
  
        - etcd

      ```shell
        $ kubectl get statefulset
        NAME            READY   AGE
        etcd            3/3     7m25s
      ```

      The statefulset, etcd, will have the hostAliases details
    
      ```shell
        $ kubectl edit statefulset etcd

          spec:
            template:
              spec:
                hostAliases:
                - hostnames:
                  - host1
                  ip: 1.2.3.4
      ```

<b> Note: We are using the operator app in this example. You can create resources for different apps based on the use cases. <b>

### Cleanup

Delete resources, prerequisites, and generated files

    ```shell
      $ ./delete.sh test-strategic-merge.yaml
    ```
