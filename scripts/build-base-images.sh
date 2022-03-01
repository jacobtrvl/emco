#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

#########################################################

# build the "build base image" that will be used as the base for Helm
echo "Building ${BUILD_BASE_IMAGE_NAME}:${BUILD_BASE_IMAGE_VERSION} image (with Helm v3.5.2) from ${SERVICE_BASE_IMAGE_NAME}:${SERVICE_BASE_IMAGE_VERSION}"
docker build --build-arg HTTP_PROXY=${HTTP_PROXY} --build-arg HTTPS_PROXY=${HTTPS_PROXY} --build-arg MAINDOCKERREPO=${MAINDOCKERREPO} --build-arg SERVICE_BASE_IMAGE_NAME=${SERVICE_BASE_IMAGE_NAME} --build-arg SERVICE_BASE_IMAGE_VERSION=${SERVICE_BASE_IMAGE_VERSION} -t ${BUILD_BASE_IMAGE_NAME} -f build/docker/Dockerfile.build-base .
${DIR}/deploy-docker.sh ${BUILD_BASE_IMAGE_NAME} ${BUILD_BASE_IMAGE_VERSION}
${DIR}/deploy-docker.sh ${BUILD_BASE_IMAGE_NAME} latest

#########################################################
