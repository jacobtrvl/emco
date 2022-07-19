```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2022 Intel Corporation
```

# TAC Deploying Worker To Temporal

This document will be an overview of the design
for TAC's ability to automatically deploy workers
on behalf of the user before it would execute any of its hooks.

### Assumptions

This scenario will assume a few things about the state of
EMCO when the possible deployment of workers would take place.

 - We would assume there is a DIG out there somewhere containing the Worker that
 the TAC intent would call.
 - We would assume this DIG as already been approved, but has not been instantiatied
 by anyone.
 

 Assuming these two items we can make a proposed solution to this issue.

 ### Solution

The solution would be to add logic that executes everytime before any
intents are sent to temporal. TAC would look up the all of the relevant
information needed to query the DIG related to the worker their intents
are trying execute. 

Once it has this information it will query each DIG to get the current status
of them. TAC will have four possible scenarios it could encounter, and will have these responses:

 - DNE or Exists but isn't approved: fail, and send an error back to orch.
 - Approved: fail, and send an error back to orch, but start the worker and let the user know.
 - Creating: fail, and send error back to orch, but let the user know the worker is still spinning up.
 - Exists: succeed, and move on to the next worker to check or move onto the intents.

For this possible solution TAC needs more information. It would need a database that could
tie some identifying information it already has to a DIG. To do this we could add a few new
endpoints to TAC that would create a new mongo db entry that would tie task queue to DIG. The
next section will cover those new APIs.

Question: what if the Worker DIG has resources that we don't handle 'ready' status for?

### API

* Post API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-controller/{tac-intent}/workers`

Body:

```
metadata:
  name: ABCD
  description: tac intent
spec:
  taskQueue: MIGRATION_WORKFLOW
  dig: digOne
  cApp: cAppone
  project: projectOne
  cAppVersion: cAppVersionOne

```

* Get Many API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-controller/{tac-intent}/workers/`

* Get One API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-controller/{tac-intent}/workers/{worker}`

* Put API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-controller/{tac-intent}/workers/{worker}`

Body:

```
metadata:
  name: ABCD
  description: tac intent
spec:
  taskQueue: MIGRATION_WORKFLOW
  dig: digOne
  cApp: cAppone
  project: projectOne
  cAppVersion: cAppVersionOne

```

Question: Should we assume these are in the same project?

* Delete API

URL: `v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-controller/{tac-intent}/workers/{worker}`



### Updated Solution

Boil situations down to 3 possible scenarios.


1. There is a registered DIG for the task queue, and it has been approved, and we can guarantee that it is the only item in the DIG.
  * In this case we would instantiate the DIG and move onto next item.
2. There is a registered DIG for the task queue, but it has not been approved, or it is not the only item inside of the DIG
 * In this case we would fail and report why. We don't want to start an unapproved item, and we don't know how long the items would take to start.
3. There is no registered DIG for the task queue.
 * We assume the user has it dealt with externally and continue on. Maybe print a warning about it.

The action controller should bring the workers down every time as well. The temporal workers are built to be lightweight, so registering them should not take too much time at all.
In the "pre" stage of the controller it will execute the logic above, and spin up any workers it might need to. Then it will execute all the workflows it might need to. For both Pre and Post.
In the post part of the controller it will execute the remaining workflows then complete a "cleanup" cycle where it will tell the registered DIG to delete, and clean up the worker.

Q: Same workers used in multiple workflows?
Q: Registering multiple workers to one workflow?
Q: Easy way to check if the DIG only contains temporal worker?