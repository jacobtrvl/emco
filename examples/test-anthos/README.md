[//]: # "SPDX-License-Identifier: Apache-2.0"
[//]: # "Copyright (c) 2020-2022 Intel Corporation"

# Running Google Anthos testcase with emcoctl

This folder contains the collectd testcase to be run with EMCO deployed to Google Anthos clusters. This test assumes that a GKE cluster with Anthos Config Management enabled has been created as mentioned in https://gitlab.com/project-emco/core/emco-base/-/tree/main/docs/design/gitops_support.md.

See https://cloud.google.com/anthos-config-management/docs for Google Anthos-specific documentation.

## Setup Environment variables

Setup environment variables as mentioned in https://gitlab.com/project-emco/core/emco-base/-/tree/main/docs/design/gitops_support.md.

## Generating files

Creates artifacts needed to run the testcase.

```
$ cd emco-base/examples/test-anthos
```
```
$ ./setup.sh create
```

## Creating the controllers

Apply 00-controllers.yaml, this creates the controllers.

```
$ emcoctl --config emco-cfg.yaml apply -f 00-controllers.yaml -v values.yaml
```

## Applying the prerequisites

Apply 01-prerequisites.yaml.

```
$ emcoctl --config emco-cfg.yaml apply -f 01-prerequisites.yaml -v values.yaml
```

## Create the logical cloud

Apply 02-logicalcloud.yaml. This creates a Privileged Logical Cloud.

```
$ emcoctl --config emco-cfg.yaml apply -f 02-logicalcloud.yaml -v values.yaml -s
```

After this step, we should see resources created under the clusters/ directory of the git repo, and Logical Cloud resources will show up in the Google Anthos cluster.

## Create the deployment intent

Apply 03-deployment-intent.yaml.

```
$ emcoctl --config emco-cfg.yaml apply -f 03-deployment-intent.yaml -v values.yaml
```

## Running the test case

```
$ emcoctl --config emco-cfg.yaml apply -f 04-instantiation.yaml -v values.yaml -s
```

After this step, we should see resources created under the namespaces/ directory of the git repo, and DIG resources will show up in the Google Anthos cluster.


## Cleanup

1. Terminate the DIG.

    ```
    $ emcoctl --config emco-cfg.yaml delete -f 04-instantiation.yaml -v values.yaml -s
    ```

2. Cleanup deployment intent resources.

    ```
    $ emcoctl --config emco-cfg.yaml delete -f 03-deployment-intent.yaml -v values.yaml
    ```

3. Terminate and delete Logical Cloud

    ```
    $ emcoctl --config emco-cfg.yaml delete -f 02-logicalcloud.yaml -v values.yaml
    ```

    Note: you may need to repeat the command a second time, after a few seconds, due to the delay in the terminate operation.

3. Cleanup prerequisites

    ```
    $ emcoctl --config emco-cfg.yaml delete -f 01-prerequisites.yaml -v values.yaml
    ```

4. Cleanup controllers

    ```
    $ emcoctl --config emco-cfg.yaml delete -f 00-controllers.yaml -v values.yaml
    ```

5. Cleanup generated files

    ```
    $ ./setup.sh cleanup
    ```
