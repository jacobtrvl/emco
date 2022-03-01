# Getting Started

## EMCO

The ```Edge Multi-Cluster Orchestrator (EMCO)``` is a software framework for the intent-based deployment of cloud-native applications to a set of Kubernetes clusters, spanning enterprise data centers, multiple cloud service providers, and numerous edge locations. It is architected to be flexible, modular, and highly scalable.

Refer to [EMCO Documentation](docs/design/emco-design.md) for details on EMCO architecture.

## EMCO Controllers

The EMCO orchestrator supports `placement` and `action` controllers to control the deployment of applications. `Placement controllers` allow the orchestrator to choose the exact locations to place the application in the composite application. `Action controllers` can modify the state of a resource(create additional resources to be deployed, modify or delete the existing resources). You can define your packages and functionalities based on your need and expose these functionalities using the gRPC server. In EMCO, we have separate controllers for action and placement. 

## Writing a new controller

This document provides a high-level overview of developing a new controller, the components required for a new controller, and building and deploying the new controller using helm charts.

### Defining Packages

### main

Package ```main``` is the entry point to the controller. This package initializes the databases(mongo and etcd) grpc server and starts the webserver.

```

package main

// initDataBases initializes the emco databases
func initDataBases() error {
	// Initialize the emco database(Mongo DB)
	err := db.InitializeDatabaseConnection("emco")
	if err != nil {
		return err
	}
	// Initialize etcd
	err = contextdb.InitializeContextDatabase()
	if err != nil {
		return err
	}
	return nil
}

// initGrpcServer start the gRPC server
func initGrpcServer() {
	go func() {
		if err := grpc.StartGrpcServer(); err != nil {
			log.Fatal(err)
		}
	}()
}

// serve start the controller and handle requests on incoming connections
func serve() error {
	p := config.GetConfiguration().ServicePort
	r := api.NewRouter(nil)
	h := handlers.LoggingHandler(os.Stdout, r)
	server := &http.Server{
		Handler: h,
		Addr:    ":" + p,
	}
	c, err := auth.GetTLSConfig("ca.cert", "server.cert", "server.key")
	if err != nil {
		return server.ListenAndServe()
	}

	return server.ListenAndServeTLS("", "")
}

```

### api

Package ```api``` defines all the routes and their associated handler functions. This example implements two HTTP methods. It registers two routes to create and retrieve the intents associated with a deployment group.

```

package api

// NewRouter creates a router that registers the various routes.
func NewRouter(mockClient interface{}) *mux.Router {
	const baseURL string = "/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/sampleIntents"
	r := mux.NewRouter().PathPrefix("/v2").Subrouter()
	c := module.NewClient()
	h := intentHandler{
		client: setClient(c.SampleIntent, mockClient).(module.SampleIntentManager),
	}
    r.HandleFunc(baseURL, h.handleSampleIntentCreate).Methods("POST")
	r.HandleFunc(baseURL+"/{SampleIntent}", h.handleSampleIntentGet).Methods("GET")
	return r
}

```

### action

Package ```action``` applies the specific action invoked by the application scheduler(orchestrator) using the gRPC call(s). The action is associated with the controller type. Placement controllers decide where the specific application should get placed, and the action controller modifies the current state of the resources.

**In EMCO, we have separate controllers for action and placement.**

ref: https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/hpa-ac - `HPA action controller`

ref: https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/hpa-plc - `HPA placement controller`

In the case of an `action controller`, the controller updates the app context based on the intent name and the app context id. You can create a new resource(e.g. ConfigMap, Secret, Network Policy, etc.) or update an existing resource(e.g. add new data sections to a ConfigMap, annotate or label a resource, etc.) based on your requirement using an action controller.

```

package action

// UpdateAppContext applies the supplied intent against the given AppContext ID
func UpdateAppContext(intentName, appContextId string) error {
    var ac appcontext.AppContext
	_, err := ac.LoadAppContext(appContextId)
	if err != nil {
		return err
	}
	_, err = ac.GetCompositeAppHandle()
	if err != nil {
		return err
	}
	appContext, err := context.ReadAppContext(appContextId)
	if err != nil {
		return err
	}
	project := appContext.CompMetadata.Project
	app := appContext.CompMetadata.CompositeApp
	version := appContext.CompMetadata.Version
	group := appContext.CompMetadata.DeploymentIntentGroup
	// Look up all  Intents
	intents, err := module.NewClient().SampleIntent.GetSampleIntents("", project, app, version, group)
	if err != nil {
		return err
	}
	if len(intents) == 0 {
		return errors.Errorf("No intents are defined for the deploymentIntentGroup: %s", group)
	}
	for _, i := range intents {
		// Implement the action controller specific logic here.
		// Action controllers modifies the current state of the resources.
		// ref: https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/hpa-ac/internal/action - HPA action controller
		logutils.Info(i.Metadata.Name,
			logutils.Fields{})
	}
	return nil
}

```

