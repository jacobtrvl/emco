```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2019-2020 Intel Corporation
```

# EMCO JSON tag changes

This document summarizes the changes in EMCO data body changes due to
the camel case convention and database integrity changes.

| Microservice | API                                      | Old JSON tag                     | New JSON tag |
|  :---        | :---                                     | :---                             | :---         |
|  clm         | cluster label                            | label-name                       | clusterLabel |
| orchestrator | app profile                              | spec.app-name                    | spec.app |
|              | deployment intent group                  | spec.profile                     | spec.compositeProfile |
|              |                                          | spec.logical-cloud               | spec.logicalCloud |
|              |                                          | spec.override-values             | spec.overrideValues |
|              |                                          | spec.override-values[].app-name  | spec.overrideValues[].app |
|              | app generic placement intent             | spec.app-name                    | spec.app |
|              |   - all usage inside 'allOf' and 'anyOf' | spec.intent.*.provider-name      | spec.intent.*.clusterProvider |
|              |   - all usage inside 'allOf' and 'anyOf' | spec.intent.*.cluster-name       | spec.intent.*.cluster |
|              |   - all usage inside 'allOf' and 'anyOf' | spec.intent.*.cluster-label-name | spec.intent.*.clusterLabel |
|  dcm         | cluster reference                        | spec.cluster-provider            | spec.clusterProvider |
|              |                                          | spec.cluster-name                | spec.cluster |
|              |                                          | spec.loadbalancer-ip             | spec.loadbalancerIp |
|  dcm         | logical cloud                            | spec.user.user-name              | spec.user.userName |
|  dtc         | inbound intent                           | spec.appName                     | spec.app |
|  ovnaction   | workload intent                          | spec.application-name            | spec.app |
|              |                                          | spec.workload-resource           | spec.workloadResource |
|  gac         | customization intent                     | spec.clusterspecific             | spec.clusterSpecific |
|              |                                          | spec.clusterinfo                 | spec.clusterInfo |
|              |                                          | spec.clusterinfo.clusterprovider | spec.clusterInfo.clusterProvider |
|              |                                          | spec.clusterinfo.clustername     | spec.clusterInfo.cluster |
|              |                                          | spec.clusterinfo.clusterlabel    | spec.clusterInfo.clusterLabel |
|              | generic resource intent                  | spec.appname                     | spec.app |
|              |                                          | spec.newobject                   | spec.newObject |
|              |                                          | spec.resourcegvk                 | spec.resourceGVK |
|              |                                          | spec.resourcegvk.apiversion      | spec.resourceGVK.apiVersion |
|  sfcclient   | sfc client intent                        | spec.chainName                   | spec.sfcIntent |
|              |                                          | spec.chainCompositeApp           | spec.compositeApp |
|              |                                          | spec.chainCompositeAppVersion    | spec.compositeAppVersion |
|              |                                          | spec.chainDeploymentIntentGroup  | spec.deploymentIntentGroup |
|              |                                          | spec.chainNetConrolIntent        | \<removed\> |
|              |                                          | spec.appName                     | spec.app |
|  hpa-plc     | hpa intent                               | spec.app-name                    | spec.app |
|              | hpa consumer intent                      | spec.api-version                 | spec.apiVersion |
|              |                                          | spec.container-name              | spec.container |

# Other DB changes

1. The EMCO microservices now connect to a Mongo databased named *emco* instead of *mco* .
1. A single Mongo collection named *resources* is used for all documents instead of several collections.
