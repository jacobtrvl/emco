#### SPDX-License-Identifier: Apache-2.0
#### Copyright (c) 2020 Intel Corporation

#################################################################
# Running EMCO dependency testcase with Helm Hook support (using emcoctl)
#################################################################

For testing inter app dependency with Helm hook using ngnix, collectd and operatior apps using 2 clusters

Test Setup has 3 apps and 2 clusters with following setup:
 * collectd (Cluster1 and Cluster2)
 * Operator (Cluster1)
 * Ngnix (Cluster1 and Cluster2). This chart has hooks.

The test setup has following App Dependency:
 * collectd --> Operator --> Ngnix

## Setup Test Environment to run test cases

1. export environment variables

a) KUBE_PATH1, KUBE_PATH2,

b) HOST_IP: IP address of the cluster where EMCO is installed

#### NOTE: For HOST_IP, assuming here that nodeports are used to access all EMCO services both from outside and between the EMCO services.

2. Setup script

    Run setup.sh script

    ```
        ./setup.sh create
    ```

    Output of this command are 1) values.yaml file and  2) emco_cfg.yaml 3) Helm chart and profiles tar.gz files for all the usecases.

    values.yaml is created with the desired test setup as described above.


## Create all resources and instatiate a
1. Apply 00-controllers.yaml and 01-prerequisites.yaml. This is required for all the tests. This creates controllers, one project, number of  clusters and default admin logical cloud. Create deployment resources add app dependency and then run instantiation:

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f 00-controllers.yaml -v values.yaml`

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f 01-prerequisites.yaml -v values.yaml`

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f 02-deployment-intent.yaml  -v values.yaml`

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f 03-dependency.yaml  -v values.yaml`

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f 04-instantiation.yaml  -v values.yaml`


### Expected outcome

* Nginx app will come up on both clusters. For the Nginx App there are 2 hooks: Pre-install and Post-install. Pre-install should complete and then the main resources including Nginx pod will comeup. After the Pod is in running state there is a wait time of 10 secs as specified in the 03-dependency.yaml.
* Operator will next come up on cluster1. 4 Pods come up for this app. After all the Pods are up, there is a wait time of 10 sec as specified in the 03-dependency.yaml.
* Next collectd app will come up on cluster1 and cluster2

2. Cleanup

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f 04-instantiation.yaml  -v values.yaml`

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f 03-dependency.yaml  -v values.yaml`

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f 02-deployment-intent.yaml  -v values.yaml`

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f 01-prerequisites.yaml -v values.yaml`

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f 00-controllers.yaml -v values.yaml`


3. Cleanup generated files

    `$ ./setup.sh cleanup`

#### NOTE: Known issue with the test cases: Deletion of the resources fails sometimes as some resources can't be deleted before others are deleted. This can happen due to timing issue. In that case try deleting again and the deletion should succeed.
