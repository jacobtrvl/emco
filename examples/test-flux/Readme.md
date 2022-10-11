[//]: # "SPDX-License-Identifier: Apache-2.0"
[//]: # "Copyright (c) 2022 Intel Corporation"

# Running EMCO with Flux v2 based cluster (using emcoctl)

For testing an application (collectd) deployed on 2 clusters - a Flux v2 based cluster and a direct access cluster.

## Setup Test Environment to run test cases

1. Export environment variables:

    - For direct-access clusters, export `KUBE_PATH1`

    - For Flux v2 based clusters export variables `GIT_USER`, `GIT_TOKEN`, `GIT_URL`, `GIT_BRANCH` and `GIT_REPO`

    - `HOST_IP`: IP address of the cluster where EMCO is installed

    - `LOGICAL_CLOUD_LEVEL`: Set it to "admin" to use admin(default) logical cloud and set it to "standard" to use standard logical cloud.

    > NOTE: For `HOST_IP`, assuming here that nodeports are used to access all EMCO services both from outside and between the EMCO services.

    Example variables for GitHub repository https://github.com/myusername/emcoflux:

    ```
    export GIT_OWNER=myusername
    export GIT_REPO=emcoflux
    export GIT_TOKEN=<token previously obtained from GitHub API>
    export GIT_BRANCH=main
    export GIT_URL=https://github.com/myusername/emcoflux
    ```
2. Select API type to interact with Git server

    By default the example uses the core Git APIs for interacting with the Git Server. To switch to GitHub REST APIs uncomment the gitType line in templates/prerequisites-common.yaml

    ```
    version: emco/v2
    resourceContext:
    anchor: cluster-providers/{{ $.ClusterProvider }}/cluster-sync-objects
    metadata:
    name: {{ $.GitObj}}
    spec:
    kv:
    #- gitType: github  # Uncomment to use GitHub Rest API
    - userName: {{ $.GitUser }}
    - gitToken:  {{ $.GitToken }}
    - repoName: {{ $.GitRepo }}
    - branch: {{ $.GitBranch }}
    - url: {{ $.GitUrl }}
    ```

3. Setup script:

    Run setup.sh script

    ```
        ./setup.sh create
    ```

    Output of this command are 1) values.yaml file and  2) emco_cfg.yaml 3) 00-prerequisites.yaml  4) Helm chart and profiles tar.gz files for all the usecases.

    values.yaml is created with the desired test setup as described above.


4. Configure a Kubernetes cluster to work with Flux v2 support:

    Install flux using the fluxctl instructions. (https://fluxcd.io/docs/installation/#github-and-github-enterprise). Provide the name of the cluster to be of the format providername+clustername format. In the following example the provider name registered with EMCO is provider1flux and the cluster name registerd with EMCO is cluster2.

    ```
    $ flux bootstrap github --owner=$GIT_USER --repository=$GIT_REPO --branch=main --path=./clusters/provider1flux+cluster2 --personal
    ```

    On the edge cluster that supports Flux v2 install monitor like the example below. Note the name of the cluster. That name should match the name provided above.

    By default monitor uses the core Git APIs for interacting with Git Server.
    ```
    $ helm install --kubeconfig $KUBE_PATH --set git.token=$GIT_TOKEN --set git.repo=$REPO --set git.username=$OWNER --set git.clustername="provider1flux+cluster2" --set git.gitbranch="main" --set git.giturl=$GIT_URL --set git.enabled=true -n emco monitor .
    ```
    To use GitHub REST APIs for interacting with Github server set git.gittype=github
    ```
    $ helm install --kubeconfig $KUBE_PATH --set git.token=$GIT_TOKEN --set git.repo=$REPO --set git.username=$OWNER --set git.clustername="provider1flux+cluster2" --set git.gittype=github --set git.gitbranch="main" --set git.giturl=$GIT_URL --set git.enabled=true -n emco monitor .
    ```
    > NOTE: If using the GitHub REST APIs for interacting with Github server ensure that the monitor as well as the resources are installed with gittype set to github.

## Create all resources and instantiate

1. Apply 00-prerequisites.yaml and 01-logical-cloud.yaml. This is required for all the tests. This creates controllers, one project, number of  clusters and default admin logical cloud. Create deployment resources and then run instantiation:

    ```
    $ emcoctl --config emco-cfg.yaml apply -f 00-prerequisites.yaml -v values.yaml
    ```
    ```
    $ emcoctl --config emco-cfg.yaml apply -f 01-logical-cloud.yaml -v values.yaml
    ```
    ```
    $ emcoctl --config emco-cfg.yaml apply -f 02-deployment-intent.yaml  -v values.yaml
    ```
    ```
    $ emcoctl --config emco-cfg.yaml apply -f 03-instantiation.yaml  -v values.yaml
    ```

    **Expected outcome:**
    * Collectd app gets deployed on direct access cluster.
    * Collectd app gets deployed on cluster with Flux v2 support. This may take more time than the direct access cluster as the deployment happens through a Git location.

2. Cleanup

    ```
    $ emcoctl --config emco-cfg.yaml delete -f 03-instantiation.yaml  -v values.yaml
    ```
    ```
    $ emcoctl --config emco-cfg.yaml delete -f 02-deployment-intent.yaml  -v values.yaml
    ```
    ```
    $ emcoctl --config emco-cfg.yaml delete -f 01-logical-cloud.yaml -v values.yaml
    ```
    ```
    $ emcoctl --config emco-cfg.yaml delete -f 00-prerequisites.yaml -v values.yaml
    ```

3. Cleanup generated files

    ```
    $ ./setup.sh cleanup
    ```

    > NOTE: Known issue with the test cases: Deletion of the resources fails sometimes as some resources can't be deleted before others are deleted. This can happen due to timing issue. In that case try deleting again and the deletion should succeed.

## Running testcase using test-aio.sh script
The test-aio.sh script can be used to simplify the process of running a test case. It makes use of the status query APIs to ensure that the logical cloud and deployment intent group are instantiated (During creation) or terminated (During deletion) before moving to the next step.

To run the test case run the following commands after setting up the environment variables. There is no need to run the setup.sh script as it's taken care of by the test-aio.sh script.

1. Apply the testcase

    ```
    $ ./test-aio.sh apply
    ```
2. Delete the testcase

    ```
    $ ./test-aio.sh delete