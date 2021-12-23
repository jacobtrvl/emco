# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

# Overview of EMCO data resources and structure of Database documents

The EMCO database is a NoSQL database (i.e. MongoDB) and is used primarily to store data that is input via the EMCO REST APIs.
With MongoDB, the unit of information stored in the DB is called a `document`.  These documents contain the data supplied by the user as well as other information used by EMCO.  Data elements within a `document` are key/value pairs, where a value may be a simple string or a complex object.

Each document (or resource) in the DB has the following generic structure:

```
<key information>
<data>
<other data content>
<EMCO object state information>
  <reference information>
```

## EMCO Data Resource Key Information
EMCO data resources are identified by a Key, which is just a map of key value pairs.  There is a connection between the REST API URL for a data resource and the corresponding Key.

Example GET API URL for an `app` resource:

```
Generic URL:
GET  <https//EMCO_Microservice:Port>/v2/project/{project}/composite-app/{compositeApp}/{compositeAppVersion}/app/{app}

Specific URL:
GET  <https//EMCO_Microservice:Port>/v2/project/projectOne/composite-app/CA-123/v3/app/nginx
```
The above URL shows the general form for querying an `app` and then a specific example for `nginx`.

The Key for the app `nginx` looks as follows:

```
{
  "project": "projectOne",
  "compositeApp": "CA-123",
  "compositeAppVersion": "v3",
  "app": "nginx"
}
```
Some important points about EMCO DB resources and Keys:

- The EMCO REST API and corresponding DB resources have a hierarchcial (parent/child) arrangement.  In the above example, the `nginx` app is a child of the composite app version resource `CA-123/v3` and the project `projectOne` is the top most parent.
  - NOTE: `compositeApp/compositeAppVersion` is a special case where two key elements are used to identify a single data resource.
- The names of the Key elements (e.g. `project`, `compositeAppVersion`) are required to be unique accross all EMCO resources.  The result is that a given Key element name also can be used to identify a given data resource.

The `<key information>` part of the MongoDB document that holds the `nginx` data resource includes the above Key and an additional attribute called `keyId`.  The `keyId` is a sorted sequence of the Key element names in the Key.  The `keyId` field essentially identifies the type of resource this document is holding - e.g. an `app` resource.

```
Mongo Document format with example Key shown
{
  "project": "projectOne",
  "compositeApp": "CA-123",
  "compositeAppVersion": "v3",
  "app": "nginx",
  "keyId": "{app,compositeApp,compositeAppVersion,project,}",

  <data>
  <other data content>
  <EMCO object state information>
  <reference information>
}
```

## EMCO Data Resource Data
All of the EMCO REST APIs that create EMCO data resources will supply a JSON data object.  In virtually all cases (see exception section below), the JSON data object conforms to the following format:

```
{
  "metadata": {
    "name": <required name value>,
    "description": <optional description string>,
    "userData1": <optional user supplied data string>,
    "userData2": <optional user supplied data string>
  },
  "spec": {
    <JSON data appropriate for the specific EMCO data resource>
  }
}
```
This JSON data object is included in the EMCO data resource document in the `data` attribute, as shown via the following example.
Note that in this example, the JSON data for the `app` resource does not have a `spec` section.


```
Mongo Document format with example Key and data shown
{
  "project": "projectOne",
  "compositeApp": "CA-123",
  "compositeAppVersion": "v3",
  "app": "nginx",
  "keyId": "{app,compositeApp,compositeAppVersion,project,}",

  "data": : {
    "metadata": {
      "name": "nginx",
    },
  }
  <other data content>
  <EMCO object state information>
  <reference information>
}
```

Some important points:
- The value of `metadata.name` is the name of the EMCO data resource and is the value used in the Key.  In the above example, the data resource is an `app` resource.  The `app` element of the Key takes the value of `nginx` - which is the same value as `metadata.name`.

## EMCO Data Resource - Other data content

Some EMCO data resources include an additional element in the Mongo document to hold other data.  Typically, this is the content of a file, such as the Helm chart for an `app` or the kubeconfig file for a `cluster` or similar.  The name of this additional data content attribute and the contents are handled by the code for that specific data resource.

Following shows an example for the example `nginx` app document:

```
Mongo Document format with example Key and data shown
{
  "project": "projectOne",
  "compositeApp": "CA-123",
  "compositeAppVersion": "v3",
  "app": "nginx",
  "keyId": "{app,compositeApp,compositeAppVersion,project,}",
  "data": : {
    "metadata": {
      "name": "nginx",
    },
  },
  "appcontent": {
    "filecontent": ".... <long base64 encodes string of nginx Helm chart file> ..."
  },
  <EMCO object state information>
  <reference information>
}
```

