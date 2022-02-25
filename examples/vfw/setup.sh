#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

HOST_IP=${HOST_IP:-"oops"}
KUBE_PATH=${KUBE_PATH:-"oops"}
PRIVILEGED=${2:-"admin"}

# tar files
# test_folder=../tests/
# demo_folder=../demo/
# deployment_folder=../../deployments/
firewall_folder=../helm_charts/composite-cnf-firewall
http_client_folder=../helm_charts/http-client
http_server_folder=../helm_charts/http-server
collectd_folder=../helm_charts/collectd
prometheus_operator_folder=../helm_charts/prometheus-operator
operator_latest_folder=../helm_charts/operators-latest
m3db_folder=../helm_charts/m3db
monitor_folder=../helm_charts/monitor

function create {
    mkdir -p output
    tar -czf output/collectd.tar.gz -C $collectd_folder/helm .
    tar -czf output/collectd_profile.tar.gz -C $collectd_folder/profile .
    tar -czf output/prometheus-operator.tar.gz -C $prometheus_operator_folder/helm .
    tar -czf output/prometheus-operator_profile.tar.gz -C $prometheus_operator_folder/profile .
    tar -czf output/operator.tar.gz -C $operator_latest_folder/helm .
    tar -czf output/operator_profile.tar.gz -C $operator_latest_folder/profile .
    tar -czf output/m3db.tar.gz -C $m3db_folder/helm .
    tar -czf output/m3db_profile.tar.gz -C $m3db_folder/profile .
    tar -czf output/http-client.tar.gz -C $http_client_folder/helm http-client
    tar -czf output/http-server.tar.gz -C $http_server_folder/helm http-server
    tar -czf output/http-server-profile.tar.gz -C $http_server_folder/profile/network_policy_overrides/http-server-profile .
    tar -czf output/http-client-profile.tar.gz -C $http_client_folder/profile/network_policy_overrides/http-client-profile .
    tar -czf output/firewall.tar.gz -C $firewall_folder/helm firewall
    tar -czf output/packetgen.tar.gz -C $firewall_folder/helm packetgen
    tar -czf output/sink.tar.gz -C $firewall_folder/helm sink
    tar -czf output/profile.tar.gz -C $firewall_folder/profile manifest.yaml override_values.yaml
    tar -czf output/monitor.tar.gz -C $monitor_folder/helm monitor

        cat << NET > values.yaml
    ProjectName: proj1
    ClusterProvider: provider1
    Cluster1: cluster1
    ClusterLabel: edge-cluster
    ClusterLabelNetworkPolicy: networkpolicy-supported
    Cluster1Ref: cluster1-ref
    AdminCloud: default
    PrivilegedCloud: privileged-cloud
    PrimaryNamespace: ns1
    ClusterQuota: quota1
    StandardPermission: standard-permission
    PrivilegedPermission: privileged-permission
    CompositeApp: prometheus-collectd-composite-app
    App1: prometheus-operator
    App2: collectd
    App3: operator
    App4: http-client
    App5: http-server
    AppMonitor: monitor
    KubeConfig: $KUBE_PATH
    HelmApp1: output/prometheus-operator.tar.gz
    HelmApp2: output/collectd.tar.gz
    HelmApp3: output/operator.tar.gz
    HelmApp4: output/http-client.tar.gz
    HelmApp5: output/http-server.tar.gz
    HelmAppMonitor: output/monitor.tar.gz
    HelmAppFirewall: output/firewall.tar.gz
    HelmAppPacketgen: output/packetgen.tar.gz
    HelmAppSink: output/sink.tar.gz
    ProfileFw: output/profile.tar.gz
    ProfileApp1: output/prometheus-operator_profile.tar.gz
    ProfileApp2: output/collectd_profile.tar.gz
    ProfileApp3: output/operator_profile.tar.gz
    ProfileApp4: output/http-client-profile.tar.gz
    ProfileApp5: output/http-server-profile.tar.gz
    CompositeProfile: collection-composite-profile
    GenericPlacementIntent: collection-placement-intent
    DeploymentIntent: collection-deployment-intent-group
    RsyncPort: 30441
    CompositeAppGac: gac-composite-app
    GacIntent: collectd-gac-intent
    CompositeAppDtc: dtc-composite-app
    DtcIntent: collectd-dtc-intent
    CompositeAppMonitor: monitor-composite-app
    ConfigmapFile: info.json
    GacPort: 30493
    OvnPort: 30473
    DtcPort: 30483
    NpsPort: 30485
    HostIP: $HOST_IP

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
NET

# head of prerequisites.yaml
cat << NET > prerequisites.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

---
#create project
version: emco/v2
resourceContext:
  anchor: projects
metadata :
   name: {{.ProjectName}}
---
#creating controller entries
version: emco/v2
resourceContext:
  anchor: controllers
metadata :
   name: rsync
spec:
  host:  {{.HostIP}}
  port: {{.RsyncPort}}

---
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
#creating dtc controller entries
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
#creating cluster provider
version: emco/v2
resourceContext:
  anchor: cluster-providers
metadata :
   name: {{.ClusterProvider}}

---
#creating cluster
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters
metadata :
   name: {{.Cluster1}}
file:
  {{.KubeConfig}}

---
#Add label cluster
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters/{{.Cluster1}}/labels
clusterLabel: {{.ClusterLabel}}

NET

if [ "$PRIVILEGED" = "privileged" ]; then
# rest of prerequisites.yaml for a privileged cloud
cat << NET >> prerequisites.yaml
---
#create privileged logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds
metadata:
  name: {{.PrivilegedCloud}}
spec:
  namespace: {{.PrimaryNamespace}}
  user:
    userName: user-1
    type: certificate

---
#create cluster quotas
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.PrivilegedCloud}}/cluster-quotas
metadata:
    name: {{.ClusterQuota}}
