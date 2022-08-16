[//]: # "SPDX-License-Identifier: Apache-2.0"
[//]: # "Copyright (c) 2020-2022 Intel Corporation"

## Example worker.

Quick readme to describe bringing up the example worker, so 
the test-tac workflow can be tested easier.

1. in the example-worker folder type "make all" into the terminal
2. tag the docker image that was made, and push it to your correct repo
3. go to deployments/helm/worker/values.yaml and edit temporalServer to have your temporal server ip and repository to be pointing to the docker image you just tagged and pushed.
4. inside of deployments/helm/ use the scripts up.sh and down.sh to bring up the worker in the demo namespace of kubernetes.