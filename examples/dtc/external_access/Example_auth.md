```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2022 Intel Corporation
```
<!-- omit in toc -->
# GMS application to demonstrate istio traffic controller
This document describes how to deploy an example application with istio traffic controller. The deployment consists of httpbin and sleep microservices, running in two different edge clusters. Once deployed, httpbin service can be accessed from the outside cluster as well as from the sleep service from the other cluster. The auth policy allows the service to be accessed through GET method and "/status/" path only. The curl output shows the succesful mutual handshake details (In case of mutual tls) as well as connection to the service. 

- Requirements
- Install EMCO and emcoctl
- Build the app Images
- Prepare the edge cluster
- Configure
- Install the application
- Verify resource instantiation
- Verify the connectivity using mutual tls
- Verify the connectivity using http
- Verify the connectivity from the sleep pod
- Uninstall the application

## Requirements
- The edge clusters where the application is installed should support istio with single root CA and sidecar injection enabled.

## Install EMCO and emcoctl
Install EMCO and emcoctl as described in the tutorial.

## Build the app Images
Build the httpbin-client images. Refer to [this Readme](../../test-apps/README.md) for more details. Update the examples/helm_charts/httpbin-client/helm/sleep/values.yaml with the image repo.

## Prepare the edge cluster
Install the Kubernetes edge clusters and make sure it supports istio with single root CA. Note down the kubeconfig for the edge cluster which is required later during configuration.

## Configure
(1) Set the KUBE_PATH1 and KUBE_PATH2 enviromnet variables to cluster kubeconfig file path.

(2) Set the HOST_IP enviromnet variables to the address where emco is running.

(3) Generate certs and keys 
```shell
cd examples/certs/httpbin/server
./gen-server-certs.sh
./base64.sh
```
Use the output from the above command and update the examples/dtc/external_access/setup.sh files with ca.crt, tls.crt and tls.key
```shell
cd ../client
./gen-client-certs.sh
```
(4) Modify examples/dtc/external_access/setup.sh files to change controller port numbers if they are diffrent than your emco installation.

(5) Modify examples/dtc/external_access/clusters.yaml to reflect istio ingress address for both the clusters.

## Install the application
Install the app using the commands:
```shell
$ cd examples/dtc/external_access/
$ ./apply.sh
```

## Verify resource instantiation
Run this command on the cluster where the httpbin is installed:
```shell
$ kubectl get AuthorizationPolicy -A
NAMESPACE    NAME      AGE
httpbin-ns   httpbin   73s
```
```shell
$ kubectl get gw -A
NAMESPACE      NAME                         AGE
httpbin-ns     httpbin-ext-http-gateway     2m22s
httpbin-ns     httpbin-ext-mutual-gateway   2m24s
istio-system   httpbin-gateway              2m23s
```
```shell
$ kubectl get vs -A
NAMESPACE    NAME                                    GATEWAYS                         HOSTS                     AGE
httpbin-ns   httpbin-vs-httpbin-ext-http-gateway     ["httpbin-ext-http-gateway"]     ["httpbin.example.com"]   3m54s
httpbin-ns   httpbin-vs-httpbin-ext-mutual-gateway   ["httpbin-ext-mutual-gateway"]   ["httpbin.example.com"]   3m55s
```
## Verify the connectivity using mutual tls

```shell
$ cd examples/certs/httpbin/client
$ curl-mutual.sh
```

Output:
```shell
* Added httpbin.example.com:30830:172.16.16.200 to DNS cache
* Hostname httpbin.example.com was found in DNS cache
*   Trying 172.16.16.200...
* TCP_NODELAY set
* Connected to httpbin.example.com (172.16.16.200) port 30830 (#0)
* ALPN, offering h2
* ALPN, offering http/1.1
* successfully set certificate verify locations:
*   CAfile: ../example.com.crt
  CApath: /etc/ssl/certs
* TLSv1.3 (OUT), TLS handshake, Client hello (1):
* TLSv1.3 (IN), TLS handshake, Server hello (2):
* TLSv1.3 (IN), TLS Unknown, Certificate Status (22):
* TLSv1.3 (IN), TLS handshake, Unknown (8):
* TLSv1.3 (IN), TLS handshake, Request CERT (13):
* TLSv1.3 (IN), TLS handshake, Certificate (11):
* TLSv1.3 (IN), TLS handshake, CERT verify (15):
* TLSv1.3 (IN), TLS handshake, Finished (20):
* TLSv1.3 (OUT), TLS change cipher, Client hello (1):
* TLSv1.3 (OUT), TLS Unknown, Certificate Status (22):
* TLSv1.3 (OUT), TLS handshake, Certificate (11):
* TLSv1.3 (OUT), TLS Unknown, Certificate Status (22):
* TLSv1.3 (OUT), TLS handshake, CERT verify (15):
* TLSv1.3 (OUT), TLS Unknown, Certificate Status (22):
* TLSv1.3 (OUT), TLS handshake, Finished (20):
* SSL connection using TLSv1.3 / TLS_AES_256_GCM_SHA384
* ALPN, server accepted to use h2
* Server certificate:
*  subject: CN=httpbin.example.com; O=httpbin organization
*  start date: Feb 22 16:36:41 2022 GMT
*  expire date: Feb 22 16:36:41 2023 GMT
*  common name: httpbin.example.com (matched)
*  issuer: O=example Inc.; CN=example.com
*  SSL certificate verify ok.
* Using HTTP2, server supports multi-use
* Connection state changed (HTTP/2 confirmed)
* Copying HTTP/2 data in stream buffer to connection buffer after upgrade: len=0
* TLSv1.3 (OUT), TLS Unknown, Unknown (23):
* TLSv1.3 (OUT), TLS Unknown, Unknown (23):
* TLSv1.3 (OUT), TLS Unknown, Unknown (23):
* Using Stream ID: 1 (easy handle 0x557d3aefb620)
* TLSv1.3 (OUT), TLS Unknown, Unknown (23):
> GET /status/418 HTTP/2
> Host:httpbin.example.com
> User-Agent: curl/7.58.0
> Accept: */*
>
* TLSv1.3 (IN), TLS Unknown, Certificate Status (22):
* TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
* TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
* TLSv1.3 (IN), TLS Unknown, Unknown (23):
* Connection state changed (MAX_CONCURRENT_STREAMS updated)!
* TLSv1.3 (OUT), TLS Unknown, Unknown (23):
* TLSv1.3 (IN), TLS Unknown, Unknown (23):
< HTTP/2 418
< server: istio-envoy
< date: Mon, 02 May 2022 21:50:28 GMT
< x-more-info: http://tools.ietf.org/html/rfc2324
< access-control-allow-origin: *
< access-control-allow-credentials: true
< content-length: 135
< x-envoy-upstream-service-time: 3
<

    -=[ teapot ]=-

       _...._
     .'  _ _ `.
    | ."` ^ `". _,
    \_;`"---"`|//
      |       ;/
      \_     _/
        `"""`
