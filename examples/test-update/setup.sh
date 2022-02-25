#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail


source ../multi-cluster/_common.sh

test_folder=../../tests/
demo_folder=../../demo/
deployment_folder=../../../deployments/

HOST_IP=${HOST_IP:-"oops"}
KUBE_PATH1=${KUBE_PATH1:-"oops"}
KUBE_PATH2=${KUBE_PATH2:-"oops"}
OUTPUT_DIR=output

function create_common_values {
    local output_dir=$1
    local host_ip=$2

    create_apps $output_dir
    create_config_file $host_ip

        cat << NET > values.yaml
    PackagesPath: $output_dir
    ProjectName: proj-update-1
    ClusterProvider: provider-update
    ClusterLabel: edge-cluster
    AdminCloud: default
    CompositeApp: test-composite-app
    CompositeProfile: test-composite-profile
    GenericPlacementIntent: test-placement-intent
    DeploymentIntent: test-deployment-intent
    Intent: intent
    RsyncPort: 30441
    GacPort: 30493
    OvnPort: 30473
    DtcPort: 30483
    NpsPort: 30485
    HostIP: $host_ip
    APP1: collectd
    Clusters:
      - Name: cluster1
        KubeConfig: $KUBE_PATH1
      - Name: cluster2
        KubeConfig: $KUBE_PATH2
    Applist:
      - Name: collectd
        Cluster:
          - cluster1
        UpdateCluster:
          - cluster2

NET
}


function usage {
    echo "Usage: $0 create|cleanup"
}

function cleanup {
    rm -f *.tar.gz
    rm -f values.yaml
    rm -f emco-cfg.yaml
    rm -rf $OUTPUT_DIR
}

case "$1" in
    "create" )
        if [ "${HOST_IP}" == "oops" ] ; then
            echo -e "ERROR - Environment varaible HOST_IP must be set"
            exit
        fi
        if [ "${KUBE_PATH1}" == "oops"  ] ; then
            echo -e "ERROR - KUBE_PATH1 must be defined"
            exit
        fi
        if [ "${KUBE_PATH2}" == "oops" ] ; then
            echo -e "KUBE_PATH2 must be defined"
        else
            create_common_values $OUTPUT_DIR $HOST_IP
            echo "Done create!!!"
        fi
        ;;
    "cleanup" )
        cleanup
    ;;
    *)
        usage ;;
esac