In the case of a `placement controller`, the controller allow the orchestrator to choose the exact locations to place the application in the composite application. 

```

package action

// FilterClusters applies the supplied intent against the given AppContext ID
func FilterClusters(appContextID string) error {
	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(appContextID)
	if err != nil {
		return err
	}
	ca, err := ac.GetCompositeAppMeta()
	if err != nil {
		return err
	}
	project := ca.Project
	app := ca.CompositeApp
	version := ca.Version
	group := ca.DeploymentIntentGroup
	// Look up all  Intents
	intents, err := module.NewClient().SampleIntent.GetSampleIntents("", project, app, version, group)
	if err != nil {
		return err
	}
	if len(intents) == 0 {
		return errors.Errorf("No intents defined for the deploymentIntentGroup: %s", group)
	}
	for _, i := range intents {
		// Implement the placement controller specific logic here.
		// Placement controllers decide where the specific application should get placed.
		// ref: https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/hpa-plc - HPA placement controller
		logutils.Info(i.Metadata.Name,
			logutils.Fields{})
	}
	return nil
}

```

### grpc

Package ```grpc``` creates a new network listener using the provided port and a gRPC server. Then register the service and its implementation to the gRPC server. In EMCO, each controller communicates with the application scheduler(orchestrator) through the gRPC calls.

```

package grpc

// StartGrpcServer start the gRPC server and register with the application scheduler(orchestrator)
func StartGrpcServer() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	server := grpc.NewServer([]grpc.ServerOption{}...)

	// In this sample controller, we have shown how to register the action and placement controllers.
	// Registering the same controller as action and placement may or may not work.
	// This is for illustration purposes only since the code structure is the same for action or placement controller.
	// In EMCO, we have separate controllers for action and placement.

	// ref: https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/hpa-ac - HPA action controller

	// ref: https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/hpa-plc - HPA placement controller

	// Register the action controller
	contextupdate.RegisterContextupdateServer(server, actioncontroller.NewActionControllerServer())

	// Register the placement controller
	orchplacementcontroller.RegisterPlacementControllerServer(server, placementcontroller.NewPlacementControllerServer())

	if err = server.Serve(listener); err != nil {
		return err
	}

	return nil
}

```

### controller

The other packages in `pkg/grpc` expose the gRPC server-specific functionalities of a controller. These functionalities and associated packages will change based on the requirement(s). As mentioned above, a controller can be either a placement or action controller. Placement controllers allow the orchestrator to choose the exact locations to place the application in the composite application. Action controllers can modify the state of a resource. You can define your packages and functionalities based on your need and expose these functionalities using the gRPC server. These functions are getting invoked by the application scheduler(orchestrator). 

The code structure is the same for both action and placement controllers. In this sample controller, we have defined two packages in grpc.

Package ```actioncontroller``` exposes the gRPC server-specific functionalities of an action controller.

```

package actioncontroller

type actionControllerServer struct {
	contextupdate.UnimplementedContextupdateServer
}

func (ac *actionControllerServer) UpdateAppContext(ctx context.Context, req *contextupdate.ContextUpdateRequest) (*contextupdate.ContextUpdateResponse, error) {
	err := action.UpdateAppContext(req.IntentName, req.AppContext)
	if err != nil {
		return &contextupdate.ContextUpdateResponse{AppContextUpdated: false, AppContextUpdateMessage: err.Error()}, err
	}

	return &contextupdate.ContextUpdateResponse{AppContextUpdated: true, AppContextUpdateMessage: "Context updated successfully."}, nil
}

// NewActionControllerServer exported
func NewActionControllerServer() *actionControllerServer {
	s := &actionControllerServer{}
	return s
}

```

Package ```placementcontroller``` exposes the gRPC server-specific functionalities of a placement controller.