* Connection #0 to host httpbin.example.com left intact
```
## Verify the connectivity using http

```shell
$ cd examples/certs/httpbin/client
$ curl-http.sh
```
Output:
```shell
* Added httpbin.example.com:31756:172.16.16.200 to DNS cache
* Hostname httpbin.example.com was found in DNS cache
*   Trying 172.16.16.200...
* TCP_NODELAY set
* Connected to httpbin.example.com (172.16.16.200) port 31756 (#0)
> GET /status/418 HTTP/1.1
> Host:httpbin.example.com
> User-Agent: curl/7.58.0
> Accept: */*
>
< HTTP/1.1 418 Unknown
< server: istio-envoy
< date: Mon, 02 May 2022 21:53:53 GMT
< x-more-info: http://tools.ietf.org/html/rfc2324
< access-control-allow-origin: *
< access-control-allow-credentials: true
< content-length: 135
< x-envoy-upstream-service-time: 2
<

    -=[ teapot ]=-

       _...._
     .'  _ _ `.
    | ."` ^ `". _,
    \_;`"---"`|//
      |       ;/
      \_     _/
        `"""`
* Connection #0 to host httpbin.example.com left intact
```
## Verify the connectivity from the client pod 
```shell
kubectl exec -it client-7d7bf44b5c-qmzhh -c client -n httpbin-ns -- /bin/sh
```
```shell
# curl -v --noproxy '*' -HHost:httpbin.example.com --resolve "httpbin.example.com:31756:172.16.16.2
00" "http://httpbin.example.com:31756/status/418"
* Added httpbin.example.com:31756:172.16.16.200 to DNS cache
* Hostname httpbin.example.com was found in DNS cache
*   Trying 172.16.16.200:31756...
* Connected to httpbin.example.com (172.16.16.200) port 31756 (#0)
> GET /status/418 HTTP/1.1
> Host:httpbin.example.com
> User-Agent: curl/7.79.1
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 418 Unknown
< server: istio-envoy
< date: Mon, 02 May 2022 23:01:06 GMT
< x-more-info: http://tools.ietf.org/html/rfc2324
< access-control-allow-origin: *
< access-control-allow-credentials: true
< content-length: 135
< x-envoy-upstream-service-time: 4
<

    -=[ teapot ]=-

       _...._
     .'  _ _ `.
    | ."` ^ `". _,
    \_;`"---"`|//
      |       ;/
      \_     _/
        `"""`
* Connection #0 to host httpbin.example.com left intact
```

```shell
# curl -v --noproxy '*' -HHost:httpbin "http://httpbin:8000/status/418"
*   Trying 240.0.0.2:8000...
* Connected to httpbin (240.0.0.2) port 8000 (#0)
> GET /status/418 HTTP/1.1
> Host:httpbin
> User-Agent: curl/7.79.1
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 418 Unknown
< server: istio-envoy
< date: Wed, 04 May 2022 23:46:51 GMT
< x-more-info: http://tools.ietf.org/html/rfc2324
< access-control-allow-origin: *
< access-control-allow-credentials: true
< content-length: 135
< x-envoy-upstream-service-time: 3
< x-envoy-decorator-operation: httpbin.httpbin-ns.svc.cluster.local:8000/*
<

    -=[ teapot ]=-

       _...._
     .'  _ _ `.
    | ."` ^ `". _,
    \_;`"---"`|//
      |       ;/
      \_     _/
        `"""`
* Connection #0 to host httpbin left intact
```
## Uninstall the application
Uninstall the app using the command:
```shell
$ ./delete.sh 
```
