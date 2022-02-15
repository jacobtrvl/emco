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
  
  # head of emco-cfg.yaml
  cat << NET > emco-cfg.yaml
    orchestrator:
      host: $HOST_IP
      port: 30415
    dtc:
      host: $HOST_IP
      port: 30481
    clm:
      host: $HOST_IP
      port: 30461
    dcm:
      host: $HOST_IP
      port: 30477
    gac:
      host: $HOST_IP
      port: 30491

NET

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
    GacIntent: collectd-gac-intent
    RsyncPort: 30441
    GacPort: 30493
    DtcPort: 30483
    NpsPort: 30485
    HostIP: $HOST_IP

NET

# head of prerequisites.yaml
  cat << NET > prerequisites.yaml
    # SPDX-License-Identifier: Apache-2.0
    # Copyright (c) 2020 Intel Corporation

    ---
    # register rsync controller
    version: emco/v2
    resourceContext:
      anchor: controllers
    metadata :
      name: rsync
    spec:
      host:  {{.HostIP}}
      port: {{.RsyncPort}}

    ---
    # register dtc controller
    version: emco/v2
    resourceContext:
      anchor: controllers
    metadata :
      name: dtc
    spec:
      host: {{.HostIP}}
      port: {{.DtcPort}}
      type: "action"
      priority: 1

    ---
    # register dtc sub controller nps
    version: emco/v2
    resourceContext:
      anchor: dtc-controllers
    metadata :
      name: nps
    spec:
      host:  {{.HostIP}}
      port: {{.NpsPort}}
      type: "action"
      priority: 1

    ---
    # create project
    version: emco/v2
    resourceContext:
      anchor: projects
    metadata :
      name: {{.ProjectName}}

    ---
    # create cluster provider
    version: emco/v2
    resourceContext:
      anchor: cluster-providers
    metadata :
      name: {{.ClusterProvider}}

    ---
    # create cluster
    version: emco/v2
    resourceContext:
      anchor: cluster-providers/{{.ClusterProvider}}/clusters
    metadata :
      name: {{.Cluster}}
    file:
      {{.KubeConfig}}

    ---
    # add cluster label
    version: emco/v2
    resourceContext:
      anchor: cluster-providers/{{.ClusterProvider}}/clusters/{{.Cluster}}/labels
    clusterLabel: {{.ClusterLabel}}

    ---
    # create admin logical cloud
    version: emco/v2
    resourceContext:
      anchor: projects/{{.ProjectName}}/logical-clouds
    metadata:
      name: {{.AdminCloud}}
    spec:
      level: "0"

    ---
    # add cluster reference to logical cloud
    version: emco/v2
    resourceContext:
      anchor: projects/{{.ProjectName}}/logical-clouds/{{.AdminCloud}}/cluster-references
    metadata:
      name: {{.ClusterRef}}
    spec:
      clusterProvider: {{.ClusterProvider}}
      cluster: {{.Cluster}}
      loadbalancerIp: "0.0.0.0"

    ---
    # instantiate logical cloud
    version: emco/v2
    resourceContext:
      anchor: projects/{{.ProjectName}}/logical-clouds/{{.AdminCloud}}/instantiate

NET

}

function usage {
    echo "Usage: $0  create|cleanup"
}

function cleanup {
    rm -f *.tar.gz
    rm -f values.yaml
    rm -f emco-cfg.yaml
    rm -f instantiation.yaml
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
