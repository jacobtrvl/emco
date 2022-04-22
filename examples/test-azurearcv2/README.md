[//]: # "SPDX-License-Identifier: Apache-2.0"
[//]: # "Copyright (c) 2020-2022 Intel Corporation"

# Running Azure Arc testcase with emcoctl

This folder contains the collectd testcase to be run with EMCO deployed to Azure Arc clusters. This test assumes that an azure arc cluster has been created as mentioned in https://gitlab.com/project-emco/core/emco-base/-/tree/main/docs/design/gitops_support.md.


## Setup Environment variables

Setup environment variables as mentioned in https://gitlab.com/project-emco/core/emco-base/-/tree/main/docs/design/gitops_support.md.

## Generating files

Creates artifacts needed to run the testcase.

```
$ cd emco-base/examples/test-azurearcv2
```
```
$ ./setup.sh create
```

## Applying the prerequisites

Apply 00-controllers.yaml, this creates the controllers. This step is required to be done only once.

```
$ emcoctl --config emco-cfg.yaml apply -f 00-prerequisites.yaml -v values.yaml
```

## Instantiating the Logical Cloud

Apply 01-prerequisites.yaml.

```
$ emcoctl --config emco-cfg.yaml apply -f 01-logical-cloud.yaml -v values.yaml
```

## Create the deployment intent

Apply 02-deployment-intent.yaml.

```
$ emcoctl --config emco-cfg.yaml apply -f 02-deployment-intent.yaml -v values.yaml
```

## Running the test case

```
$ emcoctl --config emco-cfg.yaml apply -f 03-instantiation.yaml -v values.yaml
```

After this step, we should see the resources created in the Azure Arc cluster.


## Cleanup

1. Delete usecase

    ```
    $ emcoctl --config emco-cfg.yaml delete -f 03-instantiation.yaml -v values.yaml
    ```

2. Cleanup deployment intent.

    ```
    $ emcoctl --config emco-cfg.yaml delete -f 02-deployment-intent.yaml -v values.yaml
    ```

3. Terminate Logical Cloud

    ```
    $ emcoctl --config emco-cfg.yaml delete -f 01-logical-cloud.yaml -v values.yaml
    ```

4. Cleanup prerequisites

    ```
    $ emcoctl --config emco-cfg.yaml delete -f 00-prerequisites.yaml -v values.yaml
    ```

5. Cleanup generated files

    ```
    $ ./setup.sh cleanup
    ```

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
    ```