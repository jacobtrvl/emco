#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

REGISTRY=${EMCODOCKERREPO}
EMCOBUILDROOT=$(cd ../../.. && pwd)
BIN_PATH=${EMCOBUILDROOT}/bin
TAG=${TAG}

create_workflow_client_chart() {
    echo "Creating Helm Chart."

    # copy all of the needed files to the bin dir
    mkdir -p ${BIN_PATH}/temporal/workflowclient
    cp -rf ${EMCOBUILDROOT}/deployments/temporal/workflowclient/helm Makefile ${BIN_PATH}/temporal/workflowclient
    
    # copy install script, and change permissions to execute.
    cp ../temporal-install.sh ${BIN_PATH}/temporal
    chmod +x ${BIN_PATH}/temporal/temporal-install.sh

    # package helm into .tar.gz
    make -C ${BIN_PATH}/temporal/workflowclient package

    # clean up all the source files
    rm -rf ${BIN_PATH}/temporal/workflowclient
}

create_workflow_client_chart