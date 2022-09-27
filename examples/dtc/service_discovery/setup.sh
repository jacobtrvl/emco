#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2021 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

HOST_IP=${HOST_IP:-"oops"}
KUBE_PATH1=${KUBE_PATH1:-"oops"}
KUBE_PATH2=${KUBE_PATH2:-"oops"}
PUBLIC_CLUSTER2=${PUBLIC_CLUSTER2:-"false"}
LC_LEVEL=${LC_LEVEL:-"0"}
HTTP_SERVER_IMAGE_REPOSITORY=${HTTP_SERVER_IMAGE_REPOSITORY:-"<docker-registry-url>/my-custom-httptest-server"}
HTTP_CLIENT_IMAGE_REPOSITORY=${HTTP_CLIENT_IMAGE_REPOSITORY:-"<docker-registry-url>/my-custom-httptest-client"}
# tar files
function create {
    # make the GMS helm charts and profiles
    mkdir -p output
    tar -czf output/http-server.tgz -C ../../helm_charts/http-server/helm http-server
    tar -czf output/http-client.tgz -C ../../helm_charts/http-client/helm http-client
    if [[ ${PUBLIC_CLUSTER2} == "true" ]]; then
        tar -czf output/http-server-profile.tar.gz -C ../../helm_charts/http-server/profile/service_discovery_overrides/public_cluster/http-server-profile
        tar -czf output/http-client-profile.tar.gz -C ../../helm_charts/http-client/profile/service_discovery_overrides/public_cluster/http-client-profile .
    else
        tar -czf output/http-server-profile.tar.gz -C ../../helm_charts/http-server/profile/service_discovery_overrides/private_cluster/http-server-profile .
        tar -czf output/http-client-profile.tar.gz -C ../../helm_charts/http-client/profile/service_discovery_overrides/private_cluster/http-client-profile .
    fi

    cat << NET > values.yaml
KubeConfig1: $KUBE_PATH1
KubeConfig2: $KUBE_PATH2
ProjectName: proj1
CompositeApp: collection-composite-app
DeploymentIntentGroup: collection-deployment-intent-group
HostIP: $HOST_IP
RsyncPort: 30431
DtcPort: 30448
NpsPort: 30438
SdsPort: 30439
HttpServerImageRepository: $HTTP_SERVER_IMAGE_REPOSITORY
HttpClientImageRepository: $HTTP_CLIENT_IMAGE_REPOSITORY
NET
    if [[ ${PUBLIC_CLUSTER2} == "true" ]]; then
        cat << NET >> values.yaml
PublicCluster2: true
NET
    else
        cat << NET >> values.yaml
PublicCluster2: false
NET
    fi
    if [[ ${LC_LEVEL} == "0" ]]; then
        cat << NET >> values.yaml
LogicalCloud: default
NET
    else
        cat << NET >> values.yaml
LogicalCloud: lc1
NET
    fi
    cat << NET > emco-cfg-dtc.yaml
orchestrator:
  host: $HOST_IP
  port: 30415
clm:
  host: $HOST_IP
  port: 30461
dcm:
  host: $HOST_IP
  port: 30477
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
    rm -f emco-cfg-dtc.yaml
    rm -rf output
}

if [ "$#" -lt 1 ] ; then
    usage
    exit
fi

case "$1" in
    "create" )
        if [ "${HOST_IP}" == "oops" ] || [ "${KUBE_PATH1}" == "oops" ] || [ "${KUBE_PATH2}" == "oops" ]; then
            echo -e "ERROR - HOST_IP & KUBE_PATH1 & KUBE_PATH2 environment variables need to be set"
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
