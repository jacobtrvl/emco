#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020-2022 Intel Corporation

action=$1


function create {
    ./setup.sh create
}

function cleanup {
    ./setup.sh cleanup
}

function get_variables {
    project=$(cat values.yaml | grep ProjectName: | sed -z 's/.*ProjectName: //')
    logical_cloud_name=$(cat values.yaml | grep AdminCloud: | sed -z 's/.*AdminCloud: //')
}

function apply_prerequisites {
    emcoctl --config emco-cfg.yaml apply -f 00-prerequisites.yaml -v values.yaml -s
}

function apply_logical_cloud {
    emcoctl --config emco-cfg.yaml apply -f 01-logical-cloud.yaml -v values.yaml -s
}

function apply_deployment {
    emcoctl --config emco-cfg.yaml apply -f 02-deployment-intent.yaml -v values.yaml -s
}

function apply_instantiate_testcase {
    emcoctl --config emco-cfg.yaml apply -f 03-instantiation.yaml -v values.yaml -s
}

function delete_prerequisites {
    emcoctl --config emco-cfg.yaml delete -f 00-prerequisites.yaml -v values.yaml -s
}

function delete_logical_cloud {
    emcoctl --config emco-cfg.yaml delete -f 01-logical-cloud.yaml -v values.yaml -s
}

function delete_deployment {
    emcoctl --config emco-cfg.yaml delete -f 02-deployment-intent.yaml -v values.yaml -s
}

function delete_instantiate_testcase {
    emcoctl --config emco-cfg.yaml delete -f 03-instantiation.yaml -v values.yaml -s
}

# Function to obtain logical cloud status and wait for it to get instantiated
function get_logical_cloud_apply_status {
    echo "Logical Cloud creation in progress... Please Wait"
    for try in {0..300}; do
        sleep 1
        lc_status="$(emcoctl --config emco-cfg.yaml get projects/$project/logical-clouds/$logical_cloud_name/status |  sed -z 's/.*Response://'| jq -r .deployedStatus)"
        case $lc_status in
            "InstantiateFailed")
                echo "Instantiation of logical cloud failed."
                exit
                ;;

            "Instantiating")
                echo -n "."
                continue
                ;;

            "Instantiated")
                echo "Logical cloud creation succeeded!"
                break
                ;;

            *)
                echo "Invalid apply status State."
                exit
                ;;
        esac
    done

    if (($try == 300)) ; then
        echo "Timeout of 300s exceeded!"
    fi

}

# Function to obtain deployment status and wait for it to get instantiated
function get_deployment_intent_group_apply_status {
    echo "Deployment in progress... Please Wait"
    for try in {0..300}; do
        sleep 1
        deployment_status="$(emcoctl --config emco-cfg.yaml get projects/$project/composite-apps/test-composite-app/v1/deployment-intent-groups/test-deployment-intent/status?status=deployed |  sed -z 's/.*Response://' | jq -r .deployedStatus)"

        case $deployment_status in
            "InstantiateFailed")
                echo "Instantiation of deployment intent group failed."
                exit
                ;;

            "Instantiating")
                echo -n "."
                continue
                ;;

            "Instantiated")
                echo "Deployment succeeded!"
                break
                ;;

            *)
                echo "Invalid apply status State."
                exit
                ;;
        esac
    done

    if (($try == 300)) ; then
        echo "Timeout of 300s exceeded!"
    fi
}

# Function to obtain logical cloud status and wait for it to get terminated
function get_logical_cloud_delete_status {
    echo "Logical Cloud deletion in progress... Please Wait"
    for try in {0..300}; do
        sleep 1
        lc_status="$(emcoctl --config emco-cfg.yaml get projects/$project/logical-clouds/$logical_cloud_name/status |  sed -z 's/.*Response://' | jq -r .deployedStatus)"

        case $lc_status in
            "TerminateFailed")
                echo "Termination of logical cloud failed."
                break
                ;;

            "Terminating")
                echo -n "."
                continue
                ;;

            "Terminated")
                echo "Logical Cloud termination succeeded!"
                break
                ;;

            *)
                echo "Invalid delete status State."
                break
                ;;
        esac
    done

    if (($try == 300)) ; then
        echo "Timeout of 300s exceeded!"
    fi

}

# Function to obtain deployment status and wait for it to get terminated
function get_deployment_intent_group_delete_status {
    echo "Deployment deletion in progress... Please Wait"
    for try in {0..300}; do
        sleep 1
        deployment_status="$(emcoctl --config emco-cfg.yaml get projects/$project/composite-apps/test-composite-app/v1/deployment-intent-groups/test-deployment-intent/status?status=deployed |  sed -z 's/.*Response://' | jq -r .deployedStatus)"

        case $deployment_status in
            "TerminateFailed")
                echo "Termination of deployment intent group failed."
                break
                ;;

            "Terminating")
                echo -n "."
                continue
                ;;

            "Terminated")
                echo "Deployment termination succeeded!"
                break
                ;;

            *)
                echo "Invalid delete status State."
                break
                ;;
        esac
    done

    if (($try == 300)) ; then
        echo "Timeout of 300s exceeded!"
    fi
}

function usage {

    echo "Usage: $0 apply|delete"
    echo "Example: $0 apply"

}

function apply {
    cleanup
    create
    get_variables
    apply_prerequisites
    apply_logical_cloud
    get_logical_cloud_apply_status # wait till Instantiated status is obtained
    apply_deployment
    apply_instantiate_testcase
    get_deployment_intent_group_apply_status # wait till Instantiated status is obtained
}

function delete {
    get_variables
    delete_instantiate_testcase
    get_deployment_intent_group_delete_status # wait till Terminated status is obtained
    delete_deployment
    delete_logical_cloud
    get_logical_cloud_delete_status # wait till Terminated status is obtained
    delete_prerequisites
    cleanup
}

if [ "$#" -lt 1 ] ; then
    usage
    exit
fi

case "$1" in
    "apply" ) apply ;;
    "delete" ) delete ;;
    *) usage ;;
esac
