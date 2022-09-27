```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2021 Intel Corporation
```
<!-- omit in toc -->
# Sample application to demonstrate service discovery feature 
This document describes how to deploy an example application with network policy of traffic controller. The deployment consists of server and client pods, once deployed the client sends the request to the server every five seconds. The client pod logs the message "Hello from http-server" to indicate successful connectivity with the server. 

- Requirements
- Install EMCO and emcoctl
- Build the app Images
- Prepare the edge cluster
- Configure
- Install the client/server application
- Verify network policy resource instantiation
- Verify service entry resource instantiation
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

## Testing scenarios
NOTE: Public cloud scenarios are experimental and are not tested with public clouds

(1) Set the KUBE_PATH1 environment variable to the first private cluster kubeconfig file path.

(2) Set the HOST_IP environmet variables to the address where emco is running.

(3) Set the HTTP_SERVER_IMAGE_REPOSITORY and HTTP_CLIENT_IMAGE_REPOSITORY environment variable to the location of the http-server and http-client images.

(4) Modify examples/dtc/service_discovery/setup.sh files to change controller port numbers if they are diffrent than your emco installation.

### Communication between two private clusters (logical cloud level 0)
(1) Set the KUBE_PATH2 environment variable to the second private cluster kubeconfig file path.

(2) Set the LC_LEVEL environment variable to 0.

#### Install the client/server app
Install the app using the commands:
```shell
$ ./apply.sh
```

#### Verify network policy resource instantiation
```shell
$ kubectl get networkpolicy
NAME               POD-SELECTOR      AGE
testdtc-serverin   app=http-server   28s
```

#### Verify service entry created on the cluster where the client app is running
```shell
$ kubectl get svc
NAME                    TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
service/http-service    ClusterIP   10.233.0.1   <none>        443/TCP   1d
```

#### Sample log from the client pod
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

#### Uninstall the application
Uninstall the app using the commands:
```shell
$ ./delete.sh
```

### Communication between two private clusters (logical cloud level 1)
(1) Set the KUBE_PATH2 environment variable to the second private cluster kubeconfig file path.

(2) Set the LC_LEVEL environment variable to 1.

#### Install the client/server app
Install the app using the commands:
```shell
$ ./apply.sh
```

#### Verify network policy resource instantiation
```shell
$ kubectl -n ns1 get networkpolicy
NAME               POD-SELECTOR      AGE
testdtc-serverin   app=http-server   28s
```

#### Verify service entry created on the cluster where the client app is running
```shell
$ kubectl -n ns1 get svc
NAME                    TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
service/http-service    ClusterIP   10.233.0.1   <none>        443/TCP   1d
```

#### Sample log from the client pod
```shell
$ kubectl -n ns1 logs pod/r1-http-client-54568d6c9-ftmr7
get:
2020-12-09 00:21:07 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
get:
2020-12-09 00:21:12 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
get:
2020-12-09 00:21:17 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
get:
2020-12-09 00:21:22 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd 
```

#### Uninstall the application
Uninstall the app using the commands:
```shell
$ ./delete.sh
```

### Communication between a private cluster (client app) and public cluster (server app) (logical cloud level 0)
(1) Set the KUBE_PATH2 environment variable to the public cluster kubeconfig file path and the PUBLIC_CLUSTER2 environment variable to true.

(2) Set the LC_LEVEL environment variable to 0.

#### Install the client/server app
Install the app using the commands:
```shell
$ ./apply.sh
```

#### Verify network policy resource instantiation
```shell
$ kubectl get networkpolicy
NAME               POD-SELECTOR      AGE
testdtc-serverin   app=http-server   28s
```

#### Verify service entry created on the cluster where the client app is running
```shell
$ kubectl get svc
NAME                    TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
service/http-service    ClusterIP   10.233.0.1   <none>        443/TCP   1d
```

#### Sample log from the client pod
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

#### Uninstall the application
Uninstall the app using the commands:
```shell
$ ./delete.sh
```

### Communication between a private cluster (client app) and public cluster (server app) (logical cloud level 1)
(1) Set the KUBE_PATH2 environment variable to the public cluster kubeconfig file path and the PUBLIC_CLUSTER2 environment variable to true.

(2) Set the LC_LEVEL environment variable to 1.

#### Install the client/server app
Install the app using the commands:
```shell
$ ./apply.sh
```

#### Verify network policy resource instantiation
```shell
$ kubectl -n ns1 get networkpolicy
NAME               POD-SELECTOR      AGE
testdtc-serverin   app=http-server   28s
```

#### Verify service entry created on the cluster where the client app is running
```shell
$ kubectl -n ns1 get svc
NAME                    TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
service/http-service    ClusterIP   10.233.0.1   <none>        443/TCP   1d
```

#### Sample log from the client pod
```shell
$ kubectl -n ns1 logs pod/r1-http-client-54568d6c9-ftmr7
get:
2020-12-09 00:21:07 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
get:
2020-12-09 00:21:12 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
get:
2020-12-09 00:21:17 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
get:
2020-12-09 00:21:22 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd 
```

#### Uninstall the application
Uninstall the app using the commands:
```shell
$ ./delete.sh
```
