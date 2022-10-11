```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2020-2021 Intel Corporation
```
<!-- omit in toc -->
# Sample application to demonstrate istio traffic controller
This document describes how to deploy an example application with istio traffic controller. The deployment consists of server and client pods, running in two diffrent edge clusters. Once deployed, the client sends the request to the server every five seconds. The client pod logs the message "Hello from http-server" to indicate successful connectivity with the server.

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
- The edge clusters where the application is installed should support istio with single root CA

## Install EMCO and emcoctl
Install EMCO and emcoctl as described in the tutorial.

## Build the app Images
Build the http-server and http-client images. Refer to [this Readme](../../test-apps/README.md) for more details.

## Prepare the edge cluster
Install the Kubernetes edge clusters and make sure it supports istio with single root CA. Note down the kubeconfig for the edge cluster which is required later during configuration.

## Configure
(1) Set the KUBE_PATH1 and KUBE_PATH2 environment variables to cluster kubeconfig file path.  Set the CLUSTER2_ISTIO_INGRESS_GATEWAY_ADDRESS environment variables to reflect the Istio ingress address for the cluster.

(2) Set the HOST_IP enviromnet variables to the address where emco is running.

(3) Set the HTTP_SERVER_IMAGE_REPOSITORY and HTTP_CLIENT_IMAGE_REPOSITORY environment variable to the location of the http-server and http-client images.

(4) Modify examples/dtc/istio_traffic/setup.sh files to change controller port numbers if they are diffrent than your emco installation.

## Install the application
Install the app using the commands:
```shell
$ cd examples/dtc/istio_traffic/
$ ./apply.sh
```

## Verify resource instantiation
Run this command on the cluster where the server is installed:
```shell
$ kubectl get gw -n istio-system
NAME                   AGE
http-service-gateway   8h
```
```shell
kubectl get se -n istio-system
NAME                     HOSTS                                       LOCATION        RESOLUTION   AGE
http-service-se-server   [http-service.default.provider1.cluster2]   MESH_INTERNAL   DNS          8h
```
```shell
$ kubectl get dr
NAME                        HOST                                      AGE
http-service-se-server-dr   http-service.default.provider1.cluster2   8h
```
Run this command on the cluster where the client is installed:
```shell
$ kubectl get se
NAME                      HOSTS                                       LOCATION        RESOLUTION   AGE
http-service-se-client0   [http-service.default.provider1.cluster2]   MESH_INTERNAL   DNS          8h
http-service-se-client1   [http-service]                              MESH_INTERNAL   DNS          8h
```
```shell
$ kubectl get dr
NAME                             HOST                                      AGE
http-service-dr-client-svcname   http-service                              8h
http-service-dr-client0          http-service.default.provider1.cluster2   8h
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
Uninstall the app using the command:
```shell
$ ./delete.sh
```
