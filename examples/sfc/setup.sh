#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2021 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

HOST_IP=${HOST_IP:-"oops"}
KUBE_PATH=${KUBE_PATH:-"oops"}
# tar files
function create {
    # make the SFC helm charts and profiles
    mkdir -p output
    tar -czf output/ngfw.tar.gz -C chainCA/helm ngfw
    tar -czf output/nets.tar.gz -C chainCA/helm nets
    tar -czf output/profile.tar.gz -C chainCA manifest.yaml override_values.yaml

    if [ "$1" == "npn"  ] ; then
        cp -r chainCA/helm/slb-npn /tmp/slb
        tar -czf output/slb.tar.gz -C /tmp slb
	rm -r /tmp/slb
        cp -r chainCA/helm/sdewan-npn /tmp/sdewan
        tar -czf output/sdewan.tar.gz -C /tmp sdewan
	rm -r /tmp/sdewan
    else
        tar -czf output/sdewan.tar.gz -C chainCA/helm sdewan
        tar -czf output/slb.tar.gz -C chainCA/helm slb
    fi

    tar -czf output/left-nginx.tar.gz -C clientCA/helm left-nginx
    tar -czf output/right-nginx.tar.gz -C clientCA/helm right-nginx
    tar -czf output/clientprofile.tar.gz -C clientCA manifest.yaml override_values.yaml

        cat << NET > values.yaml
    ClusterProvider: provider1
    Cluster1: cluster1
    KubeConfig: $KUBE_PATH
    AdminCloud: default
    LeftCloud: left
    RightCloud: right
    LeftNamespace: sfc-head
    RightNamespace: sfc-tail
    LeftLabel: head
    RightLabel: tail
    LeftCloud2: left2
    RightCloud2: right2
    LeftNamespace2: sfc-head-two
    RightNamespace2: sfc-tail-two
    LeftLabel2: head2
    RightLabel2: tail2

    # virtual network names
    SfcDynNet1: dynamic-net1
    NgfwDynNet1If: net2
    SlbDynNet1If: net4
    SfcDynNet2: dynamic-net2
    NgfwDynNet2If: net3
    SdewanDynNet2If: net3
    SfcVirNet1: virtual-net1
    SlbVirNet1If: net2
    SfcVirNet2: virtual-net2
    SdewanVirNet2If: net2

    # chain links
    SfcLinkVnet1: linkVnet1
    SfcLinkVnet2: linkVnet2
    SfcLinkDnet1: linkDnet1
    SfcLinkDnet2: linkDnet2
    SfcLinkSlb: linkSlb
    SfcLinkNgfw: linkNgfw
    SfcLinkSdewan: linkSdewan

    # provider network names
    SfcLeftPNet: left-pnetwork
    SlbLeftPNetIf: net3
    SfcRightPNet: right-pnetwork
    SdewanRightPNetIf: net4

    ProjectName: proj1

    SfcCompositeApp: sfc-ca
    SfcCompositeProfile: sfc-profile
    FnNgfw: ngfw
    HelmAppNgfw: output/ngfw.tar.gz
    FnSdewan: sdewan
    HelmAppSdewan: output/sdewan.tar.gz
    FnSlb: slb

    NetsCompositeApp: nets-ca
    NetsCompositeProfile: nets-profile
    AppNets: nets
    HelmNets: output/nets.tar.gz
    ProfileNets: output/profile.tar.gz

    HelmAppSlb: output/slb.tar.gz
    ProfileNgfw: output/profile.tar.gz
    ProfileSdewan: output/profile.tar.gz
    ProfileSlb: output/profile.tar.gz

    # Deployment intent group for the Nets
    NetsDeploymentIntentGroup: nets-deployment-intent-group
    NetsGenericPlacementIntent: nets-generic-placement
    NetsPlacementIntent: nets-placement

    # Deployment intent group for the SFC chain
    SfcDeploymentIntentGroup: sfc-deployment-intent-group
    SfcGenericPlacementIntent: sfc-generic-placement
    NgfwPlacementIntent: sfc-ngfw-placement
    SdewanPlacementIntent: sfc-sdewan-placement
    SlbPlacementIntent: sfc-slb-placement

    # ovnaction intents for the ovnaction controller
    # most of these (the virtual networks) will be
    # deleted once nodus support automatic creation
    # of the virtual networks
    OvnactionSfcFns: sfc-fn-intent
    OvnactionSfcFnNgfw: sfc-fn-ngfw
    OvnactionSfcFnNgfwIf1: sfc-fn-ngfw-if1
    OvnactionSfcFnNgfwIf2: sfc-fn-ngfw-if2
    OvnactionSfcFnSdewan: sfc-fn-sdewan
    OvnactionSfcFnSdewanIf1: sfc-fn-sdewan-if1
    OvnactionSfcFnSdewanIf2: sfc-fn-sdewan-if2
    OvnactionSfcFnSdewanIf3: sfc-fn-sdewan-if3
    OvnactionSfcFnSlb: sfc-fn-slb-intent
    OvnactionSfcFnSlbIf1: sfc-fn-slb-if1
    OvnactionSfcFnSlbIf2: sfc-fn-slb-if2
    OvnactionSfcFnSlbIf3: sfc-fn-slb-if3

    # SFC intents for the sfc controller
    SfcIntent: sfc-intent
    SfcLeftClientSelectorIntent: left-client-selector
    SfcRightClientSelectorIntent: right-client-selector
    SfcLeftProviderNetworkIntent: right-provider-network
    SfcRightProviderNetworkIntent: left-provider-network

    # SFC Client CA definitions
    SfcLeftClientCA: sfc-left-client-ca
    SfcRightClientCA: sfc-right-client-ca
    SfcClientCompositeProfile: sfc-client-profile
    LeftNginx: left-nginx
    RightNginx: right-nginx
    HelmAppLeftNginx: output/left-nginx.tar.gz
    HelmAppRightNginx: output/right-nginx.tar.gz
    ProfileNginx: output/profile.tar.gz

    # Deployment intent group for the SFC chain
    SfcLeftDig: sfc-left-dig
    SfcLeftDig2: sfc-left-dig-2
    SfcRightDig: sfc-right-dig
    SfcRightDig2: sfc-right-dig-2
    SfcClientGenericPlacementIntent: sfc-client-generic-placement
    LeftNginxPlacementIntent: sfc-left-nginx-placement
    RightNginxPlacementIntent: sfc-right-nginx-placement

    # SFC client intents
    SfcLeftClientIntent: sfc-left-client-intent
    SfcRightClientIntent: sfc-right-client-intent

    # controller port numbers
    RsyncPort: 9031
    OvnPort: 9053
    SfcPort: 9056
    SfcClientPort: 9058

NET
cat << NET > emco-cfg.yaml
  orchestrator:
    host: $HOST_IP
    port: 30415
    statusPort: 30416
  clm:
    host: $HOST_IP
    port: 30461
  ncm:
    host: $HOST_IP
    port: 30431
    statusPort: 30482
  ovnaction:
    host: $HOST_IP
    port: 30471
  dcm:
    host: $HOST_IP
    port: 30477
    statusPort: 30478
  gac:
    host: $HOST_IP
    port: 30491
  dtc:
   host: $HOST_IP
   port: 30481
  sfc:
   host: $HOST_IP
   port: 30455
  sfcclient:
   host: $HOST_IP
   port: 30457
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
            create $2
        fi
        ;;
    "cleanup" )
        cleanup
    ;;
    *)
        usage ;;
esac
