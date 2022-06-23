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
GITHUB_USER=${GITHUB_USER:-"oops"}
GITHUB_TOKEN=${GITHUB_TOKEN:-"oops"}
GITHUB_REPO=${GITHUB_REPO:-"oops"}
OUTPUT_DIR=output
CLIENT_ID=${CLIENT_ID:-"oops"}
TENANT_ID=${TENANT_ID:-"oops"}
CLIENT_SECRET=${CLIENT_SECRET:-"oops"}
SUB_ID=${SUB_ID:-"oops"}
ARC_CLUSTER=${ARC_CLUSTER:-"oops"}
ARC_RG=${ARC_RG:-"oops"}
GIT_BRANCH=${GIT_BRANCH:-"oops"}
TIME_OUT=${TIME_OUT:-"60"}
SYNC_INTERVAL=${SYNC_INTERVAL:-"60"}
RETRY_INTERVAL=${RETRY_INTERVAL:-"60"}

function create_common_values {
    local output_dir=$1
    local host_ip=$2

    create_apps $output_dir
    create_config_file $host_ip

        cat << NET > values.yaml
    PackagesPath: $output_dir
    ProjectName: proj-arc-1
    ClusterProvider: provider-arc
    ClusterLabel: edge-cluster
    AdminCloud: default
    CompositeApp: test-composite-app
    CompositeProfile: test-composite-profile
    GenericPlacementIntent: test-placement-intent
    DeploymentIntent: test-deployment-intent
    Intent: intent
    RsyncPort: 30431
    GacPort: 30433
    OvnPort: 30432
    DtcPort: 30448
    NpsPort: 30438
    HostIP: $host_ip
    APP1: http-server
    APP2: collectd
    APP3: operator
    GitObj: GitObjectFluxRepo
    GithubUser: $GITHUB_USER
    GithubToken: $GITHUB_TOKEN
    GithubRepo: $GITHUB_REPO
    Branch: $GIT_BRANCH
    GitResObj: GitObjectAzure
    ClientID: $CLIENT_ID
    TenantID: $TENANT_ID
    ClientSecret: $CLIENT_SECRET
    SubscriptionID: $SUB_ID
    ArcCluster: $ARC_CLUSTER
    ArcResourceGroup: $ARC_RG
    TimeOut: $TIME_OUT
    SyncInterval: $SYNC_INTERVAL
    RetryInterval: $RETRY_INTERVAL


    Clusters:
      - Name: cluster1
      - Name: cluster2
    K8sClusters:
      - Name: cluster1
        KubeConfig: $KUBE_PATH1
    FluxClusters:
      - Name: cluster2
    Applist:
      - Name: collectd
        Cluster:
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
        if [ "${GITHUB_TOKEN}" == "oops"  ] ; then
            echo -e "ERROR - GITHUB_TOKEN must be defined"
            exit
        fi
        if [ "${GITHUB_REPO}" == "oops"  ] ; then
            echo -e "ERROR - GITHUB_REPO must be defined"
            exit
        fi
        if [ "${GITHUB_USER}" == "oops" ] ; then
            echo -e "GITHUB_USER must be defined"
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
