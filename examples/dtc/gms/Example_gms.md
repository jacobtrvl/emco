```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2020-2021 Intel Corporation
```
<!-- omit in toc -->
# GMS application to demonstrate istio traffic controller
This document describes how to deploy an example application with istio traffic controller. The deployment consists of several microservices, running in two different edge clusters. Once deployed, the clients in one cluster sends the request to the productcatalog, cart and shipping service in different cluster. Cart service in turn makes request to redis cache running alongside in the same cluster. The curl output shows the products through frontend service to indicate successful connectivity.

- Requirements
- Install EMCO and emcoctl
- Prepare the edge cluster
- Configure
- Install the application
- Verify resource instantiation
- Verify the connectivity
- Uninstall the application

## Requirements
- The edge clusters where the application is installed should support istio with single root CA

## Install EMCO and emcoctl
Install EMCO and emcoctl as described in the tutorial.

## Prepare the edge cluster
Install the Kubernetes edge clusters and make sure it supports istio with single root CA. Note down the kubeconfig for the edge cluster which is required later during configuration.

## Configure
(1) Set the KUBE_PATH1 and KUBE_PATH2 enviromnet variables to cluster kubeconfig file path.

(2) Set the HOST_IP enviromnet variables to the address where emco is running.

(3) Modify examples/dtc/gms/setup.sh files to change controller port numbers if they are diffrent than your emco installation.

(4) Modify examples/dtc/gms/clusters.yaml to reflect istio ingress address for both the clusters.

## Install the application
Install the app using the commands:
```shell
$ cd examples/dtc/gms/
$ ./apply.sh
```

## Verify resource instantiation
Run this command on the cluster where the productcatalogservice is installed:
```shell
$ kubectl get gw -n istio-system
NAME                            AGE
cartservice-gateway             23s
productcatalogservice-gateway   22s
shippingservice-gateway         22s
```
```shell
kubectl get se -n istio-system
NAME                              HOSTS                                                      LOCATION        RESOLUTION   AGE
cartservice-se-server             [cartservice.default.gmsprovider1.gmscluster2]             MESH_INTERNAL   DNS          59s
productcatalogservice-se-server   [productcatalogservice.default.gmsprovider1.gmscluster2]   MESH_INTERNAL   DNS          59s
shippingservice-se-server         [shippingservice.default.gmsprovider1.gmscluster2]         MESH_INTERNAL   DNS          59s
```
```shell
$ kubectl get dr -n istio-system
NAME                                 HOST                                                     AGE
cartservice-se-server-dr             cartservice.default.gmsprovider1.gmscluster2             111s
productcatalogservice-se-server-dr   productcatalogservice.default.gmsprovider1.gmscluster2   111s
shippingservice-se-server-dr         shippingservice.default.gmsprovider1.gmscluster2         111s
```
Run this command on the cluster where the frontend is installed:

```shell
$ kubectl get se | grep productcatalogservice
NAME                               HOSTS                                                LOCATION        RESOLUTION   AGE
productcatalogservice-se-client0   [productcatalogservice.default.provider1.cluster2]   MESH_INTERNAL   DNS          19h
productcatalogservice-se-client1   [productcatalogservice]                              MESH_INTERNAL   DNS          19h
```
```shell
$ kubectl get dr | grep productcatalogservice
NAME                                      HOST                                               AGE
productcatalogservice-dr-client-svcname   productcatalogservice                              19h
productcatalogservice-dr-client0          productcatalogservice.default.provider1.cluster2   19h
```

## Verify the connectivity

```shell
Get the external ip to access the fronend:
$ kubectl get svc frontend-external
NAME                TYPE           CLUSTER-IP    EXTERNAL-IP      PORT(S)        AGE
frontend-external   LoadBalancer   10.244.4.61   192.168.121.10   80:31420/TCP   20h
```

Curl the external ip and make sure products are present in the output:
```shell
$ curl 192.168.121.10 | grep product
0          <div class="col-md-4 hot-product-card">
              <a href="/product/OLJCESPC7Z">
 1              <img alt="" src="/static/img/products/sunglasses.jpg">
944      0 --:--:--  0:00:04 --:--:--  1945              <div class="hot-product-card-img-overlay"></div>
              <div class="hot-product-card-name">Sunglasses</div>
10              <div class="hot-product-card-price">$19.99</div>
0           <div class="col-md-4 hot-product-card">
             <a href="/product/66VCHSJNUP">
9              <img alt="" src="/static/img/products/tank-top.jpg">
8              <div class="hot-product-card-img-overlay"></div>
7              <div class="hot-product-card-name">Tank Top</div>
0               <div class="hot-product-card-price">$18.99</div>
           <div class="col-md-4 hot-product-card">
              <a href="/product/1YMWWN1N4O">
0              <img alt="" src="/static/img/products/watch.jpg">
               <div class="hot-product-card-img-overlay"></div>
               <div class="hot-product-card-name">Watch</div>
9              <div class="hot-product-card-price">$109.99</div>
8          <div class="col-md-4 hot-product-card">
7            <a href="/product/L9ECAV7KIM">
0              <img alt="" src="/static/img/products/loafers.jpg">
               <div class="hot-product-card-img-overlay"></div>
               <div class="hot-product-card-name">Loafers</div>
               <div class="hot-product-card-price">$89.99</div>
           <div class="col-md-4 hot-product-card">
0            <a href="/product/2ZYFJ3GM2N">
               <img alt="" src="/static/img/products/hairdryer.jpg">
               <div class="hot-product-card-img-overlay"></div>
               <div class="hot-product-card-name">Hairdryer</div>
               <div class="hot-product-card-price">$24.99</div>
            <p class="footer-text">This website is hosted for demo purposes only. It is not an actual shop. This is not a Google product.</p>
            <p class="footer-text">This website is hosted for demo purposes only. It is not an actual shop. This is not a Google product.</p>
```

## Uninstall the application
Uninstall the app using the command:
```shell
$ ./delete.sh 
```
