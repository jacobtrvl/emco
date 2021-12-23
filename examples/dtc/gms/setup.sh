#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2021 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

HOST_IP=${HOST_IP:-"oops"}
KUBE_PATH1=${KUBE_PATH1:-"oops"}
KUBE_PATH2=${KUBE_PATH2:-"oops"}
# tar files
function create {
    # make the GMS helm charts and profiles
    mkdir -p output
    tar -czf output/adservice.tgz -C ../../helm_charts/gms/helm adservice
    tar -czf output/cartservice.tgz -C ../../helm_charts/gms/helm cartservice
    tar -czf output/checkoutservice.tgz -C ../../helm_charts/gms/helm checkoutservice
    tar -czf output/currencyservice.tgz -C ../../helm_charts/gms/helm currencyservice
    tar -czf output/emailservice.tgz -C ../../helm_charts/gms/helm emailservice
    tar -czf output/frontend.tgz -C ../../helm_charts/gms/helm frontend
    tar -czf output/loadgenerator.tgz -C ../../helm_charts/gms/helm loadgenerator
    tar -czf output/paymentservice.tgz -C ../../helm_charts/gms/helm paymentservice
    tar -czf output/recommendationservice.tgz -C ../../helm_charts/gms/helm recommendationservice
    tar -czf output/redis.tgz -C ../../helm_charts/gms/helm redis
    tar -czf output/shippingservice.tgz -C ../../helm_charts/gms/helm shippingservice
    tar -czf output/productcatalogservice.tgz -C ../../helm_charts/gms/helm productcatalogservice
    tar -czf output/gms-profile.tar.gz -C ../../helm_charts/gms/profile .


        cat << NET > values.yaml
    ClusterProvider: gmsprovider1
    Cluster1: gmscluster1
    Cluster2: gmscluster2
    KubeConfig1: $KUBE_PATH1
    KubeConfig2: $KUBE_PATH2
    ProjectName: gmsproj1
    LogicalCloud1RefName: gms-lc-cl-1
    LogicalCloud2RefName: gms-lc-cl-2
    Cluster1Label: edge-cluster
    Cluster2Label: edge-cluster1
    Cluster1IstioIngressGatewayKvName: gmsistioingresskvpairs1
    Cluster2IstioIngressGatewayKvName: gmsistioingresskvpairs2
    AdminCloud: default
    CompositeApp: gms-collection-composite-app
    CompositeAppVersion: v1
    Applist: 
      - adservice
      - cartservice
      - checkoutservice
      - currencyservice
      - emailservice
      - loadgenerator
      - paymentservice
      - recommendationservice
      - redis
      - shippingservice
      - frontend
      - productcatalogservice
    AppsInCluster1: 
      - adservice
      - checkoutservice
      - currencyservice
      - emailservice
      - loadgenerator
      - paymentservice
      - recommendationservice
      - frontend
    AppsInCluster2: 
      - productcatalogservice
      - cartservice
      - redis
      - shippingservice
    CompositeProfile: gms-collection-composite-profile
    DeploymentIntentGroup: gms-collection-deployment-intent-group
    DeploymentIntent: gms-collection-deployment-intent
    GenericPlacementIntent: gms-collection-placement-intent
    DtcIntent: gmstestdtc
    DtcProductcatalogServerIntent: productcatalogserver
    DtcCartserviceServerIntent: cartserviceserver
    DtcRediscartServerIntent: rediscartserver
    DtcShippingServerIntent: shippingserver
    Intent: gmsintent
    RsyncPort: 30441
    DtcPort: 30483
    ItsPort: 30487
    HostIP: $HOST_IP
NET
cat << NET > emco-cfg.yaml
  orchestrator:
    host: $HOST_IP
    port: 30415
  clm:
    host: $HOST_IP
    port: 30461
  ncm:
    host: $HOST_IP
    port: 30431
  ovnaction:
    host: $HOST_IP
    port: 30471
  dcm:
    host: $HOST_IP
    port: 30477
  gac:
    host: $HOST_IP
    port: 30491
  dtc:
   host: $HOST_IP
   port: 30481
  rsync:
   host: $HOST_IP
   port: 30441
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
        if [ "${HOST_IP}" == "oops" ] || [ "${KUBE_PATH1}" == "oops" ] || [ "${KUBE_PATH2}" == "oops" ]; then
            echo -e "ERROR - HOST_IP, KUBE_PATH1 & KUBE_PATH2 environment variable needs to be set"
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