## EMCO Data Resource - EMCO Object state information
Some EMCO data resources also act as the endpoints for EMCO lifecycle operations (e.g. `instantiate`, `terminate`, `update`, etc.).  The Mongo document for these resources (e.g. `deploymentIntentGroup`, `cluster`) will have an attribute named `stateInfo`.  Further details are not relevant for this document.

## EMCO Data Resource - Reference Information

As already described, all EMCO data resources are part of a hierachical parent/child structure.  Every child resource references its parent resource.  This child to parent reference is already built into the Key information of the resource, so no further information is required.

However, some resources may refer to other resources.  In this case, EMCO keeps a list of the `Key` and `keyId` of those references in the `references` attribute of the Mongo document for the resource.

The following example illustrates the `references` list of a `deploymentIntentGroup` resource.  There are two references - to `logicalCloud` and `compositeProfile` resources.


```
Mongo Document format with example references list:
{
  ...
  "references": [
       {
            "key" : {
                "logicalCloud" : "adminLogCloud",
                "project" : "projectOne"
            },
            "keyid" : "{logicalCloud,project,}"
        }, 
        {
            "key" : {
                "compositeProfile" : "Example-composite-profile",
                "compositeApp" : "CA-123",
                "project" : "projectOne",
                "compositeAppVersion" : "v3"
            },
            "keyid" : "{compositeApp,compositeAppVersion,compositeProfile,project,}"
        }
  ]
}
```

# DB Referential Integrity

Referential integrity for EMCO data resources means the following rules must be satisfied:

1. When a data resource is created or updated, any referenced resources must already exist.
     - The parent resource must exist
     - Any resources that are in the `references` list must exist
1. A data resource may not be deleted if it is referenced by another resource
     - The data resource has no child resources
     - The data resource must not be present in any other resource's `references` list

To accomplish this, each EMCO microservice that provides a REST API to CRUD data resources must define a referential schema for those resources.  The aggregation of all EMCO microservice referential schemas combines to define the overall referential schema for all EMCO data resources.  The referential schema is used by the database interface package to enforce the referential integrity rules.

## The Referential Schema

The overall referential schema is just a YAML document with a list of resources, as such:

```
resources:
  - <resource entry 1>
  - <resource entry 2>
  ...
```

Each resource entry has the following format:

```
  - name: <resource name>        # [required] this is the key element name of the resource - e.g. app, project, etc.
    parent: <resource name>      # [required, except for top level resources] name of the parent resource
    references:                  # [optional] needed when resource has references
      - name: <resource name>    #   [required] name of the refereced resource
        type: ["map" | "many"]   #   [optional] only needed if a "map" or "many" reference type is being defined
        map: <JSON tag of map> 
        fixedKv:
          <fixedKey>: <fixedValue>
        filterKeys:
          - <filterKeyName> 
        commonKey: <resource name>
```

The schema for each resource defines the parent child relationships and any additional referenced resources.
For the referenced resource elements, the `name` is the primary information required.  The additional schema attributes are needed to handle specific types and condtions.  These will be described below.


### References - `type`

Each `references` element in the schema has a `name` which identifies the type of resource that is referenced.  The `type` of reference can be:

1. Unspecified (i.e. default) - there is a reference to one instance of the identified resource type.
1. Map - there can be multiple references to the identified resource type.  The instances will be the key names of a map object in the JSON data of this resource
1. Many - there can be multiple references to the identified resource type. The key elements of any given reference must be simple key/value attributes in a given JSON object within a possibly more complex JSON object.  For example, the `genericAppPlacement` resource defines a complex object which has multiple arrays and nested arrays.  Each element of these arrays may contain a `cluster` reference.

### References - `map`

When the reference `type` is `"map"`, then the `map` attribute value identifies the name of a JSON map object in the resource data.  For example, the `spec` part of the JSON data of the `groupIntent` resource looks like the following:

```
        "spec" : {
            "intent" : {
                "genericPlacementIntent" : "generic-placement-intent",
                "ovnaction" : "vfw_ovnaction_intent"
            }
        }
```
The key values of the `intent` map identify the names of controller resources, so the value of the `map` attribute in the schema will be `"intent"`.

### References - `fixedKv`

