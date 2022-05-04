## Temporal action controller (TAC) API

Temporal action controller (TAC) as name suggests is an EMCO action controller. TAC API are for users to provide intents for running workflows at various LCM Hooks like pre-install, post-install, pre-terminate, pre-terminate, pre-update and post-update. During the LCM events like instantiation, terminate, update TAC will be invoked and based on the intents TAC will in turn start workflows.

### Temporal action intent API

#### Intent API

Open: Is there any common items in Temoral configuration that can be moved in this API? If not then this API can be removed and only hook api can be used.

* Post API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-intents`

Body:

```
metadata:
  name: ABCD
  description: tac intent
spec:
  emcoURL: http://1.1.1.1:2

```

* Get API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-intents/{temporalActionIntent}`

* Get All API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-intents`

* Delete API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-intents/{temporalActionIntent}`


#### Hook API


This API is to add configuration per LCM hook for a DIG. `workflowClient` and `temporal` sections are similar to workflow manage API's. The field `hookType` is used to provide the LCM hook and `hookBlocking` is to specify if wait in TAC is required for the workflow to complete before returning to the orchestrator. `hookBlockingTimeout` field is used if blocking is true.


* Post API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-intents/{temporalActionIntent}/hooks`

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

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-intents/{temporalActionIntent}/hooks/{hook}`

* Get All API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-intents/{temporalActionIntent}/hooks`

* Delete API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-intents/{temporalActionIntent}/hooks/{hook}`

* Put API:

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-intents/{temporalActionIntent}/hooks/{hook}`

Same body as post


### Temporal cancel API

* Post API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-intents/{temporalActionIntent}//hooks/{hook}/cancel`

Note: Same as the workflow API

### Temporal status API

* Get API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-intents/{temporalActionIntent}//hooks/{hook}/status`

Note: Same as the workflow API


## New Orchestrator API to support workflows

Orchestrator API to get the appcontext information expected to be used by the temporal workflows.

* Get all applications in an appcontext

URL: `v2/appcontext/{appContextId}/app`

* Get all clusters for an application in an appcontext

URL: `v2/appcontext/{appContextId}/app/{appName}/cluster`

* Get all resources for a cluster for an application in an appcontext

URL: `v2/appcontext/{appContextId}/app/{appName}/cluster/{cluster}/resource`

* Get a resources for a cluster for an application in an appcontext

URL: `v2/appcontext/{appContextId}/app/{appName}/cluster/{cluster}/resource/{resource}`