spec:
    limits.cpu: '400'
    limits.memory: 1000Gi
    requests.cpu: '300'
    requests.memory: 900Gi
    requests.storage: 500Gi
    requests.ephemeral-storage: '500'
    limits.ephemeral-storage: '500'
    persistentvolumeclaims: '500'
    pods: '500'
    configmaps: '1000'
    replicationcontrollers: '500'
    resourcequotas: '500'
    services: '500'
    services.loadbalancers: '500'
    services.nodeports: '500'
    secrets: '500'
    count/replicationcontrollers: '500'
    count/deployments.apps: '500'
    count/replicasets.apps: '500'
    count/statefulsets.apps: '500'
    count/jobs.batch: '500'
    count/cronjobs.batch: '500'
    count/deployments.extensions: '500'

---
#add primary user permission
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.PrivilegedCloud}}/user-permissions
metadata:
    name: {{.StandardPermission}}
spec:
    namespace: {{.PrimaryNamespace}}
    apiGroups:
    - ""
    - "apps"
    - "k8splugin.io"
    resources:
    - secrets
    - pods
    - configmaps
    - services
    - deployments
    - resourcebundlestates
    verbs:
    - get
    - watch
    - list
    - create
    - delete

---
#add privileged cluster-wide user permission
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.PrivilegedCloud}}/user-permissions
metadata:
    name: {{.PrivilegedPermission}}
spec:
    namespace: ""
    apiGroups:
    - "*"
    resources:
    - "*"
    verbs:
    - "*"

---
#add cluster reference to logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.PrivilegedCloud}}/cluster-references
metadata:
  name: {{.Cluster1Ref}}
spec:
  clusterProvider: {{.ClusterProvider}}
  cluster: {{.Cluster1}}
  loadbalancerIp: "0.0.0.0"

NET

# instantiation.yaml specifically to instantiate a privileged logical cloud
cat << NET >> instantiation.yaml
---
#instantiate logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.PrivilegedCloud}}/instantiate

NET

else
# rest of prerequisites.yaml for an admin cloud
cat << NET >> prerequisites.yaml
---
#create admin logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds
metadata:
  name: {{.AdminCloud}}
spec:
  level: "0"

---
#add cluster reference to logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.AdminCloud}}/cluster-references
metadata:
  name: {{.Cluster1Ref}}
spec:
  clusterProvider: {{.ClusterProvider}}
  cluster: {{.Cluster1}}
  loadbalancerIp: "0.0.0.0"

---
#instantiate logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.AdminCloud}}/instantiate

NET

fi

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
