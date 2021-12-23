```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2020-2021 Intel Corporation
```
<!-- omit in toc -->
# Sample application to demonstrate network policy of traffic controller 
This document describes how to deploy an example application with network policy of traffic controller. The deployment consists of server and client pods, once deployed the client sends the request to the server every five seconds. The client pod logs the message "Hello from http-server" to indicate successful connectivity with the server. 

- Requirements
- Install EMCO and emcoctl
- Build the app Images
- Prepare the edge cluster
- Configure
- Install the client/server applicationls
- Verify network policy resource instantiation
- Sample log from the client pod
- Uninstall the client/server application

## Requirements
- The edge cluster where the application is installed should support network policy

## Install EMCO and emcoctl
Install EMCO and emcoctl as described in the tutorial.

## Build the app Images
Build the http-server and http-client images. Refer to [this Readme](../../test-apps/README.md) for more details.

## Prepare the edge cluster
Install the Kubernetes edge cluster and make sure it supports network policy. Note down the kubeconfig for the edge cluster which is required later during configuration.

## Configure
(1) Copy the config file
```shell
$ cp src/tools/emcoctl/examples/emco-cfg-remote.yaml examples/dtc/network_policy/emco-cfg-dtc.yaml
```
(2) Modify examples/dtc/network_policy/emco-dtc-single-cluster.yaml and examples/dtc/network_policy/emco-cfg-dtc.yaml files to change host name, port number and kubeconfig path.

(3) Compress the profile and helm files

Update the profile files with right proxy address and create tar.gz of profiles
```shell
$ cd examples/helm_charts/http-server/profile/network_policy_overrides/http-server-profile
$ tar -czvf ../../../../../dtc/network_policy/http-server-profile.tar.gz .
$ cd ../../../../http-client/profile/network_policy_overrides/http-client-profile
$ tar -czvf ../../../../../dtc/network_policy/http-client-profile.tar.gz .
```
Create and copy .tgz of application helm charts
```shell
$ cd ../../../../http-server/helm
$ tar -czvf http-server.tgz http-server/
$ cp *.tgz ../../../dtc/network_policy/
$ cd ../../http-client/helm
$ tar -czvf http-client.tgz http-client/
$ cp *.tgz ../../../dtc/network_policy/
```

## Install the client/server app
Install the app using the commands:
```shell
$ cd ../../../dtc/network_policy/
$ emcoctl --config emco-cfg-dtc.yaml apply -f emco-dtc-single-cluster.yaml
$ emcoctl --config emco-cfg-dtc.yaml apply -f instantiate.yaml
```

## Verify network policy resource instantiation
```shell
$ kubectl get networkpolicy
  NAME               POD-SELECTOR      AGE
  testdtc-serverin   app=http-server   28s
```

## Sample log from the client pod

```shell
$ kubectl logs pod/r1-http-client-54568d6c9-ftmr7
get:
 2020-12-09 00:21:07 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
get:
 2020-12-09 00:21:12 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
get:
 2020-12-09 00:21:17 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
get:
 2020-12-09 00:21:22 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd 
```

## Uninstall the application
Uninstall the app using the commands:
```shell
$ emcoctl --config emco-cfg-dtc.yaml apply -f emco-dtc-terminate.yaml
$ emcoctl --config emco-cfg-dtc.yaml delete -f emco-dtc-single-cluster.yaml
```
