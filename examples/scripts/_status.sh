#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020-2022 Intel Corporation

# Function to obtain logical cloud status and wait for it to get instantiated
function get_logical_cloud_apply_status {
	emco_config_file=$1
	project=$2
	logical_cloud_name=$3
    echo "Logical Cloud creation in progress... Please Wait"
    for try in {0..300}; do
        sleep 1
        lc_status="$(emcoctl --config $emco_config_file get projects/$project/logical-clouds/$logical_cloud_name/status |  sed -z 's/.*Response://'| jq -r .deployedStatus)"
        case $lc_status in
            "InstantiateFailed")
                echo "Instantiation of logical cloud failed."
                return 2
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
                return 2
                ;;
        esac
    done

    if (($try == 300)) ; then
        echo "Timeout of 300s exceeded!"
    fi

}

# Function to obtain deployment status and wait for it to get instantiated
function get_deployment_intent_group_apply_status {
	emco_config_file=$1
	project=$2
	composite_app=$3
	deployment_intent_group_name=$4
    echo "Deployment in progress... Please Wait"
    for try in {0..300}; do
        sleep 1
        deployment_status="$(emcoctl --config $emco_config_file get projects/$project/composite-apps/$composite_app/v1/deployment-intent-groups/$deployment_intent_group_name/status?status=deployed |  sed -z 's/.*Response://' | jq -r .deployedStatus)"
        case $deployment_status in
            "InstantiateFailed")
                echo "Instantiation of deployment intent group failed."
                return 2
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
                return 2
                ;;
        esac
    done

    if (($try == 300)) ; then
        echo "Timeout of 300s exceeded!"
    fi
}

# Function to obtain logical cloud status and wait for it to get terminated
function get_logical_cloud_delete_status {
	emco_config_file=$1
	project=$2
	logical_cloud_name=$3
    echo "Logical Cloud deletion in progress... Please Wait"
    for try in {0..300}; do
        sleep 1
        lc_status="$(emcoctl --config $emco_config_file get projects/$project/logical-clouds/$logical_cloud_name/status |  sed -z 's/.*Response://'| jq -r .deployedStatus)"
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
	emco_config_file=$1
	project=$2
	composite_app=$3
	deployment_intent_group_name=$4
    echo "Deployment deletion in progress... Please Wait"
    for try in {0..300}; do
        sleep 1
        deployment_status="$(emcoctl --config $emco_config_file get projects/$project/composite-apps/$composite_app/v1/deployment-intent-groups/$deployment_intent_group_name/status?status=deployed |  sed -z 's/.*Response://' | jq -r .deployedStatus)"
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