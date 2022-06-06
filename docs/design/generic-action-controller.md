```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2019-2022 Intel Corporation
```

# GENERIC ACTION CONTROLLER (GAC)

The EMCO orchestrator supports placement and action controllers to control the deployment of applications. Placement controllers allow the orchestrator to choose the exact locations to place the application in the composite application. Action controllers can modify the state of a resource (create additional resources to be deployed, modify or delete the existing resources). 

GAC is an action controller registered with the central orchestrator. 

## Overview Of GAC

  ### Functionalities provided by GAC

  - ***Create Kubernetes Object(s)***

    GAC allows you to create a new Kubernetes object and deploy it with a specific application. The app should be part of the composite application. You can create and deploy this object in two ways.

      - ***Default***           : Apply the new object to every cluster where the app has deployed.
      - ***Cluster-Specific***  : Apply the new object only to cluster(s) where the app has deployed, specified by a `name`, single cluster, or a  `label`, multiple clusters.

  - ***Modify Kubernetes Object(s)***
  
    GAC allows you to modify an existing Kubernetes object which may have deployed using the helm chart for an app or GAC itself. Modification may correspond to specific fields in the YAML definition of the Kubernetes object.

  ### Components of GAC

   - ***GenericK8sIntent*** - Specifies the parent intent supports the resource and their customizations.
   - ***Resource***         - Specifies the Kubernetes object that need to be created/updated.
   - ***Customizations***   - Specifies the configurations for the Kubernetes object. Customization allows you to customize an existing Kubernetes object or a newly created one. Customization supports uploading multiple data files with configurations data for a ConfigMap/Secret. Customization also helps modify existing Kubernetes objects using JSON patch.

  ### API Definitions

   - ***GenericK8sIntent***

        `Path`
        ```shell
          /projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-k8s-intents
        ```
        `Body`
        ```shell
          {
            "metadata": {
              "name": "operator-gac-intent",
              "description": "descriptionf of operator-gac-intent",
              "userData1": "user data 1 for operator-gac-intent",
              "userData2": "user data 2 for operator-gac-intent"
            }
          }
        ```
   - ***Resource***

        `Path`
        ```shell
          /projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-k8s-intents/{genericK8sIntent}/resources
        ```
        - GAC parses the request body as a `multipart/form-data`.
        - GAC expects the payload in the form-data with `KEY metadata`.
        - GAC expects the resource template file in the form-data with `KEY file`.
        - GAC reads the Kubernetes object configurations from the attached resource template file.

        `Body - form-data`
        - `KEY - metadata`
            ```shell
              {
                "metadata": {
                  "name": "operator-gac-resource",
                  "description": "descriptionf of operator-gac-resource",
                  "userData1": "user data 1 for operator-gac-resource",
                  "userData2": "user data 2 for operator-gac-resource"
                },
                "spec": {
                  "app": "operator",
                  "newObject": "true",
                  "resourceGVK": {
                    "apiVersion": "v1",
                    "kind": "ConfigMap",
                    "name": "cm-game"
                  }
                }
              }
            ```
        - `KEY - file`
            [configmap-game.yaml](../../examples/test-gac/configmap-game.yaml)
          
          - `app`         : Name of the application of interest, in this example operator.
          - `newObject`   : Indicates whether this resource defines a new Kubernetes object. If true, the resource template file must be present in the request except for ConfigMap/Secret. The file should have the resource definition in YAML format. 
          - `resourceGVK` : A reference to the Kubernetes object, in this example ConfigMap.
            - `apiVersion` : A string that identifies the version of the schema the object should have, in this example v1.
            - `kind` : A string that identifies the schema this object should have, in this example ConfigMap.
            - `name` : A string that uniquely identifies the object, in this example cm-game.


   - ***Customizations***

        `Path`
        ```shell
          /projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-k8s-intents/{genericK8sIntent}/resources/{genericResource}/customizations
        ```
        - GAC parses the request body as a `multipart/form-data`.
        - GAC expects the payload in the form-data with `KEY metadata`.
        - GAC expects the customization files in the form-data with `KEY files`.
        - GAC reads the Kubernetes object configurations from the attached customization files.

        `Body - form-data`
        - `KEY - metadata`
            ```shell
              {
                "metadata": {
                  "name": "operator-gac-customization",
                  "description": "description for operator-gac-customization",
                  "userData1": "user data 1 for operator-gac-customization",
                  "userData2": "user data 2 for operator-gac-customization"
                },
                "spec": {
                  "clusterSpecific": "true",
                  "clusterInfo": {
                    "scope": "label",
                    "clusterProvider": "provider_1",
                    "cluster": "cluster_1",
                    "clusterLabel": "label_a",
                    "mode": "allow"
                  },
                  "patchType": "json",
                  "patchJson": [
                    {
                      "op": "replace",
                      "path": "/spec/replicas",
                      "value": 3
                    }
                  ],
                  "configMapOptions": {
                    "dataKeyOptions": [
                      {
                        "fileName": "data-game.yaml",
                        "keyName": "game.properties"
                      },
                      {
                        "fileName": "data-userinterface.yaml",
                        "keyName": "user-interface.properties"
                      }
                    ]
                  }
                }
              }
            ```  
        - `KEY - files`
            [data-game.yaml](../../examples/test-gac/data-game.yaml)
            [data-userinterface.yaml](../../examples/test-gac/data-userinterface.yaml)

          - `clusterSpecific`   : Indicates whether the customizations are specific to clusters where the app has deployed. If true, the `clusterInfo` must be present.
          - `clusterInfo`       : The clusters to which this customization applies.
            - `scope`           : Defines how to identify the clusters to apply the customization. Set the value to `label` to apply the customizations to multiple clusters with a specific label. Set the value to `name` to apply the customizations to a specific cluster.
            - `clusterProvider` : Name of the provider hosting the cluster.
            - `cluster`         : Name of the cluster. Required only if the scope is set to `name`.
            - `clusterLabel`    : A label on the cluster. Required only if the scope is set to `label`.
            - `mode`            : Determines whether the customization is allowed on a cluster or not. Set the value to `allow` to apply the customizations only to the specified set of clusters. Set the value to `deny` to apply the customizations to all clusters except those specified.
          - `patchType`         : Specifies the patch type to modify a Kubernetes object. Set the value to `json` to modify a Kubernetes object using JSON patch. Set the value to `merge` to update an object using Kubernetes strategic merge patch.
          - `patchJson`         : Provides the format for describing changes to a Kubernetes object. Please refer to [JSON Patch](https://github.com/evanphx/json-patch)
            - `op`              : Indicates the operation to perform. Its value MUST be one of `"add", "remove", "replace", "move", "copy", or test"`; other values are errors.
            - `path`            : References a location within the YAML definition of the Kubernetes object where the operation is performed. e.g. you can find the replica count at `/spec/replicas`
            - `value`           : Specifies the new value for the referenced location. e.g. the new value is 3 for the replica count
          - `secretOptions`     : Provides the Secret specific customizations.
          - `configMapOptions`  : Provides the ConfigMap specific customizations.
            - `dataKeyOptions`  : Maps the customization values with the configuration data key for the ConfigMap/Secret. Please refer to [Create ConfigMap/Secret](#using-customization-files)
              - `fileName`      : Name of the customization file. e.g. gameproperties.yaml
              - `keyName`       : Data key name for the ConfigMap/Secret configurations in the customization file. e.g game.properties for all the values in gameproperties.yaml
              - `mergePatch`    : Indicates whether the customization files contain strategic merge patch data.

  > **NOTE**: Please see the `Generic Controller Intent` section in the [OpenAPI](../../docs/swagger-specs-for-APIs/emco_apis.yaml) for the detailed definition of APIs exposed by GAC.

## Creating Kubernetes Objects

GAC allows you to create a new Kubernetes object and deploy it with a specific application. You can create a template file with all the required configurations for the resource and upload it. GAC reads the configurations and applies them to the resource. A resource template file is mandatory except for ConfigMap/Secret. 

For example, if you want to create a cluster-specific `NetworkPolicy` 
  - Create a `GenericK8sIntent`
  - Add a `Resource` to the intent with the `resourceGVK` details and the resource template file
  - Create the `Customization` for the `NetworkPolicy`

> **NOTE**: The materials are available in [test-gac](../../examples/test-gac/). Please follow the [Readme.md](../../examples/test-gac/Readme.md) to execute the tests.
This example illustrates the creation of the `NetworkPolicy` with the configuration data using the resource template file.

  * ##### Add the GAC intent

    Create the `GenericK8sIntent`
    ```shell
      version: emco/v2
      resourceContext:
        anchor: projects/proj1/composite-apps/gac-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/generic-k8s-intents
      metadata:
        name: operator-gac-intent       
    ```
  * ##### Add resource to GAC intent

    The next step is to create the `Resource`. Specify the resource details in the `resourceGVK` section. Upload the template file, which has the configurations for the `NetworkPolicy`.
    ```shell
      version: emco/v2
      resourceContext:
        anchor: projects/proj1/composite-apps/gac-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/generic-k8s-intents/operator-gac-intent/resources
      metadata:
        name: operator-gac-resource
      spec:
        app: "operator"
        newObject: "true"
        resourceGVK:
          apiVersion: "networking.k8s.io/v1"
          kind: "networkpolicy"
          name: "netpol-web"
      file:
        networkpolicy-web.yaml
    ```
  * ##### Add customization for the resource

    The next step is to create cluster-specific `Customization`. You want to deploy the `NetworkPolicy` on clusters with label `label_a`.    
    ```shell
      version: emco/v2
      resourceContext:
        anchor: projects/proj1/composite-apps/gac-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/generic-k8s-intents/operator-gac-intent/resources/operator-gac-resource/customizations
      metadata:
        name: operator-gac-customization
      spec:
        clusterSpecific: "true"
        clusterInfo:
          scope: "label"
          clusterProvider: "provider_1"
          cluster: "cluster_1"
          clusterLabel: "label_a"
          mode: "allow"
      files:
        - customization-dummy.yaml
    ```
    **NOTE**: The emcoctl requires a file to identify the request as a multipart request. You don't have to upload any customization files if you are using postman or any other direct API call (like curl). GAC will ignore the content of the customization files if it's not a ConfigMap/Secret. Customizations using files are currently supported only for ConfigMap/Secret.

## Creating ConfigMap/Secrets

You can create a ConfigMap/Secret in multiple ways. You can create a resource template file and upload it. You can use customization files instead of resource template. You can also use JSON patch to create/modify a ConfigMap/Secret.

  ### Using Resource Template File

  You can create a template file with all the required configurations for the ConfigMap/Secret and upload it. GAC reads these configurations and applies them to the ConfigMap/Secret. Please refer to [Creating Kubernetes Objects](#creating-kubernetes-objects)

  ### Using Customization Files

  GAC supports the creation of ConfigMap/Secrets using customization files. A resource template file is not mandatory for a ConfigMap/Secret. GAC will create the base struct for the ConfigMap/Secret using the values provided in the resourceGVK (`APIVersion, Kind, Name`) and apply the configuration data provided in the JSON patch or Customization files. For example, if you want to create a cluster-specific ConfigMap, and the configuration data for the ConfigMap comes from an external JSON file like info.json, you can accomplish that using GAC. 
   - Create a `GenericK8sIntent`
   - Add a `Resource` to the intent with the `resourceGVK` details
   - Create the `Customization` for the ConfigMap with all the customization files

  GAC, by default, uses the filename as the data key when creating the ConfigMap/Secret using the customization files. For example, if you have uploaded a customization file with the name info.json, the data key for the values in the file will be `info.json`.
  ```shell
    data:
      info.json: |+
        address-pools:
           - addresses:
            - 192.168.20.220-192.168.20.250
            name: default
            protocol: layer2
  ```
  There are cases where we need a ConfigMap with a custom key instead of the filename. You can pass the ConfigMap specific data key option in the customizations. GAC maps the filename with the data key name and replaces the filename with the exact data key name. For example, if you want to upload a customization file with the data key as `network`, you can add the ConfigMap options to the customization request to map the filename info.json to the network.
  ```shell
    configMapOptions:
      dataKeyOptions:
        - fileName: info.json
          keyName: network
  ```
  Once GAC has created this ConfigMap, you can see that the data key for the values in the uploaded file would be `network` instead of the filename.
  ```shell
    data:
      network: |+
        address-pools:
        - addresses:
          - 192.168.20.220-192.168.20.250
          name: default
          protocol: layer2
  ```
  ### Using JSON patch

  You can create a ConfigMap/Secret using the JSON patch. You can pass the JSON patch data in the customization. GAC will create the base struct for the ConfigMap/Secret using the values provided in the resourceGVK(`APIVersion, Kind, Name`) and apply the configuration data provided in the JSON patch. For example, if you want to create a cluster-specific ConfigMap, and the configuration data for the ConfigMap comes from JSON patch, you can accomplish that using GAC. 
   - Create a `GenericK8sIntent`
   - Add a `Resource` to the intent with the `resourceGVK` details
   - Create the `Customization` for the ConfigMap with the JSON patch

  In the JSON patch, you can also get the value from an external service. GAC allows you to pass an HTTP lookup URL in the JSON patch. If the value matches a specif pattern, (has prefix `$(http` and suffix `)$`) GAC will call the external service to get the actual patch value. For example, if you want to get the patch value from a CLM Key-Value pair, you can pass the CLM query endpoint as the value in the patchJson. GAC, at run time, invoke the CLM query endpoint and replace the value in the JSON patch with the service response.
  ```shell
    patchType: "json"
      patchJson: [
        {
          "op": "add",
          "path": "/data/istioingressgatewayaddress",
          "value": "$(http://1.2.3.4:56789/v2/cluster-providers/provider1/clusters/cluster1/kv-pairs/istioingressgatewaykvpairs?key=istioingressgatewayaddress)$" 
        },
      ]
  ```
  In this case, GAC invokes the CLM query endpoint at run time to get the value of the key `istioingressgatewayaddress` and replace the JSON patch value with its response.
  ```shell
    {
      "op": "add",
      "path": "/data/istioingressgatewayaddress",
      "value": "10.10.10.1" 
    },
  ```
  GAC can also replace the `cluster-providers` and `clusters` in the URL at runtime. You can have a placeholder for `cluster-providers` and `clusters` in the URL and pass the `clusterProvider` and `cluster` in the customization. 
  ```shell
    clusterInfo:
      scope: "label"
      clusterProvider: "provider_1"
      cluster: "cluster_1"
      clusterLabel: "label_a"
      mode: "allow"
    patchType: "json"
    patchJson: [
      {
        "op": "add",
        "path": "/data/istioingressgatewayport",
        "value": "$(http://1.2.3.4:56789/v2/cluster-providers/{clusterProvider}/clusters/{cluster}/kv-pairs/istioingressgatewaykvpairs?key=istioingressgatewayport)$"  
      }
    ]
  ```
  In this case, GAC first formats the URL by replacing the placeholders, `{clusterProvider}` and `{cluster}` with values `clusterInfo.clusterProvider` and `clusterInfo.cluster`, and invokes the CLM query endpoint at run time to get the value of the key `istioingressgatewayaddress`, and replace the JSON patch value with its response.
  ```shell
    {
      "op": "add",
      "path": "/data/istioingressgatewayport",
      "value": "1234" 
    },
  ```
  The URL in the JSON patch value will be used as the patch value if it does not follow the format (has prefix `$(http` and suffix `)$`). 
 ```shell
    {
      "op": "add",
      "path": "/data/dns",
      "value": "https://example.dns.com"
    }
```
  In this case, GAC will not invoke the URL.

  > **NOTE**: The materials are available in [test-gac](../../examples/test-gac/). Please follow the [Readme.md](../../examples/test-gac/Readme.md) to execute the tests.
  This example illustrates the creation of the `ConfigMap` with the configuration data using the resource template file, JSON patch, and customization files.
  
   * ##### Add the GAC intent
      ```shell
        version: emco/v2
        resourceContext:
          anchor: projects/proj1/composite-apps/gac-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/generic-k8s-intents
        metadata:
          name: operator-gac-intent
      ```
   * ##### Add resource to GAC intent
      ```shell
        version: emco/v2
        resourceContext:
          anchor: projects/proj1/composite-apps/gac-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/generic-k8s-intents/operator-gac-intent/resources
        metadata:
          name: operator-gac-resource
        spec:
          app: "operator"
          newObject: "true"
          resourceGVK:
            apiVersion: "v1"
            kind: "ConfigMap"
            name: "cm-team"
        file:
          configmap-team.yaml
      ```
   * ##### Add customization for the resource
      ```shell
        version: emco/v2
        resourceContext:
          anchor: projects/proj1/composite-apps/gac-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/generic-k8s-intents/operator-gac-intent/resources/operator-gac-resource/customizations
        metadata:
          name: operator-gac-customization
        spec:
          clusterSpecific: "true"
          clusterInfo:
            scope: "label"
            clusterProvider: "provider_1"
            cluster: "cluster_1"
            clusterLabel: "label_a"
            mode: "allow"
          patchType: "json"
          patchJson: [
            {
              "op": "replace",
              "path": "/data/team_size",
              "value": "10" # original value `5`
            }
          ]
          configMapOptions:
            dataKeyOptions:
              - fileName: data-game.yaml
                keyName: game.properties
              - fileName: data-userinterface.yaml
                keyName: user-interface.properties
        files:
          - data-game.yaml # data key will be `game.properties`
          - data-userinterface.yaml # data key will be `user-interface.properties`
      ```
## Modifying Existing Kubernetes Objects

GAC supports modifying an existing Kubernetes object. For example, you have an `etcd-cluster` in your composite application configured for replica count 3, and you want to update the replica count to 6.

  * ##### Add the GAC intent
    ```shell
      version: emco/v2
      resourceContext:
        anchor: projects/proj1/composite-apps/gac-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/generic-k8s-intents
      metadata:
        name: operator-gac-intent
    ```
  * ##### Add resources to GAC intent
    ```shell
      version: emco/v2
      resourceContext:
        anchor: projects/proj1/composite-apps/gac-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/generic-k8s-intents/operator-gac-intent/resources
      metadata:
        name: operator-gac-resource
      spec:
        app: "operator"
        newObject: "false"
        resourceGVK:
          apiVersion: "apps/v1"
          kind: "StatefulSet"
          name: "etcd"
      file:
        statefulset-etcd.yaml
    ```
  * ##### Add customization to increase the replica count
    ```shell
      version: emco/v2
      resourceContext:
        anchor: projects/proj1/composite-apps/gac-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/generic-k8s-intents/operator-gac-intent/resources/operator-gac-resource/customizations
      metadata:
        name: operator-gac-customization
      spec:
        clusterSpecific: "true"
        clusterInfo:
          scope: "label"
          clusterProvider: "provider_1"
          cluster: "cluster_1"
          clusterLabel: "label_a"
          mode: "allow"
        patchType: "json"
        patchJson: [
          {
            "op": "replace",
            "path": "/spec/replicas",
            "value": 6 # original value `3`
          }
        ]
      files:
        - customization-dummy.yaml
    ```
    **NOTE**: The emcoctl requires a file to identify the request as a multipart request. You don't have to upload any customization files if you are using postman or any other direct API call (like curl). GAC will ignore the content of the customization files if it's not a ConfigMap/Secret. Customizations using files are currently supported only for ConfigMap/Secret.

GAC also supports Kubernetes `strategic merge patch`. Set your patch type as `merge` in your resource customization. You can pass the patch content in the customization file. GAC will utilize the Kubernetes strategic merge patch to update the resource document with the patch data. For example, if you want to modify your Deployment container list using a strategic merge patch, 

  * ##### Add the GAC intent
    ```shell
      version: emco/v2
      resourceContext:
        anchor: projects/proj1/composite-apps/gac-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/generic-k8s-intents
      metadata:
        name: operator-gac-intent
    ```
  * ##### Add resources to GAC intent
    ```shell
      version: emco/v2
      resourceContext:
        anchor: projects/proj1/composite-apps/gac-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/generic-k8s-intents/operator-gac-intent/resources
      metadata:
        name: "deploy-web"
      spec:
        app: "operator"
        newObject: "true"
        resourceGVK:
          apiVersion: "apps/v1"
          kind: "Deployment"
          name: "deploy-web"
      file:
        deployment-web.yaml
    ```
  * ##### Create a customization to add a new container
    ```shell
      version: emco/v2
      resourceContext:
        anchor: projects/proj1/composite-apps/gac-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/generic-k8s-intents/operator-gac-intent/resources/deploy-web/customizations
      metadata:
        name: deploy-web-customization
      spec:
        clusterSpecific: "true"
        clusterInfo:
          scope: "label"
          clusterProvider: "provider_1"
          cluster: "cluster_1"
          clusterLabel: "label_a"
          mode: "allow"
        patchType: "merge"
      files:
        - container-patch.yaml
    ```
    **NOTE**: The materials are available in [test-gac](../../examples/test-gac/). Please refer to the official [Kubernetes documentation](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/) for more details.

## Kubernetes Objects Lifecycle And Status

Creating/Updating/Deleting a GenericK8sIntent/Resource/Customization will modify only the database records. It will not deploy/delete/update the state of the Kubernetes object immediately. You have to `approve` and `instantiate` the intent, for the changes to take effect. Once you instantiate, the EMCO orchestrator will invoke the GAC via gRPC call. GAC actions then update the `app context` and save the details to the `etcd`. The EMCO rsync then organizes the Kubernetes objects in the clusters, based on the data in `etcd`. Please refer to [EMCO Resource Lifecycle and Status](../design/Resource_Lifecycle_and_Status.md) for more details.

  * ##### Approve
    ```shell
      version: emco/v2
      resourceContext:
        anchor: projects/proj1/composite-apps/gac-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/generic-k8s-intents/approve
    ```
  * ##### Instantiate
    ```shell
      version: emco/v2
      resourceContext:
        anchor: projects/proj1/composite-apps/gac-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/generic-k8s-intents/instantiate
    ```
You can further create/update Kubernetes objects after instantiating a deployment intent. For example, assume that you have a statefulset resource in one of the clusters, with the name, `etcd`. You want to update the replica count of this statefulset. You can use GAC to update the replica count. We assume you have created this statefulset using GAC, and a resource and customization are already available in GAC. 

> **NOTE**: The materials are available in [test-gac](../../examples/test-gac/). Please follow the [Readme.md](../../examples/test-gac/Readme.md) to execute the tests.
This example illustrates the update/rollback of a deployment intent.

  * ##### Update customization to increase the replica count
    ```shell
      version: emco/v2
      resourceContext:
        anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeAppGac}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/generic-k8s-intents/{{.GacIntent}}/resources/sts-etcd/customizations
      metadata:
        name: sts-etcd-customization
      spec:
        clusterSpecific: "true"
        clusterInfo:
          scope: "label"
          clusterProvider: "provider_1"
          cluster: "cluster_1"
          clusterLabel: "label_a"
          mode: "allow"
        patchType: "json"
        patchJson: [
          {
            "op": "replace",
            "path": "/spec/replicas",
            "value": 2 # original value `1`
          }
        ]
      files:
        - customization-dummy.yaml
    ```
    **NOTE**: The emcoctl requires a file to identify the request as a multipart request. You don't have to upload any customization files if you are using postman or any other direct API call (like curl). GAC will ignore the content of the customization files if it's not a ConfigMap/Secret. Customizations using files are currently supported only for ConfigMap/Secret.

  * ##### Update GAC compositeApp deploymentIntentGroup
    ```shell
      version: emco/v2
      resourceContext:
        anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeAppGac}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/update
    ```
  Once the deployment is updated successfully, we can see that the replica count of etcd statefulset is now two.

You can also rollback the statefulset state to a previous version. For example, if you want to roll back the above changes,

  * ##### Rollback GAC compositeApp deploymentIntentGroup
    ```shell
      version: emco/v2
      resourceContext:
        anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeAppGac}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/rollback
      metadata:
        description: "rollback to revision 1"
      spec:
        revision: "1"
    ```
  Once the deployment is updated successfully, we can see that the replica count of etcd statefulset is now back to one. Please see [test-update](../../examples/test-update/) for more details.
   
