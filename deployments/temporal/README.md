[//]: # "SPDX-License-Identifier: Apache-2.0"
[//]: # "Copyright (c) 2020-2022 Intel Corporation"

# Deployments - Temporal

This role of this document is to educate the user on how to deploy temporal clients and workers to their machine using docker and helm.

This document assumes the user has a basic understanding of temporal, if not please visit [temporal's site](https://docs.temporal.io/), and familiarize
themselves with the software.

## Workflow Client

The workflow client is written in a way where the user should never need to write their own version.
If something arises where the user must use another version follow this [doc](https://gitlab.com/project-emco/core/emco-base/-/blob/main/docs/developer/Temporal_Workflow_Client.md) 
to understand how to write your own.

To deploy the workflow client into kuberenetes the user must write in the following
information into the values.yaml in [./workflowclient/helm/values.yaml](./workflowclient/helm/values.yaml).
There are four values you need to make sure are filled out, and they are:

 - temporalServer: This is the address of the temporal server you are using.
 - temporalPort: This is the port that your temporal server is using. It is set to the default 7223.
 - containerPort: This is the ingress port for the workflow client container itself. It is set to the deafult 9090. You shouldn't need to change it.
 - repository: This is the container repo where your container is stored. If you used the Makefile it should be in your usual main emco repo.

On top of these values inside of the helm charts, also make sure that your EMCODOCKERREPO and MAINDOCKERREPO are set.

Once all of these values are set correctly it should be as easy as packaging it, and running it. To do so follow the following steps:

1. Navigate to workflowclient folder. Use the makefile to compile the code, and package it in a docker image.

```
$ cd workflowclient
$ make all
```

2. Once it is packaged into a docker image use the deploy_client.sh script to package the helm charts for you, and put the output into emco-base's bin directory.

```
$ ./deploy_client.sh
```

3. Finally navigate to the temporal folder inside of emco-base's bin directory, and run the temporal-install.sh script to deploy the workflowclient

```
$ cd ${EMCO_DIR}/bin/temporal
$ ./temporal-install.sh install
```