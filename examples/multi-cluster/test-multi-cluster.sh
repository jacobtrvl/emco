#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020-2022 Intel Corporation

source ../scripts/_status.sh

test_case_name=$1
action=$2

function create {
    ./setup.sh "$@" create
}

function cleanup {
    ./setup.sh cleanup
}

function get_variables {
    project=$(cat values.yaml | grep ProjectName: | sed -z 's/.*ProjectName: //')
    logical_cloud_name=$(cat values.yaml | grep AdminCloud: | sed -z 's/.*AdminCloud: //')
    composite_app=$(cat values.yaml | grep CompositeApp: | sed -z 's/.*CompositeApp: //')
    deployment_intent_group_name=$(cat values.yaml | grep DeploymentIntent: | sed -z 's/.*DeploymentIntent: //')
}

function apply_prerequisites {
    emcoctl --config emco-cfg.yaml apply -f 00-controllers.yaml -v values.yaml -s
    emcoctl --config emco-cfg.yaml apply -f 01-prerequisites.yaml -v values.yaml -s
}

function apply_logical_cloud {
    emcoctl --config emco-cfg.yaml apply -f 02-instantiate-lc.yaml -v values.yaml -s
}

function apply_deployment {
    emcoctl --config emco-cfg.yaml apply -f 03-deployment-intent.yaml -v values.yaml -s
}

function apply_instantiate_testcase {
    emcoctl --config emco-cfg.yaml apply -f 04-deployment-instantiate.yaml -v values.yaml -s
}

function delete_prerequisites {
    emcoctl --config emco-cfg.yaml delete -f 01-prerequisites.yaml -v values.yaml -s
    emcoctl --config emco-cfg.yaml delete -f 00-controllers.yaml -v values.yaml -s
}

function delete_logical_cloud {
    emcoctl --config emco-cfg.yaml delete -f 02-instantiate-lc.yaml -v values.yaml -s
}

function delete_deployment {
    emcoctl --config emco-cfg.yaml delete -f 03-deployment-intent.yaml -v values.yaml -s
}

function delete_instantiate_testcase {
    emcoctl --config emco-cfg.yaml delete -f 04-deployment-instantiate.yaml -v values.yaml -s
}

function check_exit_code {
    if (($? == 2)) ; then
        echo "Exiting script!"
        exit
    fi
}

function usage {

    echo "Usage: $0 apply|delete <args>"
    echo "Example: $0 apply -a collectd:cluster1:cluster2 -b operator:cluster1"

}

function apply {
    cleanup
    create "$@"
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

if [ "$#" -lt 1 ] ; then
    usage
    exit
fi

action=$1
shift
case "$action" in
    "apply" ) apply "$@" ;;
    "delete" ) delete ;;
    *) usage ;;
esac