```

package placementcontroller

type placementControllerServer struct {
}

func (ac *placementControllerServer) FilterClusters(ctx context.Context, req *placementcontroller.ResourceRequest) (*placementcontroller.ResourceResponse, error) {
	if (req != nil) && (len(req.AppContext) > 0) {
		err := action.FilterClusters(req.AppContext)
		if err != nil {
			return &placementcontroller.ResourceResponse{AppContext: req.AppContext, Status: false, Message: "Failed to filter clusters."}, err
		}

		return &placementcontroller.ResourceResponse{AppContext: req.AppContext, Status: true, Message: ""}, nil
	}

	return &placementcontroller.ResourceResponse{Status: false, Message: "Invalid request."}, errors.New("invalid request")
}

// NewPlacementControllerServer exported
func NewPlacementControllerServer() *placementControllerServer {
	s := &placementControllerServer{}
	return s
}

``` 

### model

Package ```model``` contains the data model necessary for the implementations. In this example, SampleIntent

```

package model

// SampleIntent defines the high level structure of a network chain document
type SampleIntent struct {
	Metadata Metadata   `json:"metadata" yaml:"metadata"`
	Spec     SampleIntentSpec `json:"spec" yaml:"spec"`
}

// SampleIntentKey is the key structure that is used in the database
type SampleIntentKey struct {
	Project               string `json:"project"`
	CompositeApp          string `json:"compositeApp"`
	CompositeAppVersion   string `json:"compositeAppVersion"`
	DeploymentIntentGroup string `json:"deploymentIntentGroup"`
	SampleIntent          string `json:"sampleIntent"`
}

```

### module

Package ```module``` implements all the business logic. It is a middleware/facade between the handler and the database.

```

package module

// Client combines different clients into a single type. Every handler is associated with a client. The handler then uses its associated client to perform the requested operation. 
// You can have different clients based on the requirement and its implementation. In this example, we only have one client.
type Client struct {
	SampleIntent *SampleIntentClient
	// Add other required clients here.
	// ref: https://gitlab.com/project-emco/core/emco-base/-/blob/main/src/ncm/pkg/module/module.go
}

// NewClient returns a new client instance
func NewClient() *Client {
	c := &Client{}
	c.SampleIntent = NewIntentClient()
	// If you have multiple clients, init them here.
	// ref: https://gitlab.com/project-emco/core/emco-base/-/blob/main/src/ncm/pkg/module/module.go
	return c
}

// DBInfo represents the mongoDB properties
type DBInfo struct {
	collection string // name of the mongodb collection to use for client documents
	tag        string // attribute key name for the json data of a client document
}

// SampleIntentClient implements the SampleIntentManager. It will also be used to maintain some localized state.
type SampleIntentClient struct {
	dbInfo DBInfo
}

NewIntentClient returns a new intent client instance
func NewIntentClient() *SampleIntentClient {
	return &SampleIntentClient{
		dbInfo: DBInfo{
			collection: "resources", // should remain the same
			tag:        "data", // should remain the same
		},
	}
}

// A manager is an interface for exposing the client's functionalities. You can have multiple managers based on the requirement and its implementation. 
// In this example, SampleIntentManager exposes the SampleIntentClient functionalities.
type SampleIntentManager interface {
	CreateSampleIntent(intent model.SampleIntent, project, app, version, deploymentIntentGroup string, failIfExists bool) (model.SampleIntent, error)
	GetSampleIntents(name, project, app, version, deploymentIntentGroup string) ([]model.SampleIntent, error)
}

// CreateSampleIntent insert a new SampleIntent in the database
func (i *SampleIntentClient) CreateSampleIntent(intent model.SampleIntent, project, app, version, deploymentIntentGroup string, failIfExists bool) (model.SampleIntent, error) {
	// Construct key and tag to select the entry.
	key := model.SampleIntentKey{
		Project:               project,
		CompositeApp:          app,
		CompositeAppVersion:   version,
		DeploymentIntentGroup: deploymentIntentGroup,
		SampleIntent:          intent.Metadata.Name,
	}
	// Check if this SampleIntent already exists.
	intents, err := i.GetSampleIntents(intent.Metadata.Name, project, app, version, deploymentIntentGroup)
	if err == nil &&
		len(intents) > 0 &&
		intents[0].Metadata.Name == intent.Metadata.Name &&
		failIfExists {
		return model.SampleIntent{}, errors.New("SampleIntent already exists")
	}
	err = db.DBconn.Insert(i.dbInfo.collection, key, nil, i.dbInfo.tag, intent)
	if err != nil {
		return model.SampleIntent{}, err
	}
	return intent, nil
}

// GetSampleIntents returns the SampleIntent for the corresponding name
func (i *SampleIntentClient) GetSampleIntents(name, project, app, version, deploymentIntentGroup string) ([]model.SampleIntent, error) {
	// Construct key and tag to select the entry.
	key := model.SampleIntentKey{
		Project:               project,
		CompositeApp:          app,
		CompositeAppVersion:   version,
		DeploymentIntentGroup: deploymentIntentGroup,
		SampleIntent:          name,
	}
	values, err := db.DBconn.Find(i.dbInfo.collection, key, i.dbInfo.tag)
	if err != nil {
		return []model.SampleIntent{}, err
	}
	if len(values) == 0 {
		return []model.SampleIntent{}, errors.New("SampleIntent not found")
	}
	intents := []model.SampleIntent{}
	for _, v := range values {
		i := model.SampleIntent{}
		err = db.DBconn.Unmarshal(v, &i)
		if err != nil {
			return []model.SampleIntent{}, errors.Wrap(err, "Unmarshalling Value")
		}

		intents = append(intents, i)
	}
	return intents, nil
}

```

