#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

HOST_IP=${HOST_IP:-"oops"}
KUBE_PATH=${KUBE_PATH:-"oops"}

collectd_folder=../helm_charts/collectd
operator_latest_folder=../helm_charts/operators-latest

function create {
    mkdir -p output
    tar -czf output/collectd.tar.gz -C $collectd_folder/helm .
    tar -czf output/collectd_profile.tar.gz -C $collectd_folder/profile .
    tar -czf output/operator.tar.gz -C $operator_latest_folder/helm .
    tar -czf output/operator_profile.tar.gz -C $operator_latest_folder/profile .

# head of values.yaml
cat << NET > values.yaml
ProjectName: proj1
ClusterProvider: provider1
Cluster: cluster1
ClusterLabel: edge-cluster1
ClusterRef: cluster1-ref
AdminCloud: default
App1: operator
App2: collectd
KubeConfig: $KUBE_PATH
HelmApp1: output/operator.tar.gz
HelmApp2: output/collectd.tar.gz
ProfileApp1: output/operator_profile.tar.gz
ProfileApp2: output/collectd_profile.tar.gz
CompositeProfile: collection-composite-profile
GenericPlacementIntent: collection-placement-intent
DeploymentIntent: collection-deployment-intent-group
CompositeAppGac: gac-composite-app
GacIntent: operator-gac-intent
IstioIngressGatewayKvName: istioingressgatewaykvpairs
DatabaseAuthorizationKvName: databaseauthorizationkvpairs
RsyncPort: 30431
GacPort: 30433
DtcPort: 30448
NpsPort: 30438
HostIP: $HOST_IP
NET

# head of emco-cfg.yaml
cat << NET > emco-cfg.yaml
orchestrator:
  host: $HOST_IP
  port: 30415
  statusPort: 30416
clm:
  host: $HOST_IP
  port: 30461
dcm:
  host: $HOST_IP
  port: 30477
  statusPort: 30478
gac:
  host: $HOST_IP
  port: 30420
dtc:
  host: $HOST_IP
  port: 30418
NET
}

function usage {
    echo "Usage: $0  create|cleanup"
}

function cleanup {
    rm -f *.tar.gz
    rm -f values.yaml
    rm -f emco-cfg.yaml
    rm -rf output
}

if [ "$#" -lt 1 ] ; then
    usage
    exit
fi

case "$1" in
    "create" )
        if [ "${HOST_IP}" == "oops" ] || [ "${KUBE_PATH}" == "oops" ] ; then
            echo -e "ERROR - HOST_IP & KUBE_PATH environment variable needs to be set"
        else
            create
        fi
        ;;
    "cleanup" )
        cleanup
    ;;
    *)
        usage ;;
esac
