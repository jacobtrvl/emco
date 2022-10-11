#### SPDX-License-Identifier: Apache-2.0
#### Copyright (c) 2020 Intel Corporation

#################################################################
# Running EMCO multicluster testcases with emcoctl
#################################################################

This tests supports 10 clusters and a composit app with 3 applications. The applications can be one of the following:
collectd, prometheus-operator, operator, http-client, http-server

## Setup Test Environment to run test cases

1. export environment variables

a) KUBE_PATH1, KUBE_PATH2, KUBE_PATH3,... upto KUBE_PATH10 where the kubeconfig for each of the edge cluster is located. Atleast KUBE_PATH1 needs to be defined and

b) HOST_IP: IP address of the cluster where EMCO is installed

#### NOTE: For HOST_IP, assuming here that nodeports are used to access all EMCO services both from outside and between the EMCO services.

2. Setup script

    Run setup.sh script for creating artifacts needed to test EMCO on multiple clusters. From the command line setup script takes up to 3 applications using flags -a, -b, -c. Along with the name of the app, the clustes on which these apps need to installed can also be specified.

    ```
        ./setup.sh -a <appname>:<cluster1>:<cluster2>:.....<cluster10> -b <appname>:<cluster1>:<cluster2>:.....<cluster10> -c <appname>:<cluster1>:<cluster2>.....<cluster10> create
    ```

    For example: this will create a composite app with 2 apps collectd and operator and collectd to be installed on cluster1 and cluster2 and operator on cluster1

    `$ ./setup.sh  -a collectd:cluster1:cluster2 -b operator:"cluster1" create`

    Output of this command are 1) values.yaml file for the current environment 2) emco_cfg.yaml for the current environment and 3) Helm chart and profiles tar.gz files for all the usecases.

    Cleanup artifacts generated above with cleanup command

    `$ ./setup.sh cleanup`

## Create Prerequisites to run Tests
1. Apply 00-controllers.yaml and 01-prerequisites.yaml. This is required for all the tests. This creates controllers, one project, number of  clusters and default admin logical cloud. This step is required to be done only once for all usecases:

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f 00-controllers.yaml -v values.yaml`

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f 01-prerequisites.yaml -v values.yaml`

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f 02-instantiate-lc.yaml -v values.yaml`

## Running test cases

1. Run the testcase

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f 03-deployment-intent.yaml -v values.yaml`

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f 04-deployment-instantiate.yaml -v values.yaml`

2. Cleanup

    `$ $bin/emcoctl --config emco-cfg.yaml delete -f 04-deployment-instantiate.yaml -v values.yaml`

    `$ $bin/emcoctl --config emco-cfg.yaml delete -f 03-deployment-intent.yaml -v values.yaml`

    `$ $bin/emcoctl --config emco-cfg.yaml delete -f 02-instantiate-lc.yaml -v values.yaml`

    `$ $bin/emcoctl --config emco-cfg.yaml delete -f 01-prerequisites.yaml -v values.yaml`

    `$ $bin/emcoctl --config emco-cfg.yaml delete -f 00-controllers.yaml -v values.yaml`


3. Cleanup generated files

    `$ ./setup.sh cleanup`

#### NOTE: Known issue with the test cases: Deletion of the resources fails sometimes as some resources can't be deleted before others are deleted. This can happen due to timing issue. In that case try deleting again and the deletion should succeed.
