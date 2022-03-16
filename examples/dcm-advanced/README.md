[//]: # "SPDX-License-Identifier: Apache-2.0"
[//]: # "Copyright (c) 2020-2022 Intel Corporation"

# Advanced DCM (Logical Cloud) operations

The examples herein provide an easy way to test add/removing of clusters to an already-instantiated Logical Cloud.

Here's one possible order of commands to test this:

    emcoctl --config emco-cfg.yaml apply -f prerequisites.yaml -v values.yaml
    emcoctl --config emco-cfg.yaml apply -f 1stcluster.yaml -v values.yaml 
    emcoctl --config emco-cfg.yaml apply -f instantiate-lc.yaml -v values.yaml
    emcoctl --config emco-cfg.yaml apply -f 2ndcluster.yaml -v values.yaml 
    curl -X PUT http://localhost:9077/v2/projects/proj1/logical-clouds/lc1
    emcoctl --config emco-cfg.yaml delete -f instantiate-lc.yaml -v values.yaml 
    emcoctl --config emco-cfg.yaml delete -f 1stcluster.yaml -v values.yaml
    emcoctl --config emco-cfg.yaml delete -f 2ndcluster.yaml -v values.yaml 
    emcoctl --config emco-cfg.yaml delete -f prerequisites.yaml -v values.yaml 
