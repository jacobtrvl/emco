#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

HOST_IP=${HOST_IP:-"oops"}
KUBE_PATH1=${KUBE_PATH1:-"oops"}
HTTP_SERVER_IMAGE_REPOSITORY=${HTTP_SERVER_IMAGE_REPOSITORY:-"<docker-registry-url>/my-custom-httptest-server"}
HTTP_CLIENT_IMAGE_REPOSITORY=${HTTP_CLIENT_IMAGE_REPOSITORY:-"<docker-registry-url>/my-custom-httptest-client"}

function create {
    # make the SDEWAN helm charts and profiles
    mkdir -p output
    tar -czf output/http-server.tgz -C ../../helm_charts/http-server/helm http-server
    tar -czf output/http-client.tgz -C ../../helm_charts/http-client/helm http-client
    tar -czf output/http-server-profile.tar.gz -C ../../helm_charts/http-server/profile/sdewan_overrides/http-server-profile .
    tar -czf output/http-client-profile.tar.gz -C ../../helm_charts/http-client/profile/sdewan_overrides/http-client-profile .

    cat << NET > values.yaml
HostIP: $HOST_IP
KubeConfig1: $KUBE_PATH1
HttpServerImageRepository: $HTTP_SERVER_IMAGE_REPOSITORY
HttpClientImageRepository: $HTTP_CLIENT_IMAGE_REPOSITORY
ProjectName: proj1
LogicalCloud: default
CompositeApp: collection-composite-app
DeploymentIntentGroup: collection-deployment-intent-group
NET

    cat << NET > emco-cfg.yaml
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
swc:
  host: $HOST_IP
  port: 30488
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
        if [ "${HOST_IP}" == "oops" ] || [ "${KUBE_PATH1}" == "oops" ]; then
            echo -e "ERROR - HOST_IP, KUBE_PATH1 environment variable needs to be set"
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
