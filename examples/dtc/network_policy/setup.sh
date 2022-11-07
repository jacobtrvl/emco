#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2021 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

HOST_IP=${HOST_IP:-"oops"}
KUBE_PATH1=${KUBE_PATH1:-"oops"}
KUBE_PATH2=${KUBE_PATH2:-"oops"}
IMAGE_REPOSITORY=${IMAGE_REPOSITORY:-${EMCODOCKERREPO%/}}
HTTP_SERVER_IMAGE_REPOSITORY=${HTTP_SERVER_IMAGE_REPOSITORY:-"${IMAGE_REPOSITORY}/my-custom-httptest-server"}
HTTP_CLIENT_IMAGE_REPOSITORY=${HTTP_CLIENT_IMAGE_REPOSITORY:-"${IMAGE_REPOSITORY}/my-custom-httptest-client"}
# tar files
function create {
    # make the GMS helm charts and profiles
    mkdir -p output
    tar -czf output/http-server.tgz -C ../../helm_charts/http-server/helm http-server
    tar -czf output/http-client.tgz -C ../../helm_charts/http-client/helm http-client
    tar -czf output/http-server-profile.tar.gz -C ../../helm_charts/http-server/profile/network_policy_overrides/http-server-profile .
    tar -czf output/http-client-profile.tar.gz -C ../../helm_charts/http-client/profile/network_policy_overrides/http-client-profile .

    cat << NET > values.yaml
KubeConfig1: $KUBE_PATH1
ProjectName: proj1
LogicalCloud: default
CompositeApp: collection-composite-app
DeploymentIntentGroup: collection-deployment-intent-group
HostIP: $HOST_IP
RsyncPort: 30431
DtcPort: 30448
NpsPort: 30438
HttpServerImageRepository: $HTTP_SERVER_IMAGE_REPOSITORY
HttpClientImageRepository: $HTTP_CLIENT_IMAGE_REPOSITORY
NET
    if [[ ${KUBE_PATH2} != "oops" ]]; then
	cat << NET >> values.yaml
KubeConfig2: $KUBE_PATH2
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
        if [ "${HOST_IP}" == "oops" ] || [ "${KUBE_PATH1}" == "oops" ]; then
            echo -e "ERROR - HOST_IP & KUBE_PATH1 environment variables need to be set"
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