The `fixedKv` attribute was introduced to address the same `map` example described above, although it may be used in all reference `types` (although there are no examples of this yet).

The controller data resource has a key of:

```
{
    "controllerGroup": <group>,
    "controller": <controller name>
}
```
The `controllerGroup` attribute is generally fixed for any given EMCO microservice.  This allows multiple EMCO microservices to use the same controller API handling code for their own controllers.  E.g. the `orchestrator` microservice will group all of its placement and action controllers in the `"orchestrator"` group and the `dtc` microservice can group all of its subcontrollers under the `"dtc"` group.

The `fixedKv` attribute in the schema is used to identify this fixed element of the referenced resource key.  For example,

```
    references:
      - name: controller
        type: map
        map: intent
        fixedKv:
          controllerGroup: orchestrator
```
When the database package is preparing the keys for referenced resources, it will use the key/values from the `fixedKv` map.

### References - `filterKeys`

The `filterKeys` is another schema attribute that is required to handle a special case for the `map` example shown above.
In the above example, one of the keys in the `"intent"` map is `"genericPlacementIntent"`.  The generic placement controller is a built-in function of the `orchestrator` microservice, so a `controller` resource is not added to the database.  So, a reference to the `"genericPlacmentIntent"` cannot be added to the `references` list of the resource, otherwise, the referential integrity rules would prevent creation of the resource (since the a `genericPlcaementIntent` controller resource does not exist).

This issue is handled by the `filterKeys` attribute - which means - when finding references of the given resource, do not include instances which match items in the `filterKeys` list.  For example,

```
    references:
      - name: controller
        type: map
        map: intent
        fixedKv:
          controllerGroup: orchestrator
        filterKeys:
          - genericPlacementIntent
```

### References - `commonKey`
The `commonKey` schema attribute handles a case for the default (i.e. unspecified) reference `type`.  It identifies which part of the Key of the referenced resource should be taken from the referencing resource.  An example is the `sfcClientIntent` which has a `spec` and referential schema that look as follows:

```
sfcClientIntent spec:

   "spec": {
      "chainEnd": "left",
      "sfcIntent": <sfcIntent name>,            <-- Referenced resource part of a different deploymentIntentGroup
      "compositeApp": "SFC-CA",                 <-- Different than the compositeApp this sfcClientIntent is part of
      "compositeAppVersion": "v2",              <-- Different than the compositeAppVersion this sfcClientIntent is part of
      "deploymentIntentGroup": "SFC-CA-DIG",    <-- Different than the deploymentIntentGroup this sfcClientIntent is part of
      "app": <client app name>,                 <-- Referenced resource that is part of same deploymentIntentGroup
      "workloadResource": <resource name>,
      "resourceType": "pod"
   }


sfcClientIntent referential schema:

  - name: sfcClientIntent
    parent: deploymentIntentGroup
    references:
      - sfcIntent
      - app
        commonKey: compositeAppVersion
```

In the above example, the database reference finding code will find most of the key elements for the `sfcIntent` reference from the `spec` portion of the data.  However, for the `app` reference, the `commonKey` attribute says to fill the `app` key up through the `compositeAppVersion` from the referencing (i.e. `sfcClientIntent` resource) Key.  If the `app` reference schema did not specify the `commonKey` attribute, then the Key for the `app` would get filled out with elements from `spec` and would be incorrect.


## Database handling of the Referential Schema Information

### On Create/Update of Data Resources

When a given resource is passed to the database interface (i.e. Insert()), the code will do the following:

```
Look up the resources referential schema information
Verify that the parent resource intance exists
For each reference in references list
	Prepare a list of keys for this reference (0 or more based on type)
	Append the keys to the references list
Verify that all resources in the references list exist

Update the references list in the DB resource (i.e. Mongo document)
```

The key preparation process:

1. The code gets the Key from the schema
2. The `spec` section of the resource data is scanned and any elements of the key that are found are filled in (subject to modification due to the `type`, `filterKeys`, `map`, `commonKey` attributes)
3. Any remaining Key elemnents are filled in from the resource's own Key
4. Any Key elements identified in `fixedKv` are set
5. Any Key that is not fully filled out is tossed


### On Deletion of Data Resources

A query for the number of resources which match the Key elements of the resource is made.  Since all children resources will also match the Key element of the resource to be deleted, if the number of resources is greater than one, then the resource has children and cannot be deleted.

Then a query if made to find out if the Key and `KeyID` of the resource to be deleted are present in any `references` array of any document.  If the nubmer of resources is non-zero, then the resource cannot be deleted.

