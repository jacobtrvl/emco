#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

HOST_IP=${HOST_IP:-"oops"}
KUBE_PATH1=${KUBE_PATH1:-"oops"}
KUBE_PATH2=${KUBE_PATH2:-"oops"}
KUBE_PATH3=${KUBE_PATH3:-"oops"}
KUBE_PATH_ISSUING=${KUBE_PATH_ISSUING:-"oops"}

function create {
# head of values.yaml
cat << NET > values.yaml
ProjectName: proj1
ClusterProvider: provider1

# Issuing Cluster
IssuingCluster: issuer1
IssuingClusterConfig: $KUBE_PATH_ISSUING

# Clsuters
Cluster1: cluster1
KubeConfig1: $KUBE_PATH1
Cluster2: cluster2
KubeConfig2: $KUBE_PATH2
Cluster3: cluster3
KubeConfig3: $KUBE_PATH3

# Cluster Label
GroupLabel1: group1
GroupLabel23: group23

# Cert intent names
CertIntent0: cert0
CertIntent1: cert1
CertIntent2: cert2

# Cluster Issuer identifier
ClusterIssuer0: new-istio-system
ClusterIssuer1: foo
ClusterIssuer2: foobar
Kind: ClusterIssuer
Group: cert-manager.io

# CSR Info
KeySize: 4096
CommonNamePrefix: foo

# Cluster Group
ClusterGroup0a: group0a
ClusterGroup0b: group0b
ClusterGroup1a: group1a
ClusterGroup1b: group1b
ClusterGroup1c: group1c
ClusterGroup2a: group2a
ClusterGroup2b: group2b

# Logical Clouds
FooLogicalCloud: foo
BarLogicalCloud: bar
FooBarLogicalCloud: foobar
FooCloud: foo-ns
BarCloud: bar-ns
FooBarCloud: foobar-ns
NET

# head of emco-cfg.yaml
cat << NET > emco-cfg.yaml
cacert:
  host: $HOST_IP
  port: 30436
orchestrator:
  host: $HOST_IP
  port: 30415
  statusPort: 30416
clm:
  host: $HOST_IP
  port: 30461
ncm:
  host: $HOST_IP
  port: 30481
  statusPort: 30482
ovnaction:
  host: $HOST_IP
  port: 30451
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
    echo "       Set the following environment variables:"
    echo "       HOST_IP - IP address of EMCO cluster host"
    echo "       KUBE_PATH_ISSUING - Path to kubeconfig for the Issuing Cluster"
    echo "       KUBE_PATH1 - Path to kubeconfig for Edge Cluster 1"
    echo "       KUBE_PATH2 - Path to kubeconfig for Edge Cluster 2"
    echo "       KUBE_PATH3 - Path to kubeconfig for Edge Cluster 3"
}

function cleanup {
    rm -f values.yaml
    rm -f emco-cfg.yaml
}

if [ "$#" -lt 1 ] ; then
    usage
    exit
fi

case "$1" in
    "create" )
        if [ "${HOST_IP}" == "oops" ] || [ "${KUBE_PATH_ISSUING}" == "oops" ] || [ "${KUBE_PATH1}" == "oops" ] || [ "${KUBE_PATH2}" == "oops" ] || [ "${KUBE_PATH3}" == "oops" ]; then
            echo -e "ERROR - HOST_IP, KUBE_PATH_ISSUING, KUBE_PATH1, KUBE_PATH2, KUBE_PATH3 environment variables need to be set"
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
