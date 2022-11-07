```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2022 Intel Corporation
```
# AppConfig API aka Day 2 Configuration

After the application is instantiation by EMCO, some resources, require configuring. This document goes over the API's provided by EMCO to support that.

## Terminology

- </b>Resource Type</b> - Application specific Resources for an application.(For example: CRDs for Kubernetes based applications). There can be many resource types per application. Typically, these resources are not part of the original Helm chart.

- </b>App Config</b> - App config is one instance of the resource type. There can be multiple app configs per resource type.


## Requirements

Some of the requirements for App Config support are:

- Multiple Resource types per Application.

- There can be multiple App Configs per deployment intent group.

- There can be multiple App Config per application.

- Each App Config correspondes to one Resource Type.

- Multiple App Configs can be provided together

- Support to change or delete individual App Configs


## Resource types:

There can be many resource types per application. The resource types that will be supported in the beginning will be CRD based configurations. Other resource type that will be supported in future include json based, tosca etc.

For CR based resource type, Resource type API will be used to provide the GVK for the CR along with a template with the default values.
For other resource type fields corresponding to those will be provided, for example for json based it can be url etc.

```
#creating resource-type  definition
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeApp}}/{{.CompositeAppVersion}}/resource-type
metadata :
   name: myCRD1
spec:
  appName: {{.App}}
  type: crType # Other types to be supported icluding json, tosca etc.
  crType:
    apiVersion: myGroup.k8s.io/v1
    kind: MyCRDKind
file:
  {{.Template}} # Template with default values
```

## App Configurations:

App Configuration supports 2 types of file formats:

#### File

In this format the complete file is provided for configuration (and no patching is required).

#### Patch File

In this format a patch file is provided and patching is required. A patch file - either a patchesJson6902 or patchesStrategicMerge (Refer to kustomize documentation for these methods: https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/).


There are 2 types of AppConfigs:

- Application level
- Deployment Intent level


#### Application Level

These configurations applies for the whole application and are not related to any deployment intent group. At the time of
resource creation or update this configuration is applied right away.


##### App Configuration API

```
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeApp}}/{{.CompositeAppVersion}}/app-config
metadata:
  name: app-config-1
spec:
  - resourceType: myCRD1
    appConfigName: config1
    fileType: patch # 2 types supported are file or patch
    patchType: patchesStrategicMerge # 2 types supported are patchesJson6902 and patchesStrategicMerge. Only valid if patch type
    clusterSpecific: "true"
    clusterInfo:
      scope: label
      clusterProvider: {{.ClusterProvider}}
      clusterName: ""
      clusterLabel: "label_a"
    files:
      - {{.ConfigurationFile}} # This is CR file in this example
      - {{.ConfigurationFile}} # This is CR file in this example
  - resourceType: myCRD1
    appConfigName: config2
    fileType: patch # 2 types supported are file or patch
    patchType: patchesStrategicMerge # 2 types supported are patchesJson6902 and patchesStrategicMerge. Only valid if patch type
    clusterSpecific: "true"
    clusterInfo:
      scope: label
      clusterProvider: {{.ClusterProvider}}
      clusterName: ""
      clusterLabel: "label_a"
    files:
      - {{.ConfigurationFile}} # This is patch file in this example

```

##### Support for various Rest API Methods

Post - Post will be with the above url where multiple AppConfigs are provided together.

Get, Put and delete will use the Url

```
projects/{{.ProjectName}}/composite-apps/{{.CompositeApp}}/{{.CompositeAppVersion}}/app-config/{{.appConfigName}

```

Get will respond with Multipart output with the above information.

Put and Delete will be also based on one App Config

Open - Is it required to delete the whole set togther?


#### Deployment Intent Level

These configurations applies for to one instance of the application as denoted by deployment intent group. At the time of
resource creation if the deployment intent group is already instantiated then the configuration can be applied right away. If not then the configuration has to be applied after the deployment intent group is instantiated.


##### App Configuration API

API definition is same as at the Application level. The url will be:

```

projects/{{.ProjectName}}/composite-apps/{{.CompositeApp}}/{{.CompositeAppVersion}}/deployment-intent-groups/{{.DeploymentIntent}}/app-config

```

##### Lifecycle API's for configuration

Open: Should we support automatic application of the configuration or use LCM API's? In the first case if the deployment intent group is in instantiated state then the App Configs are applied directly and if not applied on instantiation. In that case lifecycle management API are not required.

Lifecycle operations:

```
APPLY - projects/{{.ProjectName}}/composite-apps/{{.CompositeApp}}/{{.CompositeAppVersion}}/deployment-intent-groups/{{.DeploymentIntent}}/app-config/{{.AppConfig}}/apply
```
```
TERMINATE - projects/{{.ProjectName}}/composite-apps/{{.CompositeApp}}/{{.CompositeAppVersion}}/deployment-intent-groups/{{.DeploymentIntent}}/app-config/{{.AppConfig}}/terminate
```
```
UPDATE - projects/{{.ProjectName}}/composite-apps/{{.CompositeApp}}/{{.CompositeAppVersion}}/deployment-intent-groups/{{.DeploymentIntent}}/app-config/{{.AppConfig}}/update
```

