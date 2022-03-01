#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

#########################################################

# build the "base build image" that will be used as the base for all containerized builds & deployments

echo "Building build-base container (version ${BUILD_BASE_IMAGE_VERSION} with go${GO_VERSION})"
docker build --build-arg GO_VERSION=${GO_VERSION} --build-arg HTTP_PROXY=${HTTP_PROXY} --build-arg HTTPS_PROXY=${HTTPS_PROXY} -t emco-service-build-base -f build/docker/Dockerfile.build-base .
${DIR}/deploy-docker.sh emco-service-build-base ${BUILD_BASE_IMAGE_VERSION}
${DIR}/deploy-docker.sh emco-service-build-base latest

#########################################################
