```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2022 Intel Corporation
```

# Overview of Distributed traffic controller (DTC)

DTC enables traffic routing between services that are running across different clusters. It supports both east-west and north-south traffic flow using client/server model. DTC is an EMCO action controller. It also provides a sub controller framework using which new action sub controllers can be added to extend the functionality. Sub-Controllers are registered with DTC using dtc-controllers endpoint. All the sub-controllers registered with DTC are called in sequence based on priority. The DTC controller provides intent APIs, and sub-controllers act based on the those to create various CR/resources.

# DTC Sub-Controllers

Following sub-controllers are available in DTC:

1. Istio (its) - 
    Its sub-controller creates the required Istio resources for traffic flow, both for in and out of the cluster as well as across the cluster. Resources such as gateways, service entry, virtual functions, destination rules and secrets are made part of the application context.
2. SDEWAN (swn) - 
   Swn sub-controller adds SDEWAN functionality, for details refer to: https://gitlab.com/project-emco/core/emco-base/-/blob/main/docs/design/dtc_sdewan_support.md
3. Network Policy (nps) - 
    Nps sub-controller creates network policy resources to enforce pod level security.
4. Service Discovery (sds) - 
   Sds sub-controller provides service discovery functionality, for details refer to: https://gitlab.com/project-emco/core/emco-base/-/blob/main/docs/design/service-discovery-design.md

## Overview of DTC Intents

The DTC intents are based on client/server model, and are grouped using traffic group intents. Further, there can be inbound and outbound intents to support both internal and external traffic flow. Using access point intents, user can enforce clients to allow or deny based on url and methods.


## DTC traffic group Intent

Traffic group Intents is the parent intent that 'holds' the corresponding inbound and outbound intents.

URL:

```
/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents
```
Body:

```
{
  "metadata": {
    "name": "trafficgroup1",
    "description": "Traffic to product catalog service",
    "userData1": "data 1",
    "userData2": "data 2"
  }
}
```
## DTC inbound server Intent

DTC inbound server Intent defines the properties of the server for the inbound traffic.

URL:

```
/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents
```
Body:

```
{
  "metadata": {
    "name": "productcatalogserver",
    "description": "product catalog server",
    "userData1": "data 1",
    "userData2": "data 2"
  },
  "spec": {
    "app": "productcatalogservice",
    "appLabel": "app=productcatalogservice",
    "serviceName": "productcatalogservice"
    "externalName": "productcatalogservice.k8s.com"
    "port" : "3550", 
    "protocol": "TCP"
    "externalSupport": "true"
    "serviceMesh": "istio",
    "management" : {
      "sidecarProxy": "yes",
      "tlsType": "MUTUAL",
      }
    },
    "external": {
      "externalCerts": {
        "serviceCertificate" : ""
        "servicePrivateKey" : ""
        "caCertificate" : ""
    },
    "edgeCNF": "sdewan"
  }
}
```
## DTC inbound clients Intent

DTC inbound clients Intent defines the properties of the clients associated with this server.

URL:

```
/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients
```
Body:

```
{
  "metadata": {
    "name": "productcatalogserverclient1",
    "description": "product catalog server clients",
    "userData1": "data 1",
    "userData2": "data 2"
  },
  "spec": {
    "app": "frontend",
    "appLabel": "app=frontend",
    "serviceName": "frontend"
    "namespaces": ["foo", "bar"] 
    "cidrs": []
  },
}
```
## DTC inbound clients access point Intent

DTC inbound clients access point Intent defines authorization policy for the client.

URL:

```
/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients/{inboundClientsIntent}/access-points
```
Body:

```
{
  "metadata": {
    "name": "productcatalogserverclient1access",
    "description": "product catalog server client access",
    "userData1": "data 1",
    "userData2": "data 2"
  },
  "spec": {
    "action": "ALLOW",
    "url": ["/status/418"],
    "access": ["GET"]
  },
}
```
## DTC outbound client Intent

DTC outbound server Intent defines the properties of the client for the outbound traffic.

URL:

```
/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/outbound-intents
```
Body:

```
{
  "metadata": {
    "name": "outboundclient",
    "description": "mysql client",
    "userData1": "data 1",
    "userData2": "data 2"
  },
  "spec": {
    "app": "mysql",
    "serviceName": mysqlservice,
    "appLabel": "app=mysql"
  },
}
```
## DTC outbound server Intent

DTC outbound server Intent defines the properties of the server associated with this client.

URL:

```
/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/outbound-intents/{outboundClientIntent}/server
```
Body:

```
{
  "metadata": {
    "name": "outboundclient",
    "description": "mysql client",
    "userData1": "data 1",
    "userData2": "data 2"
  },
  "spec": {
    "externalServiceName": "mysql.example.com",
    "port": "3306", 
    "protocol": "TCP",
    "dnsResolution": "STATIC",
    "endPointAddress": "172.25.55.65"
    "external": {
      "externalCerts": {
        "clientCertificate" : ""
        "clientPrivateKey" : ""
        "caCertificate" : ""
    },
  },
}
```
