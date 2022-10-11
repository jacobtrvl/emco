#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

# for the Logical Cloud level (LOGICAL_CLOUD_LEVEL), supported values are: admin, privileged and standard.
echo "Reading config file"
CONFIG_DNS=$(cat ./config | grep DNS | cut -d'=' -f2)
CONFIG_HOST_IP=$(cat ./config | grep HOST_IP | cut -d'=' -f2)
CONFIG_KUBE_PATH=$(cat ./config | grep KUBE_PATH | cut -d'=' -f2)
CONFIG_LOGICAL_CLOUD_LEVEL=$(cat ./config | grep LOGICAL_CLOUD_LEVEL | cut -d'=' -f2)
CONFIG_IMAGE_REPOSITORY=$(cat ./config | grep IMAGE_REPOSITORY | cut -d'=' -f2)
CLM_SERVICE_PORT=$(cat ./config | grep CLM_SERVICE_PORT | cut -d'=' -f2)
DCM_SERVICE_PORT=$(cat ./config | grep DCM_SERVICE_PORT | cut -d'=' -f2)
DCM_STATUS_PORT=$(cat ./config | grep DCM_STATUS_PORT | cut -d'=' -f2)
DTC_SERVICE_PORT=$(cat ./config | grep DTC_SERVICE_PORT | cut -d'=' -f2)
DTC_CONTROL_PORT=$(cat ./config | grep DTC_CONTROL_PORT | cut -d'=' -f2)
GAC_CONTROL_PORT=$(cat ./config | grep GAC_CONTROL_PORT | cut -d'=' -f2)
GAC_SERVICE_PORT=$(cat ./config | grep GAC_SERVICE_PORT | cut -d'=' -f2)
NCM_SERVICE_PORT=$(cat ./config | grep NCM_SERVICE_PORT | cut -d'=' -f2)
NCM_STATUS_PORT=$(cat ./config | grep NCM_STATUS_PORT | cut -d'=' -f2)
NPS_CONTROL_PORT=$(cat ./config | grep NPS_CONTROL_PORT | cut -d'=' -f2)
OVN_CONTROL_PORT=$(cat ./config | grep OVN_CONTROL_PORT | cut -d'=' -f2)
OVN_SERVICE_PORT=$(cat ./config | grep OVN_SERVICE_PORT | cut -d'=' -f2)
ORCH_SERVICE_PORT=$(cat ./config | grep ORCH_SERVICE_PORT | cut -d'=' -f2)
ORCH_STATUS_PORT=$(cat ./config | grep ORCH_STATUS_PORT | cut -d'=' -f2)
RSYNC_CONTROL_PORT=$(cat ./config | grep RSYNC_CONTROL_PORT | cut -d'=' -f2)

# priority: environment first, config file second, default value third:
DNS=${DNS:-$CONFIG_DNS}
DNS=${DNS:-"oops"}
HOST_IP=${HOST_IP:-$CONFIG_HOST_IP}
HOST_IP=${HOST_IP:-"oops"}
KUBE_PATH=${KUBE_PATH:-$CONFIG_KUBE_PATH}
KUBE_PATH=${KUBE_PATH:-"oops"}
LOGICAL_CLOUD_LEVEL=${LOGICAL_CLOUD_LEVEL:-$CONFIG_LOGICAL_CLOUD_LEVEL}
LOGICAL_CLOUD_LEVEL=${LOGICAL_CLOUD_LEVEL:-"admin"}
IMAGE_REPOSITORY=${IMAGE_REPOSITORY:-$CONFIG_IMAGE_REPOSITORY}
IMAGE_REPOSITORY=${IMAGE_REPOSITORY:-"myrepository.example.com"}

firewall_folder=../helm_charts/composite-cnf-firewall
http_client_folder=../helm_charts/http-client
http_server_folder=../helm_charts/http-server
collectd_folder=../helm_charts/collectd
prometheus_operator_folder=../helm_charts/prometheus-operator
operator_latest_folder=../helm_charts/operators-latest
m3db_folder=../helm_charts/m3db
monitor_folder=../../deployments/helm
profiles_folder=../profiles
kube_prometheus_stack_folder=../helm_charts/kube-prometheus-stack

