#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020-2022 Intel Corporation

source ../scripts/_status.sh

test_case_name=$1
action=$2


function create {
    ./setup.sh create
}

function cleanup {
    ./setup.sh cleanup
}

function get_variables {
    project=$(cat values.yaml | grep ProjectName: | sed -z 's/.*ProjectName: //')
    logical_cloud_name=$(cat values.yaml | grep LogicalCloud: | sed -z 's/.*LogicalCloud: //')

    case $test_case_name in
        "test-prometheus-collectd")
            composite_app="prometheus-collectd-composite-app"
            deployment_intent_group_name="collection-deployment-intent-group"
            ;;
        "test-dtc")
            composite_app="dtc-composite-app"
            deployment_intent_group_name="collection-deployment-intent-group"
            ;;
        "test-vfw")
            composite_app="compositevfw"
            deployment_intent_group_name="vfw_deployment_intent_group"
            ;;
        "monitor")
            composite_app="monitor-composite-app"
            deployment_intent_group_name="collection-deployment-intent-group"
            ;;
        *)
            echo "Invalid testcase file"
            exit
            ;;
    esac
}

function apply_prerequisites {
    emcoctl --config emco-cfg.yaml apply -f prerequisites.yaml -v values.yaml -s
}

function apply_logical_cloud {
    emcoctl --config emco-cfg.yaml apply -f instantiate-lc.yaml -v values.yaml -s
}

function apply_deployment {
    emcoctl --config emco-cfg.yaml apply -f "${test_case_name}-deployment.yaml" -v values.yaml -s
}

function apply_instantiate_testcase {
    emcoctl --config emco-cfg.yaml apply -f "${test_case_name}-instantiate.yaml" -v values.yaml -s
}

function delete_prerequisites {
    emcoctl --config emco-cfg.yaml delete -f prerequisites.yaml -v values.yaml -s
}

function delete_logical_cloud {
    emcoctl --config emco-cfg.yaml delete -f instantiate-lc.yaml -v values.yaml -s
}

function delete_deployment {
    emcoctl --config emco-cfg.yaml delete -f "${test_case_name}-deployment.yaml" -v values.yaml -s
}

function delete_instantiate_testcase {
    emcoctl --config emco-cfg.yaml delete -f "${test_case_name}-instantiate.yaml" -v values.yaml -s
}

function check_exit_code {
    if (($? == 2)) ; then
        echo "Exiting script!"
        exit
    fi
}

function usage {

    echo "Usage: $0 <test case file> apply|delete"
    echo "Example: $0 test-prometheus-collectd.yaml apply"

}

function apply {
    cleanup
    create
    get_variables
    apply_prerequisites
    apply_logical_cloud
    get_logical_cloud_apply_status emco-cfg.yaml $project $logical_cloud_name # wait till Instantiated status is obtained
    check_exit_code
    apply_deployment
    apply_instantiate_testcase
    get_deployment_intent_group_apply_status emco-cfg.yaml $project $composite_app $deployment_intent_group_name # wait till Instantiated status is obtained
    check_exit_code
}

function delete {
    get_variables
    delete_instantiate_testcase
    get_deployment_intent_group_delete_status emco-cfg.yaml $project $composite_app $deployment_intent_group_name # wait till Terminated status is obtained
    delete_deployment
    delete_logical_cloud
    get_logical_cloud_delete_status emco-cfg.yaml $project $logical_cloud_name # wait till Terminated status is obtained
    delete_prerequisites
    cleanup
}

if [ "$#" -lt 2 ] ; then
    usage
    exit
fi

case "$2" in
    "apply" ) apply ;;
    "delete" ) delete ;;
    *) usage ;;
esac
