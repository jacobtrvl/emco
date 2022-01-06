#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

#########################################################

# build the "base build image" that will be used as the base for all containerized builds & deployments

# if you update Dockerfile.build-base, please bump up the version here so as to not overwrite older base images
BUILD_BASE_VERSION=1.3
# TODO: get the value above from config.txt instead

echo "Building build-base container"
docker build --build-arg GO_VERSION=${GO_VERSION} --build-arg HTTP_PROXY=${HTTP_PROXY} --build-arg HTTPS_PROXY=${HTTPS_PROXY} -t emco-service-build-base -f build/docker/Dockerfile.build-base .
${DIR}/deploy-docker.sh emco-service-build-base ${BUILD_BASE_VERSION}
${DIR}/deploy-docker.sh emco-service-build-base latest

#########################################################
