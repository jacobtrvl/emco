#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

WFC_FILE="workflowclient-1.0.0.tgz"

install() {
    echo "Uninstalling Temporal Workflow Client"

    kubectl create ns temporal
    helm install  workflowclient ${WFC_FILE} -n temporal
}

uninstall() {
    echo "Installing Temporal Workflow Client"
    
    kubectl delete ns temporal
}

if [ "$1" = "install" ]; then
	install
elif [ "$1" = "uninstall" ]; then
	uninstall
else
	echo "Not a valid command: "$2
	exit 2
fi
exit 0