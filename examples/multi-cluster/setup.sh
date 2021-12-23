#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail


source _common.sh

test_folder=../../tests/
demo_folder=../../demo/
deployment_folder=../../../deployments/

HOST_IP=${HOST_IP:-"oops"}
KUBE_PATH1=${KUBE_PATH1:-"oops"}
KUBE_PATH2=${KUBE_PATH2:-"oops"}
KUBE_PATH3=${KUBE_PATH3:-"oops"}
KUBE_PATH4=${KUBE_PATH4:-"oops"}
KUBE_PATH5=${KUBE_PATH5:-"oops"}
KUBE_PATH6=${KUBE_PATH6:-"oops"}
KUBE_PATH7=${KUBE_PATH7:-"oops"}
KUBE_PATH8=${KUBE_PATH8:-"oops"}
KUBE_PATH9=${KUBE_PATH9:-"oops"}
KUBE_PATH10=${KUBE_PATH10:-"oops"}
OUTPUT_DIR=output

function create_common_values {
    local output_dir=$1
    local host_ip=$2

    create_apps $output_dir
    create_config_file $host_ip

        cat << NET > values.yaml
    PackagesPath: $output_dir
    ProjectName: proj-multi-1
    ClusterProvider: provider-multi
    ClusterLabel: edge-cluster
    AdminCloud: default
    CompositeApp: test-composite-app
    CompositeProfile: test-composite-profile
    GenericPlacementIntent: test-placement-intent
    DeploymentIntent: test-deployment-intent
    Intent: intent
    RsyncPort: 30441
    GacPort: 30493
    OvnPort: 30473
    DtcPort: 30483
    NpsPort: 30485
    HostIP: $host_ip
NET
}
# Format for clusters added
#    Clusters:
#    - KubeConfig: $KUBE_PATH1
#      Name: cluster1
#    - KubeConfig: $KUBE_PATH2
#     Name: cluster2
function add_clusters {

    i="0"

while [ $i -lt 9 ]
do
    j=$[$i+1]
    name="cluster"
    name+=$j
    kubestr="KUBE_PATH"
    kubestr+=$j
    a="oops"
    eval a=\$$kubestr
    if [[ "$a" == 'oops' ]]; then
        i=$[$i+1]
        continue
    fi
    namestr=$name index=$i ./yq eval '.Clusters[strenv(index)].Name = strenv(namestr)' -i values.yaml
    path=$a index=$i ./yq eval '.Clusters[strenv(index)].KubeConfig = strenv(path)' -i values.yaml
    i=$[$i+1]
done
}

function eval_array {
    arr=$1
    for ele in "${array[@]}"
        do
            echo "myele", $ele
    done
}

#Syntax:
#Applist:
#   - Name: $app_1
#      Cluster:
#      - cluster1
#    - Name: $app_2
#      Cluster:
#      - cluster1
#      - cluster2
function add_apps {
    local app_1=$1
    local app_2=$2
    local app_3=$3

    i="0"
    while [ $i -lt 3 ]
    do
        if [[ "$i" == 0 ]]; then
            arr=(${app_1//:/ })
        fi
        if [[ "$i" == 1 ]]; then
            arr=(${app_2//:/ })
        fi
        if [[ "$i" == 2 ]]; then
            arr=(${app_3//:/ })
        fi
        a="oops"
        a="${arr[0]}"
        if [[ "$a" == 'oops' ]]; then
            i=$[$i+1]
            continue
        fi
        namestr=$a index=$i ./yq eval '.Applist[strenv(index)].Name = strenv(namestr)' -i values.yaml
        x="0"
        for ele in "${arr[@]}"
        do
            if [[ "$x" == '0' ]]; then
                x=$[$x+1]
                continue
            fi
            y=$[$x-1]
            path=$ele index=$i inner=$y ./yq eval '.Applist[strenv(index)].Cluster[strenv(inner)] = strenv(path)' -i values.yaml
            x=$[$x+1]
        done

        i=$[$i+1]
    done
}

function create_values_yaml_clusters_apps {
    local output_dir=$1
    local host_ip=$2
    local app_1=$3
    local app_2=$4
    local app_3=$5

    create_common_values $output_dir $host_ip
    add_clusters
    add_apps $app_1 $app_2 $app_3
}

function usage {
    echo "Usage: $0 -a app1:cluster1:cluster2 -b m3db:cluster1 -c app3:cluster3 create|cleanup"
}

function cleanup {
    rm -f yq
    rm -f *.tar.gz
    rm -f values.yaml
    rm -f emco-cfg.yaml
    rm -rf $OUTPUT_DIR
}

# Install yq for parsing yaml files. It installs it locally (current folder) if it is not
# already present. The rest of this script uses this local version (so as to not conflict
# with other versions potentially installed on the system already.
function install_yq_locally {
    if [ ! -x ./yq ]; then
        echo 'Installing yq locally'
        VERSION=v4.12.0
        BINARY=yq_linux_amd64
        wget https://github.com/mikefarah/yq/releases/download/${VERSION}/${BINARY} -O yq && chmod +x yq
fi
}

app1_name="oops"
app2_name="oops"
app3_name="oops"

while getopts ":a:b:c:" flag
do
    case "${flag}" in
        a) app1_name=${OPTARG};;
        b) app2_name=${OPTARG};;
        c) app3_name=${OPTARG};;
    esac
done
shift $((OPTIND-1))

input="hello"

install_yq_locally
case "$1" in
    "create" )
        if [ "${HOST_IP}" == "oops" ] ; then
            echo -e "ERROR - Environment varaible HOST_IP must be set"
            exit
        fi
        if [ "${KUBE_PATH1}" == "oops"  ] ; then
            echo -e "ERROR - Atleast one cluster config must be provided KUBE_PATH1"
            exit
        fi
        if [ "${app1_name}" == "oops" ] ; then
            echo -e "Atleast one 1 app must be provided (ex collectd, prometheus-operator, operator, http-client, http-server) must be provided on commandline -a -b"
        else
            create_values_yaml_clusters_apps $OUTPUT_DIR $HOST_IP $app1_name $app2_name $app3_name
            echo "Done create!!!"
        fi
        ;;
    "cleanup" )
        cleanup
    ;;
    *)
        usage ;;
esac
