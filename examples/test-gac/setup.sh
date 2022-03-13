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
RsyncPort: 30441
GacPort: 30493
DtcPort: 30483
NpsPort: 30485
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
  port: 30491
dtc:
  host: $HOST_IP
  port: 30481
NET

# head of prerequisites.yaml
cat << NET > prerequisites.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

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
#Add label kvpair
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters/{{.Cluster}}/kv-pairs
metadata :
  name: {{.IstioIngressGatewayKvName}}
spec:
  kv:
    - istioingressgatewayaddress: "192.168.121.26" 
    - istioingressgatewayport: "32001"
    - istioingressgatewayinternalport: "15443" 

---
#Add label kvpair
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters/{{.Cluster}}/kv-pairs
metadata :
  name: {{.DatabaseAuthorizationKvName}}
spec:
  kv:
    - user: aGVsbG8=
    - password: MWYyZDFlMmU2N2Rm

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

---
# register gac controller
version: emco/v2
resourceContext:
  anchor: controllers
metadata :
  name: gac
spec:
  host: {{.HostIP}}
  port: {{.GacPort}}
  type: "action"
  priority: 1

---
# create gac compositeApp
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps
metadata :
  name: {{.CompositeAppGac}}
  description: test
spec:
  compositeAppVersion: v1

---
# add app to the compositeApp
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeAppGac}}/v1/apps
metadata :
  name: {{.App1}}
  description: "description for app"
file:
  {{.HelmApp1}}

---
# create gac compositeProfile
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeAppGac}}/v1/composite-profiles
metadata :
  name: {{.CompositeProfile}}
  description: test

---
# add profiles to the compositeProfile
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeAppGac}}/v1/composite-profiles/{{.CompositeProfile}}/profiles
metadata :
  name: profile1
  description: test
spec:
  app: {{.App1}}
file:
  {{.ProfileApp1}}

---
# create deployment intent group
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeAppGac}}/v1/deployment-intent-groups
metadata :
  name: {{.DeploymentIntent}}
  description: "description"
spec:
  compositeProfile: {{.CompositeProfile}}
  version: r6
  logicalCloud: {{.AdminCloud}}
  overrideValues: []

---
# create intent in deployment intent group
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeAppGac}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/intents
metadata :
  name: collection-deployment-intent
  description: "description"
spec:
  intent:
    genericPlacementIntent: {{.GenericPlacementIntent}}
    gac: {{.GacIntent}}

---
# create the generic placement intent 
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeAppGac}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/generic-placement-intents
metadata :
  name: {{.GenericPlacementIntent}}
  description: "description for app"
spec:
  logicalCloud: {{.AdminCloud}}

---
# add the app placement intent to the generic placement intent
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeAppGac}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/generic-placement-intents/{{.GenericPlacementIntent}}/app-intents
metadata:
  name: placement-intent
  description: description of placement_intent
spec:
  app: {{.App1}}
  intent:
    allOf:
    - clusterProvider: {{.ClusterProvider}}
      clusterLabel: {{.ClusterLabel}}

---
# add the gac intent
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeAppGac}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/generic-k8s-intents
metadata:
  name: {{.GacIntent}}
NET

# head of instantiate.yaml
cat << NET > instantiate.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# approve
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeAppGac}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/approve

---
# instantiate
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeAppGac}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/instantiate
NET

# head of update.yaml
cat << NET > update.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# update
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeAppGac}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/update
NET

# head of rollback.yaml
cat << NET > rollback.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# rollback
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeAppGac}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/rollback
metadata:
  description: "rollback to revision 1"
spec:
  revision: "1"
NET

}

function usage {
    echo "Usage: $0  create|cleanup"
}

function cleanup {
    rm -f *.tar.gz
    rm -f values.yaml
    rm -f emco-cfg.yaml
    rm -f prerequisites.yaml
    rm -f instantiate.yaml
    rm -f update.yaml
    rm -f rollback.yaml
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
