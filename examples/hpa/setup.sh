#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2021 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

HOST_IP=${HOST_IP:-"oops"}
KUBE_PATH1=${KUBE_PATH1:-"oops"}
KUBE_PATH2=${KUBE_PATH2:-"oops"}
KUBE_PATH3=${KUBE_PATH3:-"oops"}
KUBE_PATH4=${KUBE_PATH4:-"oops"}

# tar files
firewall_folder=../helm_charts/composite-cnf-firewall
http_client_folder=../helm_charts/http-client
http_server_folder=../helm_charts/http-server

function create {
    mkdir -p output
    tar -czf output/firewall.tar.gz -C $firewall_folder/helm firewall
    tar -czf output/packetgen.tar.gz -C $firewall_folder/helm packetgen
    tar -czf output/sink.tar.gz -C $firewall_folder/helm sink
    tar -czf output/profile.tar.gz -C $firewall_folder/profile manifest.yaml override_values.yaml
    tar -czf output/http-client.tgz -C $http_client_folder/helm http-client
    tar -czf output/http-server.tgz -C $http_server_folder/helm http-server
    tar -czf output/http-server-profile.tar.gz -C $http_server_folder/profile/network_policy_overrides/http-server-profile .
    tar -czf output/http-client-profile.tar.gz -C $http_client_folder/profile/network_policy_overrides/http-client-profile .

        cat << NET > values.yaml
    ProjectName: proj1
    
    # cluster info
    ClusterProvider: provider1
    Cluster1: cluster1
    ClusterLabel1: edge-cluster1
    KubeConfig1: $KUBE_PATH1
    Cluster2: cluster2
    ClusterLabel2: edge-cluster2
    KubeConfig2: $KUBE_PATH2
    Cluster3: cluster3
    ClusterLabel3: edge-cluster3
    KubeConfig3: $KUBE_PATH3
    Cluster4: cluster4
    ClusterLabel4: edge-cluster4
    KubeConfig4: $KUBE_PATH4
    Cluster1Ref: lc-cl-1
    Cluster2Ref: lc-cl-2
    AdminCloud: default

    # provider network info for virtual firewall app
    EmcoPrivateNet: emco-private-net
    EmcoPrivateSubnet: 10.10.20.0/24
    EmcoPrivateSubnetName: subnet1
    EmcoPrivateGateway: 10.10.20.1/24
    UnprotectedPrivateNet: unprotected-private-net
    UnprotectedPrivateSubnet: 192.168.10.0/24
    UnprotectedPrivateSubnetName: subnet1
    UnprotectedPrivateGateway: 192.168.10.1/24
    ProtectedPrivateNet: protected-private-net
    ProtectedPrivateSubnet: 192.168.20.0/24
    ProtectedPrivateSubnetName: subnet1
    ProtectedPrivateGateway: 192.168.20.1/24

    # vfw application info
    VfwCompositeApp: vfw-composite-app
    VfwCompositeProfile: vfw-composite-profile
    VfwPacketGen: packetgen
    ResourcePacketgen: r1-packetgen
    VfwPacketgenHelmApp: output/packetgen.tar.gz
    VfwPacketgenProfile: output/profile.tar.gz
    VfwFirewall: firewall
    ResourceFirewall: r1-firewall
    VfwFirewallHelmApp: output/firewall.tar.gz
    VfwFirewallProfile: output/profile.tar.gz
    VfwSink: sink
    ResourceSink: r1-sink
    VfwSinkHelmApp: output/sink.tar.gz
    VfwSinkProfile: output/profile.tar.gz

    # vfw app network interface info
    PacketgenUnprotectedIF: eth1
    PacketgenUnprotectedDefaultGW: false
    PacketgenUnprotectedIPAddr: 192.168.10.2
    PacketgenEmcoIF: eth2
    PacketgenEmcoDefaultGW: false
    PacketgenEmcoIPAddr: 10.10.20.2
    FirewallEmcoIF: eth3
    FirewallEmcoDefaultGW: false
    FirewallEmcoIPAddr: 10.10.20.3
    FirewallUnprotectedIF: eth1
    FirewallUnprotectedDefaultGW: false
    FirewallUnprotectedIPAddr: 192.168.10.3
    FirewallProtectedIF: eth2
    FirewallProtectedDefaultGW: false
    FirewallProtectedIPAddr: 192.168.20.2
    SinkProtectedIF: eth1
    SinkProtectedDefaultGW: false
    SinkProtectedIPAddr: 192.168.20.3
    SinkEmcoIF: eth2
    SinkEmcoDefaultGW: false
    SinkEmcoIPAddr: 10.10.20.4

    # controller info
    HpaPlacementControllerName: hpa-placement-controller-1
    HpaActionControllerName: hpa-action-controller-1

    # placement intent info
    VfwDeploymentIntentGroup: collection-deployment-intents
    VfwGenericPlacementIntent: collection-placement-intent
    VfwHpaActionIntent: hpa-action-intent
    VfwHpaPlacementIntent: hpa-placement-intent
    VfwNetworkIntent: ovnaction-network-intent
    PacketgenPlacementIntent: packetgen-placement-intent
    FirewallPlacementIntent: firewall-placement-intent
    SinkPlacementIntent: sink-placement-intent
    HpaPlacementIntentPacketgen: hpa-placement-intent-packetgen
    HpaPlacementIntentFirewall: hpa-placement-intent-firewall
    HpaHugepages: hpa-placement-consumer-hugepages

    # host and ports
    HostIP: $HOST_IP
    RsyncPort: 9031
    HpaPlacementPort: 9099
    OvnActionPort: 9053
    HpaActionPort:  9042

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
  port: 30493
dtc:
  host: $HOST_IP
  port: 30483
hpaplacement:
  host: $HOST_IP
  port: 30451
hpaaction:
  host: $HOST_IP
  port: 30443
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
        # if [ "${HOST_IP}" == "oops" ] || [ "${KUBE_PATH1}" == "oops" ] || [ "${KUBE_PATH2}" == "oops" ] ; then
        #     echo -e "ERROR - HOST_IP & KUBE_PATH environment variable needs to be set"
        # else
        #     create
        # fi
        create
        ;;
    "cleanup" )
        cleanup
    ;;
    *)
        usage ;;
esac