## Referential Schema Aggregation and Build Process

Each microservice will provides a piece of the referentials schema in a file located and named:
```
<EMCO repo>/src/<micoservice>/ref-schemas/v1.yaml
```

When EMCO is built from the top-level Makefile file, all of the referential schema files are concatentated together and used to prepare a `ConfigMap` which gets deployed along with the EMCO services.  Each EMCO microservice will mount this `configmap` and read and process the schema at startup.

### Possible Enhancements

For now, it should be possible for a new controller that is built and installed after EMCO has already been installed to be built and operate correctly.

1. The new controller should be built with all of the other EMCO microservices so that the referential schema will be recreated along with the new controller schema.
2. The updated `configmap` and controller are installed.
3. Existing EMCO services won't reference resources that they don't know about.  The schema is only accessed on Insert() to the database.
4. If existing EMCO services restart, they will see the new schema, which is fine.

A more robust mechanism to update the schema at runtime, while ensuring invalid schemas are not introduced, may be useful to create.  But, the initial impelmentation should work for a start.

# Schema Recommendations

1. All resource identifiers - as identified by the JSON tag - must be unique across EMCO.
1. For a given resource, the identifier must match in the referential schema file as well as in the JSON tag of the Key structure(s) for querying the resource.
1. If a resource is referenced by another resource via key element attributes in that resources `spec` object, the JSON tag of those key element attributes just use the corresponding resource identifier value.
1. Generally, it is recommended to design resource JSON data objects such that referenced resources can be identified by the default reference `type`.  Avoiding types `map` and `many` if possible.


# Exceptions and Corner Case Resources

## `compositeApp.compositeAppVersion` resource name

The `compositeApp` and `compositeAppVersion` are two separate values in the EMCO REST API and are two distinct key elements for ECMO resources that include these elememnts in their key.

However, it happens to be the case that EMCO stores a single resource for the `compositeApp` and the `compositeAppVersion` is defined in the `spec` object.

The referential schema handles this by allowing the resource name to be entered in a dotted notation - `compositeApp.compositeAppVersion`.

## `controllerGroup.controller` resource

The `controller` is another resource which is represented in the schema as a composite of two key elements.  As described earlier this is because the `controllerGroup` is statically defined for a given set of controllers - such as `orchestrator` action controllers.

## `clusterLabel` data structure

The `clusterLabel` resource does not follow the convention of using a JSON object of the form
```
{
  "metadata": {
    "name": <name>
  },
  "spec": {
    ...
  }
}
```
It has the form:
```
{
  "clusterLabel": <label>
}
```

## `clusterLabel` as a reference

Several resources (e.g. `genericAppPlacement`) can either refer to a `cluster` or `clusterLabel`.  A `clusterLabel` is a reference to any `cluster` that has a label with the specified `clusterLabel` value.

A reference can be identified for a specific `cluster`, but the `clusterLabel` is ambiguous since there could be zero to many clusters that have a given `clusterLabel`. Also, the `cluster` element of the `clusterLabel` key is essentially a wildcard.

As a result, the `clusterLabel` resource has not been listed as a reference in the referential schema.

## Network in the `ovnaction` `interfaceIntent` resource

The `interfaceIntent` resource has a `NetworkName` attribute in its `spec` object.  In practice, this network name could be referencing either a `network` or `providerNetwork` that has been deployed to a cluster by the `ncm` microservice.  

The `interfaceIntent` resource schema does not define a reference to a network resource, because there is not enough information.  It is again a possible many to one reference - because each network for each cluster is a separate EMCO resource.  Additionally, the `interfaceIntent` resource does not provide any cluster key information.  Not to mention that the referenced network could be one of two possible resource types - `network` or `providerNetwork`.

As a result, no reference for network is defined for the `interfaceIntent` resource.

## `providerNetwork` in the `sfcProviderNetwork`

The `sfcProviderNetwork` resource has a similar issue.

The `sfcProviderNetwork` has a `networkName` attribute in its `spec` object that identifies a `providerNetwork`.
This is not identified as a resource in the schema because it is difficult to determine at the time of resource creation which `providerNetwork` resources are being referenced.

## The `"map"` reference type in the referential schema

The `map` reference `type` only handles identifying references for the key elements of the map.  In the initial implementation, the one `map` example uses the key element to identify `controller` resource references.  The value of the map is the name of an intent managed by the associate controller.  These are, in principle also references of the `intentGroup` resource, but to identify these resources would require an enhancement of the schema.  Such as, each microservice needs to identify the mapping between its controller name and the intent resource it manages.  A further complication is that the `controller` name is a dynamic value that is determined by the user at runtime.