function create {
    echo "Generating tarballs from Helm charts"
    mkdir -p output
    tar -czf output/collectd.tar.gz -C $collectd_folder/helm .
    tar -czf output/collectd_profile.tar.gz -C $collectd_folder/profile .
    tar -czf output/prometheus-operator.tar.gz -C $prometheus_operator_folder/helm .
    tar -czf output/prometheus-operator_profile.tar.gz -C $prometheus_operator_folder/profile .
    tar -czf output/kube-prometheus-stack.tar.gz -C $kube_prometheus_stack_folder/helm .
    tar -czf output/kube-prometheus-stack_profile.tar.gz -C $prometheus_operator_folder/profile .
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
    tar -czf output/monitor.tar.gz -C $monitor_folder monitor
    tar -czf output/monitor_profile.tar.gz -C $profiles_folder/default/profile .


echo "Generating values.yaml"
cp templates/values-novars.yaml values.yaml
cat << NET >> values.yaml
    RsyncPort: $RSYNC_CONTROL_PORT
    GacPort: $GAC_CONTROL_PORT
    OvnPort: $OVN_CONTROL_PORT
    DtcPort: $DTC_CONTROL_PORT
    NpsPort: $NPS_CONTROL_PORT
    KubeConfig: $KUBE_PATH
    HostIP: $HOST_IP
    ImageRepository: $IMAGE_REPOSITORY
NET

echo "Generating emco-cfg.yaml"
cat << NET > emco-cfg.yaml
  orchestrator:
    host: $HOST_IP
    port: $ORCH_SERVICE_PORT
    statusPort: $ORCH_STATUS_PORT
  clm:
    host: $HOST_IP
    port: $CLM_SERVICE_PORT
  ncm:
    host: $HOST_IP
    port: $NCM_SERVICE_PORT
    statusPort: $NCM_STATUS_PORT
  ovnaction:
    host: $HOST_IP
    port: $OVN_SERVICE_PORT
  dcm:
    host: $HOST_IP
    port: $DCM_SERVICE_PORT
    statusPort: $DCM_STATUS_PORT
  gac:
    host: $HOST_IP
    port: $GAC_SERVICE_PORT
  dtc:
    host: $HOST_IP
    port: $DTC_SERVICE_PORT
NET

echo "Generating prerequisites.yaml: common section"
cp templates/prerequisites-common.yaml prerequisites.yaml

if [ "$LOGICAL_CLOUD_LEVEL" = "privileged" ]; then
echo "Generating prerequisites.yaml: Privileged Logical Cloud section"
cat templates/prerequisites-lc-privileged.yaml >> prerequisites.yaml

elif [ "$LOGICAL_CLOUD_LEVEL" = "standard" ]; then
echo "Generating prerequisites.yaml: Standard Logical Cloud section"
cat templates/prerequisites-lc-standard.yaml >> prerequisites.yaml

else
echo "Generating prerequisites.yaml: Admin Logical Cloud section"
cat templates/prerequisites-lc-admin.yaml >> prerequisites.yaml

fi

if [ "$DNS" = "true" ]; then
echo "Replacing controller addresses with hostnames, since DNS service reachability (via internal port) is enabled"
perl -p0i -e 's/rsync\nspec:\n  host: {{.HostIP}}/rsync\nspec:\n  host: rsync/' prerequisites.yaml
perl -p0i -e 's/dtc\nspec:\n  host: {{.HostIP}}/dtc\nspec:\n  host: dtc/' prerequisites.yaml
perl -p0i -e 's/nps\nspec:\n  host: {{.HostIP}}/nps\nspec:\n  host: nps/' prerequisites.yaml
perl -p0i -e 's/ovnaction\nspec:\n  host: {{.HostIP}}/ovnaction\nspec:\n  host: ovnaction/' test-vfw-deployment.yaml
fi

}

function usage {
    echo "Usage: $0 create|cleanup"
}

function cleanup {
    echo "Deleting all *.tar.gz"
    rm -f *.tar.gz
    echo "Deleting values.yaml"
    rm -f values.yaml
    echo "Deleting emco-cfg.yaml"
    rm -f emco-cfg.yaml
    echo "Deleting prerequisites.yaml"
    rm -f prerequisites.yaml
    echo "Deleting the output/ folder"
    rm -rf output

    if [ "$DNS" = "true" ]; then
    echo "Reverting test-vfw-deployment.yaml back to original version"
    perl -p0i -e 's/ovnaction\nspec:\n  host: ovnaction/ovnaction\nspec:\n  host: {{.HostIP}}/' test-vfw-deployment.yaml
    fi
}

if [ "$#" -lt 1 ] ; then
    usage
    exit
fi

case "$1" in
    "create" )
        if [ "${HOST_IP}" == "oops" ] || [ "${KUBE_PATH}" == "oops" ] ; then
            echo -e "ERROR - HOST_IP & KUBE_PATH need to be defined as environment variables or in the config file"
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
