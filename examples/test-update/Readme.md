#### SPDX-License-Identifier: Apache-2.0
#### Copyright (c) 2022 Intel Corporation

#################################################################
# Running Update API test (using emcoctl)
#################################################################

For testing update API

Test Setup has 1 apps and 2 clusters with following setup:
 * collectd (Cluster1)

After running update collectd app is removed from cluster1 and brought up on cluster2.


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


## Create all resources and instantiate app

Apply 00-controllers.yaml and 01-prerequisites.yaml. This is required for all the tests. This creates controllers, one project, number of  clusters and default admin logical cloud. Create deployment resource and then run instantiation:

    $ $bin/emcoctl --config emco-cfg.yaml apply -f 00-controllers.yaml -v values.yaml

    $ $bin/emcoctl --config emco-cfg.yaml apply -f 01-prerequisites.yaml -v values.yaml

    $ $bin/emcoctl --config emco-cfg.yaml apply -f 02-deployment-intent.yaml -v values.yaml

    $ $bin/emcoctl --config emco-cfg.yaml apply -f 03-instantiation.yaml  -v values.yaml

### Update App

     $ $bin/emcoctl --config emco-cfg.yaml update -f 04-update-deployment-intent.yaml  -v values.yaml

     $ $bin/emcoctl --config emco-cfg.yaml apply -f 05-update.yaml -v values.yaml

### Rollback App

     $ $bin/emcoctl --config emco-cfg.yaml apply -f 06-rollback.yaml -v values.yaml


## Delete the application

Delete all the resources in the reverse order.

    $ $bin/emcoctl --config emco-cfg.yaml delete -f 03-instantiation.yaml  -v values.yaml

    $ $bin/emcoctl --config emco-cfg.yaml delete -f 02-deployment-intent.yaml -v values.yaml

    $ $bin/emcoctl --config emco-cfg.yaml delete -f 01-prerequisites.yaml -v values.yaml

    $ $bin/emcoctl --config emco-cfg.yaml delete -f 00-controllers.yaml -v values.yaml

## Cleanup generated files

    $ ./setup.sh cleanup

#### NOTE: Known issue with the test cases: Deletion of the resources fails sometimes as some resources can't be deleted before others are deleted. This can happen due to timing issue. In that case try deleting again and the deletion should succeed.

