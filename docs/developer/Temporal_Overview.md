```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2022 Intel Corporation
```

# Temporal Workflow Client

This Document is a reference for both TAC, and the demo for TAC listed in examples.
The demo stands up TAC, the workflow client, workflow server, and workers and executes a simple hello world application as a proof of concept.

The reader is expected to be familiar with
[EMCO](https://gitlab.com/project-emco/core/emco-base) and
[Temporal](https://docs.temporal.io/docs/temporal-explained/introduction).
In particular, it is important to read the document [Temporal Workflows in
EMCO](https://gitlab.com/project-emco/core/emco-base/-/blob/main/docs/user/Temporal_Workflows_In_EMCO.md) first.

## Introduction

The Edge Multi-Cluster Orchestrator (EMCO), an open source project in Linux Foundation Networking, has been enhanced to launch and manage Temporal workflows. This repository contains a reference workflow that migrates a stateless application deployed by EMCO in one cluster to another specified cluster. This can be taken as a template to develop workflows to migrate stateful applications and other workflows as well.

## Temporal Background

Temporal is a scalable and reliable runtime for Reentrant Processes called Temporal Workflow Executions.

In general a workflow execution is executed inside a worker entity which is inside a worker process. There can be one or more worker entities inside of a worker process, and you can register one or more worker process with the temporal server. A worker process is a process the user writes that completes a specific job. The process is made up of two items - activities and workflows. Workflows a deterministic functions written by the user that will not fail. Activities are non-deterministic functions that are expected to fail within acceptable paramaters depending on the state of resources they are interacting with.

Worker processes are then bound to a temporal server and a task queue. When a worker is registered to a task queue they must have an identical composition of activities and workflows because task queues are associated with specific jobs to be completed. When a user wants to schedule a job to be completed they will write a workflow client that will submit a job to the temporal server along with all the data needed to complete the job. The temporal will then takes these requests off the task queue in no specific order, and spawn a worker entity to complete the requested job.

## Temporal Workflow Structure in EMCO

There are three main components that make Temporal run: workflow client - which is two different processes working in tandem in our case, the temporal server, and the worker.

The temporal workflow client is the process that is in charge of telling the temporal server to queue a job. There are two different processes that let us accomplish this in EMCO. 

The first the http server. The http server takes temporal requests from the TAC action controller in EMCO. It takes the request, and saves the data locally in a way that temporal can understand. From it runs the second process to actually queue the job. The second process, the workflow client, creates a connection to the temporal server and queues the job on the temporal server to be eventually completed when there are resources available.

The temporal server itself is the orchestrator of the entire operation. It is the entity that schedules workers for execution, and also takes in requests from workflow clients. The Temporal server requires no modification from us, and is spun up with default configuration following temporal guides. The most common, and easiest, way to bring up a temporal server on a local machine is using docker-compose while following the [temporal docs.](https://docs.temporal.io/clusters/quick-install/) Ultimately it does not matter how the temporal server is is brought up as long as both the workflow client, and worker can see it.


The final portion of the of temporal is the worker itself. The worker is a composition of activities and workflow written by the user to a complete a specific overarching job. For each job that needs to be completed the user will write a corresponding worker process and bind it to a specific queue. Everytime the user wants to invoke this job they will tell the temporal server to put another job on a specific task queue with all the required data.


## Temporal Client in EMCO

To add a new client to EMCO the user must follow the template provided in [EMCO source code](https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/tools/workflowclient)

The HTTP server portion of the workflow client will remain largely unedited. It has been abstracted to be able to accommodate generally any workflow client the user writes. The workflow clients job name is taken from its only registered route /invoke/{client-name}. The HTTP server will take all the data associated with the job request, and package for the workflow client and then store it into a locally stored JSON file. The http server will then execute the workflow client and provide the relative location of the JSON file via command line arguments.

The workflow client itself is what must be written by the user, although most of the time a barely edited version workflow client provided in the [emco source code](https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/tools/workflowclient) will suffice. The general flow of the workflow client will be unpacking the data from the JSON file, submitting the job and data to the appropriate task queue in the temporal server, and waiting for it to finish.

## Temporal Worker in EMCO

To add a new worker the user should follow the template provided in [EMCO source code](https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/tools/workflowclient) or the examples provided in [temporal docs themselves](https://docs.temporal.io/go/how-to-develop-a-worker-program-in-go/)


## Example Temporal Action Controller workflow

To see an example of how temporal action controller operates and executees workflows visit the exammple in examples/test-tac/ and follow the guide there.


## Routes in Temporal

## Temporal action controller (TAC) API

Temporal action controller (TAC) as name suggests is an EMCO action controller. TAC API are for users to provide intents for running workflows at various LCM Hooks like pre-install, post-install, pre-terminate, pre-terminate, pre-update and post-update. During the LCM events like instantiation, terminate, update TAC will be invoked and based on the intents TAC will in turn start workflows.


### Temporal Action Controller API


This API is to add configuration per LCM hook for a DIG. `workflowClient` and `temporal` sections are similar to workflow manage API's. The field `hookType` is used to provide the LCM hook and `hookBlocking` is to specify if wait in TAC is required for the workflow to complete before returning to the orchestrator. `hookBlockingTimeout` field is used if blocking is true.


* Post API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-controller`

Body:

```
metadata:
  name: ABCD
spec:
hookType: pre-install
hookBlocking: true
workflowClient:
  clientEndpointName: ABCDEFGHIJKLMNOPQRSTU
  clientEndpointPort: 121
temporal:
  workflowClientName: ABCD
  workflowStartOptions:
    id: ABCDEFGHIJK
    taskqueue: ABCDEFGHIJKLMNOPQRSTUVW
    workflowexecutiontimeout: 327
    workflowruntimeout: 983
    workflowtasktimeout: 213
    workflowidreusepolicy: 916
    workflowexecutionerrorwhenalreadystarted: true
    retrypolicy:
      initialinterval: 679
      backoffcoefficient: 358.25
      maximuminterval: 658
      maximumattempts: 127
      nonretryableerrortypes: []
  workflowParams:
    activityOptions: {}
    activityParams: {}

```

Before calling the workflow client, TAC will fill in the following `activityParams`

```
    activityParams:
      emco:
        emcoURL: http://1.1.1.1:2
        project: "proj1"
        compositeApp: "c1"
        compositeAppVersion: "v1"
        deploymentIntentGroup: "d1"
        appcontextId: "1234567890999"
```


* Get API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-controller/{tac-intent}`

* Get All API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-controller`

* Delete API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-controller/{tac-intent}`

* Put API:

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-controller/{tac-intent}`

Same body as post


### Temporal cancel API

* Post API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-controller/{tac-intent}/cancel`

Note: Same as the workflow API

### Temporal status API

* Get API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-controller/{tac-intent}/status`

Note: Same as the workflow API