Since the intents supported by the controller are siblings to the `grouptIntent` resource (they all have `deploymentIntentGroup` as a parent), complexity of solving this problem is low priority for this intial implementation.

## `cloudConfig` resource

The `cloudConfig` resource is a resource with the key
```
{
  "clusterProvier": <name>,
  "cluster": <name>,
  "level": <level>,
  "namespace": <namespace>
}
```
This object is managed by the `rsync` microservice, and is used to hold the kubeconfig for the specified cluster.  On creation of a `cluster` the `clm` microservice will also create a `cloudConfig` to hold the kubeconfig provided by the user.  The `dcm` microservice will store additional kubeconfigs (i.e. per level and namespace) as `logicalCloud` resources are created and corresponding kubeconfigs are generated for the clusters within the logical cloud.

Currently, the `cloudConfig` resource is not represented in the referential schema since management of this resource is handled completely internal to EMCO.

Also, the `level` and `namespace` elements of the `cloudConfig` key do not represent resource instances like almost all other key elements do.

# Sample EMCO Referential Schema

```
    name: emco-clm
    resources:
      - name: clusterProvider                          
      - name: cluster                                  
        parent: clusterProvider
      - name: clusterLabel                             
        parent: cluster
      - name: clusterKv                                
        parent: cluster

    name: emco-dcm
    resources:  
      - name: logicalCloud                             
        parent: project
      - name: clusterReference                         
        parent: logicalCloud
        references:
          - name: cluster
      - name: clusterQuota                             
        parent: logicalCloud
      - name: logicalCloudKv                           
        parent: logicalCloud
      - name: userPermission                           
        parent: logicalCloud

    name: emco-dtc
    resources:
      - name: trafficGroupIntent                       
        parent: deploymentIntentGroup
      - name: inboundServerIntent                      
        parent: trafficGroupIntent
        references:
          - name: app
      - name: inboundClientsIntent                     
        parent: inboundServerIntent
        references:
          - name: app

    name: emco-gac
    resources:
      - name: genericK8sIntent                         
        parent: deploymentIntentGroup
      - name: genericResource                          
        parent: genericK8sIntent
        references:
          - name: app
      - name: customization                            
        parent: genericResource
        references:
          - name: cluster

    name: emco-ncm
    resources:
      - name: providerNetwork                          
        parent: cluster
      - name: network                                  
        parent: cluster

    name: emco-orchestrator
    resources:
      - name: controllerGroup.controller               
      - name: project                                  
      - name: compositeApp.compositeAppVersion         
        parent: project
      - name: app                                      
        parent: compositeAppVersion
      - name: compositeProfile                         
        parent: compositeAppVersion
      - name: appProfile                               
        parent: compositeProfile
        references:
          - name: app
      - name: deploymentIntentGroup                    
        parent: compositeAppVersion
        references:
          - name: logicalCloud
          - name: compositeProfile
      - name: groupIntent                              
        parent: deploymentIntentGroup
        references:
          - name: controller
            type: map
            map: intent
            fixedKv:
              controllerGroup: orchestrator
            filterKeys:
              - genericPlacementIntent
      - name: genericPlacementIntent                   
        parent: deploymentIntentGroup
      - name: genericAppPlacementIntent                
        parent: genericPlacementIntent
        references:
          - name: app
          - name: cluster
            type: many

    name: emco-ovnaction
    resources:
      - name: netControllerIntent                      
        parent: deploymentIntentGroup
      - name: workloadIntent                           
        parent: netControllerIntent
        references:
          - name: app
      - name: interfaceIntent                          
        parent: workloadIntent

    name: emco-sfc
    resources:    
      - name: sfcIntent                                
        parent: deploymentIntentGroup
      - name: sfcClientSelector                        
        parent: sfcIntent
      - name: sfcProviderNetwork                       
        parent: sfcIntent

    name: emco-sfcclient
    resources: 
      - name: sfcClientIntent                           
        parent: netControllerIntent
        references:
          - name: sfcIntent
          - name: app
            commonKey: compositeAppVersion
      - name: hpaIntent
        parent: deploymentIntentGroup
        references:
          - name: app
      - name: hpaConsumer
        parent: hpaIntent
      - name: hpaResource
        parent: hpaConsumer
```

See [Extending the referential schema](Extending_the_referential_schema.md).