## Ref Schema

If the controller wants to create a new resource in the database, it should be present in the ```referential schema```. The referential schema enforces the referential integrity rules for any database insert/update/delete operation(s). For example, a new action or placement controller will likely define new resources as 'child' resources of the 'deploymentIntentGroup' resource. These resources may also have referential relationships to other resources like the 'app' resource.

See [ReferentialIntegrity](https://gitlab.com/project-emco/core/emco-base/-/tree/main/docs/developer/ReferentialIntegrity.md)

In this example, we want to create a sampleIntent and, this sampleIntent should have a deploymentIntentGroup(parent) and a referenced app.

```

name: sample
resources:
  - name: sampleIntent
    parent: deploymentIntentGroup
    references:
      - name: app

```

## Build

The following steps assume that the new controller is developed within the same code base and directory structure of the EMCO repository (https://gitlab.com/project-emco/core/emco-base/-/tree/main)

1. The source code of the new controller is present in `<emco_repo>/src/controller-name` similar to other controllers. In this example `<emco_repo>/src/sample`.

2. The `JSON schema` files for validating the JSON objects input are present in `<emco_repo>/src/controller-name/json-schemas/`. In this example `<emco_repo>/src/sample/json-schemas/`.

3. The `referential schema` file that extends the referential schema for the new sample controller is present in `<emco_repo>/src/controller-name/ref-schemas/v1.yaml`.  In this example `<emco_repo>/src/sample/ref-schemas/v1.yaml`

4. A `Dockerfile` is present in `<emco_repo>/build/docker/Dockerfile.controller-name`. In this example `<emco_repo>/build/docker/Dockerfile.sample`. The docker file contains information about the arguments, base image, the entry point of the controller, and the files required for the controller. The sample Dockerfile is available in `examples/sample-controller/build/docker/Dockerfile.sample`.

```

ARG EMCODOCKERREPO
ARG SERVICE_BASE_IMAGE_NAME
ARG SERVICE_BASE_IMAGE_VERSION
FROM ${EMCODOCKERREPO}${SERVICE_BASE_IMAGE_NAME}:${SERVICE_BASE_IMAGE_VERSION}

WORKDIR /opt/emco/sample

RUN addgroup -S emco && adduser -S -G emco emco
RUN chown emco:emco . -R

COPY --chown=emco ./sample ./
COPY --chown=emco ./config.json ./
COPY --chown=emco ./json-schemas ./json-schemas
COPY --chown=emco ./ref-schemas ./ref-schemas 

USER emco

ENTRYPOINT ["./sample"]

```

5. `Helm Charts` are present in `<emco_repo>/deployments/helm/emcoBase/controller-name`. In this example `<emco_repo>/deployments/helm/emcoBase/sample`. The sample helm charts are available in `examples/sample-controller/deployments/helm/sample/`.

## Deploying the new controller

Since we are integrating the new controller with EMCO, we can follow the same deployment techniques of EMCO. At this point, we assume you have a working EMCO environment and a basic understanding of EMCO deployments. Please see [EMCO Deployment](README.md) if you don't have a working EMCO environment and want to install EMCO.

### Integrate the new controller into a running EMCO instance

If the new controller has developed as described above, follow these steps to install and integrate it with a running EMCO instance.

1. Pull the EMCO repository with the changes including, the new controller.

2. Modify the `Makefile` at the root level to set the `MODS` variable. Remove all other controller names from the MODS variable and add the new controller name. In this example sample.

    ```

    ifndef MODS
    MODS=sample
    endif

    ```

3. Build the sample controller using the Docker image and Helm charts `make deploy EMCODOCKERREPO=<myrepo.example.com/emco> BUILD_CAUSE=DEV_TEST USER=${USER}-latest`

4. Navigate to `cd <emco_repo>/deployments/helm/emcoBase`

5. Make the helm charts `make`

6. Install the sample controller `helm -n  install --set enableDbAuth=true --timeout=30m --set db.rootPassword=pwd --set db.emcoPassword=pwd --set contextdb.rootPassword=pwd --set contextdb.emcoPassword=pwd emco-sample dist/packages/sample-0.1.0.tgz`

```Please provide the same namespace and database passwords in the above commands you used to install EMCO```

### Deploy the new controller with other EMCO controllers

Please follow these steps to install the new controller with other EMCO controllers in a single deployment.

1. Pull the EMCO repository with the changes including, the new controller.

2. Modify the `Makefile` at the root level to set the `MODS` variable and add the new controller name. In this example sample.

    ```

    ifndef MODS
    MODS=clm dcm dtc nps sds its genericactioncontroller monitor ncm orchestrator ovnaction rsync tools/emcoctl sfc sfcclient hpa-plc hpa-ac sample
    endif

    ```

3. Build the sample controller using the Docker image and Helm charts `make deploy EMCODOCKERREPO=<myrepo.example.com/emco> BUILD_CAUSE=DEV_TEST USER=${USER}-latest`

4. Navigate to `cd <emco_repo>/bin/helm/`

5. Install EMCO `./emco-base-helm-install.sh -s 'enableDbAuth=true --set db.rootPassword=pwd --set db.emcoPassword=pwd --set contextdb.rootPassword=pwd --set contextdb.emcoPassword=pwd --timeout=30m' install`

### Test the new controller

Create a new resource that looks like the following:

```

cat sample1.json
	{
	  "metadata": {
	    "name": "sample-intent"
	  },
	  "spec": {
	    "app": "abc",
	    "sampleIntentData": "some data"
	  }
	}

curl -d @sample1.json  http://10.10.10.3:30424/v2/projects/testvfw/composite-apps/compositevfw/v1/deployment-intent-groups/vfw_deployment_intent_group/sampleIntents

```

The mongo document of the new sample resource looks like this:

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

Retrieve the newly created intent:

```

curl http://10.10.10.3:30424/v2/projects/testvfw/composite-apps/compositevfw/v1/deployment-intent-groups/vfw_deployment_intent_group/sampleIntents/sample-intent

```

### Register the controller

The following steps explain registering the new sample controller with the vfw example in examples/single-cluster/test-vfw.yaml.

1. Register the controller as action or placement type

```

cat sample-controller.json
{
    "metadata": {
        "name": "sample"
    },
    "spec": {
        "host": "10.10.10.3", // where the controller is running
        "port": 30425, // the grpc port
        "type": "action", // "placement" for the placement controller
        "priority": 1 
    }
}

curl -d @sample-controller.json http://10.10.10.3:30415/v2/controllers 

```

2. Create Sample Intent

```

cat sample-intent.json
{
	"metadata": {
    	"name": "vfw_sample-intent"
  	},
  	"spec": {
    	"app": "abc",
    	"sampleIntentData": "some data"
  	}
}

curl -d @sample-intent.json http://10.10.10.3:30424/v2/projects/testvfw/composite-apps/compositevfw/v1/deployment-intent-groups/vfw_deployment_intent_group/sampleIntents

```

3. Add sampleIntent to the deployment intent group

```

cat add-intent.json
{
	"metadata": {
    	"name": "fw-deployment-sample-intent"
  	},
	"spec": {
	    "intent": {
			"genericPlacementIntent": "fw-placement-intent",
			"ovnaction": "vfw_ovnaction_intent",
			"sample": "vfw_sample_intent"
		}
	}
}

curl -d @add-intent.json http://10.10.10.3:30415/v2/projects/testvfw/composite-apps/compositevfw/v1/deployment-intent-groups/vfw_deployment_intent_group/intents

```
