#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

# tar files
firewall_folder=../helm_charts/composite-cnf-firewall
http_client_folder=../helm_charts/http-client
http_server_folder=../helm_charts/http-server
collectd_folder=../helm_charts/collectd
prometheus_operator_folder=../helm_charts/prometheus-operator
operator_latest_folder=../helm_charts/operators-latest
m3db_folder=../helm_charts/m3db
monitor_folder=../helm_charts/monitor

function create_apps {
    local output_dir=$1

    mkdir -p $output_dir
    tar -czf $output_dir/collectd.tar.gz -C $collectd_folder/helm .
    tar -czf $output_dir/collectd_profile.tar.gz -C $collectd_folder/profile .
    tar -czf $output_dir/prometheus-operator.tar.gz -C $prometheus_operator_folder/helm .
    tar -czf $output_dir/prometheus-operator_profile.tar.gz -C $prometheus_operator_folder/profile .
    tar -czf $output_dir/operator.tar.gz -C $operator_latest_folder/helm .
    tar -czf $output_dir/operator_profile.tar.gz -C $operator_latest_folder/profile .
    tar -czf $output_dir/m3db.tar.gz -C $m3db_folder/helm .
    tar -czf $output_dir/m3db_profile.tar.gz -C $m3db_folder/profile .
    tar -czf $output_dir/http-client.tar.gz -C $http_client_folder/helm http-client
    tar -czf $output_dir/http-server.tar.gz -C $http_server_folder/helm http-server
    tar -czf $output_dir/http-server_profile.tar.gz -C $http_server_folder/profile/network_policy_overrides/http-server-profile .
    tar -czf $output_dir/http-client_profile.tar.gz -C $http_client_folder/profile/network_policy_overrides/http-client-profile .
    tar -czf $output_dir/firewall.tar.gz -C $firewall_folder/helm firewall
    tar -czf $output_dir/packetgen.tar.gz -C $firewall_folder/helm packetgen
    tar -czf $output_dir/sink.tar.gz -C $firewall_folder/helm sink
    tar -czf $output_dir/profile.tar.gz -C $firewall_folder/profile manifest.yaml override_values.yaml
    tar -czf $output_dir/monitor.tar.gz -C $monitor_folder/helm monitor

}

function create_values_yaml_one_cluster {
    local output_dir=$1
    local host_ip=$2
    local kube_path=$3

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
    KubeConfig: $kube_path
    HelmApp1: $output_dir/prometheus-operator.tar.gz
    HelmApp2: $output_dir/collectd.tar.gz
    HelmApp3: $output_dir/operator.tar.gz
    HelmApp4: $output_dir/http-client.tar.gz
    HelmApp5: $output_dir/http-server.tar.gz
    HelmAppMonitor: $output_dir/monitor.tar.gz
    HelmAppFirewall: $output_dir/firewall.tar.gz
    HelmAppPacketgen: $output_dir/packetgen.tar.gz
    HelmAppSink: $output_dir/sink.tar.gz
    ProfileFw: $output_dir/profile.tar.gz
    ProfileApp1: $output_dir/prometheus-operator_profile.tar.gz
    ProfileApp2: $output_dir/collectd_profile.tar.gz
    ProfileApp3: $output_dir/operator_profile.tar.gz
    ProfileApp4: $output_dir/http-client_profile.tar.gz
    ProfileApp5: $output_dir/http-server_profile.tar.gz
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
    HostIP: $host_ip

NET
}

function create_config_file {
    local host_ip=$1
cat << NET > emco-cfg.yaml
  orchestrator:
    host: $host_ip
    port: 30415
    statusPort: 30416
  clm:
    host: $host_ip
    port: 30461
  ncm:
    host: $host_ip
    port: 30431
    statusPort: 30482
  ovnaction:
    host: $host_ip
    port: 30471
  dcm:
    host: $host_ip
    port: 30477
    statusPort: 30478
  gac:
    host: $host_ip
    port: 30491
  dtc:
   host: $host_ip
   port: 30481
NET
}

