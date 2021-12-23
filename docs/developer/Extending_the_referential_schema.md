```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2019-2021 Intel Corporation
```


# Notes on working with the EMCO referential schema

## Overview 

### Initial design

**Note:** For versions 21.12 or above, please follow [Extending the referential schema](#Developing-a-new-controller-which-extends-the-referential-schema).

	1. When deployed on Kubernetes, the referential schema is mounted as a configmap named: 'emco-db-ref-schema'
	2. Each EMCO microservice mounts this configmap in './ref-schemas/emco-db-ref-schema.yaml' relative to its the working directory.  See the '<EMCO_repo>/deployments/helm/common/templates/_deployment.tpl' file to see where this configmap mounting is done.
	3. The full schema file for the configmap is built by the EMCO helm chart Makefile:  '<EMCO_repo>/deployments/helm/emcoBase/Makefile' by concatenating all '<EMCO_repo>/src/<microservice>/ref-schema/v1.yaml' files.  See the 'ref-schema' target in that Makefile to see how this is done.
	4. When the EMCO microservices start up, one of the first tasks they do is initialize the Mongo DB by invoking 'NewMongoStore()' in '<EMCO_repo>/src/orchestrator/pkg/infra/db/mongo.go'.  This in turn invokes 'ReadRefSchema()' which will read in the schema file and validate it is 'good'.  The schema file is expected to be found at the location './ref-schemas/emco-db-ref-schema.yaml'.  See '<EMCO_repo>/src/orchestrator/pkg/infra/db/schema.go'.

How to get a copy of the full schema configmap

The schema configmap that is prepared can be pulled out of the helm charts as follows:

	1. cd '<EMCO_repo>/deployments/helm/emcoBase'
	2. run:  make  <-- to build the EMCO helm charts
	3. helm template -s templates/emco-db-ref-schema.yaml schema ./deployments/helm/emcoBase/dist/packages/emco-db-0.1.0.tgz

The output of step 3 is the content of the 'emco-db-ref-schema.yaml'.  

Alternatively, if the top level EMCO Makefile has been invoked with the 'deploy' target, then the directory '<EMCO_repo>/bin/helm' will have the ECMO helm charts.  The following sequence can be done:
	1. In '<EMCO_repo>' directory, run 'make deploy' (with appropriate environment settings) for example:
		a. EMCODOCKERREPO=mydockerrepo.example.com/emco/ BUILD_CAUSE=DEV_TEST USER=myusername make deploy
	2. This (on successful completion) will produce helm charts and an installation script in '<EMCO_repo>'/bin/helm
	3. To get the schema configmap, cd to '<EMCO_repo>'/bin/helm and run:
		a. helm template -s templates/emco-db-ref-schema.yaml schema emco-db-myusername-latest.tgz

## Using the schema outside of a Kubernetes/Helm installation

A developer may wish to run some EMCO microservices directly.  In these cases, ensure that the schema file is present in './ref-schemas/emco-db-ref-schema.yaml' relative to the working directory of the microservice.

Docker compose ?

## Developing a new controller which extends the referential schema

A developer may write a new controller to run along with the EMCO microservices.  If the new controller adds APIs to manage new database resources, the referential schema will need to be updated.  For example, a new action or placement controller will likely define new resources which exist as 'child' resources of the 'deploymentIntentGroup' resource.  These resources may also have referential relationships to other resources like the 'app' resource.  

The following illustrates the sequence of steps that a developer might go through to implement a new sample controller.
NOTE: The steps described here will proceed by adding a new controller within the same code base and directory structure of the EMCO base repository.  In principle, this effort could be done in a completely separate project.

### Step 1 - Build and Install EMCO without the new sample controller

First pull the EMCO repository and build the EMCO microservices.  Install them into a cluster.

### Step 2 - Develop a new sample controller

A new controller will be added to the EMCO.  At a high level, the following development tasks are performed:
	- A Dockerfile is created in '<EMCO_repo>/build/docker' - using other EMCO controller Dockerfiles as examples to follow.
	- A Helm chart is created in '<EMCO_repo>/deployments/helm/emcoBase/<sample controller>' - using other EMCO controller helm charts as examples to follow.
	- The source code of the new controller is created in '<EMCO_repo>'/src/<sample controller>' - using other EMCO controller source as examples to follow.
	- The JSON schema files for validating JSON objects input via the new API are created in '<EMCO_repo>'/src/<sample controller>/json-schemas/<sample json schema files>
	- The referential schema file that extends the referential schema for the new sample controller is created in '<EMCO_repo>/src/<sample controller>/ref-schemas/v1.yaml
		○ For example, the new sample referential schema might look like the following:
```
			  - name: sampleIntent
			    parent: deploymentIntentGroup
			    references:
			      - name: app
```
		○ One new database resource is defined. It is a sample intent which is a child of the 'deploymentIntentGroup'.  It references one other resource, the 'app' resource.

A branch with a commit that illustrates the above changes for a new 'sample' controller can be found: emco-base\examples\referential-integrity\sample

### Step 3 - Build the new sample controller

Since this example integrates the sample controller with the EMCO base controller, running the top level EMCO Makefile will work.
	- Just add the new controller name, e.g. 'sample' to the MODS variable at the top of the Makefile
	- Note, to just build 'sample' - set 'MODS=sample' just before the 'ifndef MODS' clause.
	- Build the sample controller, including the Docker image and Helm chart, for example:
		○ EMCODOCKERREPO=<myrepo.example.com/emco> BUILD_CAUSE=DEV_TEST USER=<myusername> make deploy

### Step 4 - Update the referential schema

In this case, since the sample Helm chart has been created in parallel to the other EMCO controllers, so the complete new referential schema configmap has been created by the Helm make that is invoked when the top level 'make deploy' is executed.  So, the new referential schema configmap can be obtain using one of the 'helm template' commands described above in this document.

Alternatively, if the new controller was created and built separately, the referential schema configmap can be obtained from the EMCO Helm chart and then the develop can manually edit the schema, adding the new resources.

Manually apply the new configmap (it may also be necessary to set the 'namespace' attribute of the updated configmap file to match the local EMCO installation), for example:
	kubectl -n <emco namespace> apply -f new-ref-schema.yaml
	
Another alternative is to directly edit the configmap in the cluster and add the new resources.

NOTE: this process is clearly developer oriented and a method for new controllers to register their referential schema updates without manual intervention should be implemented.

### Step 5 - Install the new sample controller
NOTE: The example process described here is intended to illustrate the addition of a new controller into a running EMCO installation.  If the new controller has been developed as described above, then the new controller has been essentially added to the EMCO services and a new installation of EMCO will include the new controller.

Given that EMCO is already running, just use Helm to install the new controller.  This can be done as follows:
	- cd to: '<EMCO_repo>/deployments/helm/emcoBase'
	- Make the Helm charts:  make
		○ this will make the new sample controller (as well as the others - which presumably have not changed)
	- Install the sample controller chart specifically:
		○ helm -n <emco namespace> install --set enableDbAuth=true --timeout=30m --set db.rootPassword=<db root password> --set db.emcoPassword=<db emco password> --set contextdb.rootPassword=<contextdb root password> --set contextdb.emcoPassword=<contextdb emco password> emco-sample dist/packages/sample-0.1.0.tgz
		○ Note: in the above Helm install command, the same database authentication information that was used to install EMCO must be used for the installation of the new sample controller


### Step 6 - Test the new controller
Consider that the new controller defines a new 'sampleIntent' with a resource that looks like the following:
```
	$ cat sample1.json
	{
	  "metadata": {
	    "name": "sample-intent"
	  },
	  "spec": {
	    "app": "abc",
	    "sampleIntentData": "some data"
	  }
	}
```
	
Attempt to create the resource, for example:
	$ curl -d @sample1.json  http://10.10.10.6:30424/v2/projects/testvfw/composite-apps/compositevfw/v1/deployment-intent-groups/vfw_deployment_intent_group/sampleIntents
	Cannot perform requested operation. Parent resource not found
	
Looking at the logs for the sample controller, the following message can be noted:
What is seen here is that the referential integrity determined that the parent resource, the 'deploymentIntentGroup' does not exist, so the new sample intent is not allowed to be created.
```
	$ kubectl -n emco logs -l app=sample | grep SOURCE  |  jq .
	{
	  "Error": "Creating DB Entry: db Insert parent resource not found: Parent resource not found for sampleIntent.  Parent: map[string]string map[compositeApp:compositevfw compositeAppVersion:v1 deploymentIntentGroup:vfw_deployment_intent_groupx project:testvfw] KeyID: {compositeApp,compositeAppVersion,deploymentIntentGroup,project,sampleIntent,}, Key: model.SampleIntentKey {testvfw compositevfw v1 vfw_deployment_intent_groupx sample-intent}",
	  "Module": {
	    "metadata": {
	      "name": "sample-intent",
	      "description": "description of sample-intent",
	      "userData1": "sample-intent user data 1",
	      "userData2": "sample-intent user data 2"
	    },
	    "spec": {
	      "app": "abc",
	      "sampleIntentData": "some data"
	    }
	  },
	  "Parameters": {
	    "compositeApp": "compositevfw",
	    "compositeAppVersion": "v1",
	    "deploymentIntentGroup": "vfw_deployment_intent_groupx",
	    "project": "testvfw"
	  },
	  "SOURCE": "file[apierror.go:40] func[apierror.HandleErrors]",
	  "level": "error",
	  "msg": "Error :: ",
	  "time": "2021-10-05T16:35:14.006766844Z"
	}
```

Now, assuming that the parent resource has been created (in this example, the '<EMCO_repo>/examples/scripts/vfw-test.sh' is being used to create the parent resources), the new sample resource can be created:
```
	$ curl -d @sample1.json  http://10.10.10.6:30424/v2/projects/testvfw/composite-apps/compositevfw/v1/deployment-intent-groups/vfw_deployment_intent_group/sampleIntents
	{"metadata":{"name":"sample-intent"},"spec":{"app":"abc","sampleIntentData":"some data"}}
	
	$ curl  http://10.10.10.6:30424/v2/projects/testvfw/composite-apps/compositevfw/v1/deployment-intent-groups/vfw_deployment_intent_group/sampleIntents
	[{"metadata":{"name":"sample-intent"},"spec":{"app":"abc","sampleIntentData":"some data"}}]
```

Now, check the sample controller logs.  One thing to note about the current referential integrity implementation is that the reference to the 'app' resource is not enforced (does not prevent creation of the resource).  The reference is captured in the new sample intent resource, and a warning log message is created.  For example, in this scenario, there is no 'app=abc' resource:
```
	{
	  "SOURCE": "file[mongo.go:452] func[db.(*MongoStore).verifyReferences]",
	  "level": "warning",
	  "msg": "Resource reference not found",
	  "referenceKey": {
	    "app": "abc",
	    "compositeApp": "compositevfw",
	    "compositeAppVersion": "v1",
	    "project": "testvfw"
	  },
	  "resource": "sampleIntent",
	  "time": "2021-10-05T16:45:02.407003184Z"
	}
```

The actual mongo document of the new sample resource looks like this:
```
	{
	    "_id" : ObjectId("615c810edd1338585e3e6e8c"),
	    "compositeApp" : "compositevfw",
	    "compositeAppVersion" : "v1",
	    "deploymentIntentGroup" : "vfw_deployment_intent_group",
	    "project" : "testvfw",
	    "sampleIntent" : "sample-intent",
	    "data" : {
	        "metadata" : {
	            "name" : "sample-intent"
	        },
	        "spec" : {
	            "sampleapp" : "abc",
	            "sampleintentdata" : "some data"
	        }
	    },
	    "keyId" : "{compositeApp,compositeAppVersion,deploymentIntentGroup,project,sampleIntent,}",
	    "references" : [ 
	        {
	            "key" : {
	                "compositeApp" : "compositevfw",
	                "project" : "testvfw",
	                "compositeAppVersion" : "v1",
	                "app" : "abc"
	            },
	            "keyid" : "{app,compositeApp,compositeAppVersion,project,}"
	        }
	    ]
	}
```

If the 'app=abc' was an error and needs to be corrected, use the PUT API of the sample controller to fix the resource.  Set 'app=sink' (to match with one of the apps in the vfw-test example).

```
	$ curl -X PUT -d @sample1.json  http://10.10.10.6:30424/v2/projects/testvfw/composite-apps/compositevfw/v1/deployment-intent-groups/vfw_deployment_intent_group/sampleIntents/sample-intent
	{"metadata":{"name":"sample-intent"},"spec":{"app":"sink","sampleIntentData":"some data"}}
```

Looking at the updated mongo document, it can be seen to be correct now:
```
	{
	    "_id" : ObjectId("615c810edd1338585e3e6e8c"),
	    "compositeApp" : "compositevfw",
	    "compositeAppVersion" : "v1",
	    "deploymentIntentGroup" : "vfw_deployment_intent_group",
	    "project" : "testvfw",
	    "sampleIntent" : "sample-intent",
	    "data" : {
	        "metadata" : {
	            "name" : "sample-intent"
	        },
	        "spec" : {
	            "sampleapp" : "sink",
	            "sampleintentdata" : "some data"
	        }
	    },
	    "keyId" : "{compositeApp,compositeAppVersion,deploymentIntentGroup,project,sampleIntent,}",
	    "references" : [ 
	        {
	            "key" : {
	                "project" : "testvfw",
	                "compositeAppVersion" : "v1",
	                "compositeApp" : "compositevfw",
	                "app" : "sink"
	            },
	            "keyid" : "{app,compositeApp,compositeAppVersion,project,}"
	        }
	    ]
	}
```

## Next Steps

To make the development and deployment of new controllers simpler and less error prone (i.e. manually update the referential schema configmap), a new schema management feature is being planned.
The outline is that a new controller, on database initialization at startup, will submit its referential schema segment to the schema manager and if the new segment results in a good schema, the complete schema will be returned and the controller continues.  The schema manager will, in turn, save the new schema segment in the database to persist the changes.

[Phase 1](#Phase-1)

	- Schema manager is built into the db package. Every base EMCO service runs it.
	- Schema manager is bootstrapped with the existing configmap based approach.  This is done to get a starting schema that encompasses all of the base EMCO controllers.  New controllers are generally expected to extend to the data model provided by the base EMCO.
	- Instead of configmap, the referential schema is built into the docker images of the base EMCO services. Externally developed controllers would not be expected to have access to the whole schema file.

Phase 2 ideas

	- Develop a new schema manager controller.  API calls are made via gRPC.
	- Just the schema manager controller is bootstrapped with the full base EMCO schema file.
	- Only the schema manager reads/writes to the schema resources in the databases.
	- Other EMCO microservices get the schema via gRPC from the schema manager controller.
	- New controllers submit their new schema segment via gRPC API - either get back the full schema or an error (if they provided a bad schema segment)

## Phase 1

Phase 1 is in the scope of release 21.12.
### Scope

	- The scope of phase 1 is to have every controller define its schema and register it in the schema. Every time the controller starts, use its schema defined in the schema file and the schema segments defined by other controllers to create a consolidated referential schema map. 
	- Enforce the referential integrity of the data using the referential schema map. As part of phase 1, we have enforced the referential integrity for any database insert/ update. The database insert/update will fail if the resources you are trying to create\update are not present in the referential schema map.
	- Schema update is out of the scope of phase 1.  You cannot update the schema of an existing controller. At this point, you will have to clear the database and re-install all the controllers to update the schema. Updating the schema of the controller without deleting the database/ other controllers will be addressed in a future release.

We expect each controller will have a schema file located in the **ref-schemas** folder with the file name **v1.yaml**(`ref-schemas/v1.yaml`). Every time a controller starts, we read the schema definitions from this file during the database connection initialization. We have a JSON schema validation to ensure that the data in the file matches the defined schema. The referential schema JSON validation file is present in the orchestrator(`src/orchestrator/json-schemas/db.json`) for reference.

**Note** In the case of emco, we have consolidated the schema of all the controllers into a single base schema. This base schema is available in the orchestrator(`src/orchestrator/ref-schemas/v1.yaml`). Orchestrator register this base schema in the database and, other controllers read it from the database and create the referential schema map at run time.

When the controller starts, if the controller does not have any schema defined, we will continue with the db initialization and create a referential schema map with the schema segments available in the db. When the controller has a schema file defined, we read the schema definition from the file. If the schema defined in the file is not valid or we could not read the file, in this case, we will log appropriate error messages and proceed without a referential schema map created for the controller.

**Note** In this scenario, the controller will start without issues. However, any insert/ update to the resources defined by the controller will fail.

Each schema segment registered, by the controllers, in the database will have a unique Id. We generate this segment Id using the content in the schema file. For example, if the schema file has two resources, as shown below, the segment Id will be a hash of the content in the file. Any change to the content in the file will create a new segment Id(for example, adding a space, newline, new resource, new reference) 

```
	name: "C1"
		resources:
		  - name: network
		    parent: cluster
		  - name: providerNetwork
		    parent: cluster

```

At any point in time, only one controller can register a resource. We do not allow multiple controllers to define the same resources in their schema. If two controllers define the same resource name in their schema, one of the resources will fail to create the referential schema map. For example, controller C2 has two resources(network, providerNetwork) dependent on a `cluster`. However, the `cluster` resource is defined and registered in the controller C1 schema. In this case, if both the controllers define `cluster` as a resource, we consider that as a duplicate. One of these controllers will fail to create the referential schema map. It depends on the order in which these two controllers start. The latter will fail with a duplicate resource error and continue without a referential schema map.

**Note** In this scenario, the controller will start without issues. However, any insert/ update to the resources defined by the controller will fail.

```
	name: "C1"
	resources:
		- name: cluster
	  	- name: clusterLabel
	      parent: cluster

```

A valid schema definition for controller C2 is as follows.

```
	name: "C2"
	resources:
	    - name: network
	  	  parent: cluster
		- name: providerNetwork
		  parent: cluster

```		

The controller C2 will fail to create the referential schema map with the following schema since the resource `cluster` is also part of C1.	 

```	
	name: "C2"
	resources:
		- name: cluster 
	    - name: providerNetwork
	      parent: cluster
	    - name: network
	      parent: cluster

```

As mentioned, when the controller starts, we read the schema definition from the file. Then we query the database and retrieve the schema segments registered by other controllers. We use each of these schema segments to create a consolidated referential schema map. Each resource associated with the referential schema map will have the following details.

	- parent - Name of the parent resource
	- keyId - A string identifier for this resource
	- children - List of children
	- KeyMap- A mapping structure for looking up this item
	- references- List of references
	- referencedBy - List of resources that may reference this resource

A resource can be dependent on another as a parent/ children/ references/ referencedBy. The dependent can be a part of the same schema or defined by another controller schema. In scenarios where the dependent belongs to another controller schema, we expect the schema is registered and available in the database. The controller will fail to create the referential schema map if the dependent resource is not available in the database. For example, controller C1 has three resources. Two of the resources are dependent on other resources(parent-child relationship).  In this scenario, controller C1 has all the necessary resources available in the same schema. So, controller C1 will start without any schema dependency failures.

```
	name: "C1"
	resources:
		- name: clusterProvider
		- name: cluster
		  parent: clusterProvider
		- name: clusterLabel
		  parent: cluster

```

Consider a scenario where we have a controller C2, which has a dependency on the resources defined by the controller C1 schema.

```
	name: "C2"
	resources:
		- name: providerNetwork
		  parent: cluster
		- name: network
		  parent: cluster

```
In this scenario, controller C2 will not have the referential schema map created until the schema for controller C1 becomes registered and available in the database(since it has a dependency on the resource cluster).

The scenario mentioned above can happen in many cases. One of the use cases is when we install multiple controllers as a single package(for example,  EMCO). There would be a possibility that there might be a delay in registering the schemas by individual controllers. In this case, any controllers who have a dependent resource defined by another controller will not have a referential schema map created. We have introduced a retry logic to handle this. As mentioned earlier, every controller, when it starts, reads its schema from the file. Then it queries and retrieves the schema segment registered in the database by other controllers. Next, we try to create a referential schema map using the resources defined in these schema segments. If we cannot find the dependent resources in any of these schema segments, then we will wait for a specified time(5 seconds). We query the database again and retrieve the schema segments. The controller will fail to create the referential schema map if we cannot find the dependent resource even after multiple retries.

For example, there are two controllers C1 and C2. The controller C2 has a resource, `cluster`, defined in the schema of controller C1. In this case, if controller C1 could not update its schema in the database, controller C2 will fail to create the referential schema map even after multiple retries.

```
	name: "C1"
	resources:
	  - name: cluster
	  - name: clusterLabel
	    parent: cluster
	  
	name: "C2"
	resources:
	  - name: network
	    parent: cluster
	  - name: providerNetwork
	    parent: cluster

```

**Note** The retry time interval is configurable. You can change the default values in the config.json file presented in the controller(s) root folder. For example, if you want to change the default retries and interval for the orchestrator(`emco-base\src\orchestrator`), please modify the `config.json` file accordingly.

```
	{
		"database-type": "mongo",
		"database-ip": "mongo",
		"etcd-ip": "etcd",
		"db-schema-max-backoff": 60,
		"db-schema-backoff": 5
	}

``` 
