#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

#########################################################

# build the "build base image" that will be used as the base for Helm
echo "Building ${BUILD_BASE_IMAGE_NAME}:${BUILD_BASE_IMAGE_VERSION} image (with Helm v${HELM_VERSION})"
docker build --build-arg HTTP_PROXY=${HTTP_PROXY} --build-arg HTTPS_PROXY=${HTTPS_PROXY} --build-arg HELM_VERSION=${HELM_VERSION} -t ${BUILD_BASE_IMAGE_NAME} -f build/docker/Dockerfile.git-build .
${DIR}/deploy-docker.sh ${BUILD_BASE_IMAGE_NAME} ${BUILD_BASE_IMAGE_VERSION}
${DIR}/deploy-docker.sh ${BUILD_BASE_IMAGE_NAME} latest

#########################################################

# build the "git service image" that will be used by rsync and monitor
echo "Building ${GIT_SERVICE_IMAGE_NAME}:${GIT_SERVICE_IMAGE_VERSION}"
docker build --build-arg HTTP_PROXY=${HTTP_PROXY} --build-arg HTTPS_PROXY=${HTTPS_PROXY} --build-arg BASEDOCKERREPO=${BASEDOCKERREPO} --build-arg EMCODOCKERREPO=${EMCODOCKERREPO} --build-arg BUILD_BASE_IMAGE_NAME=${BUILD_BASE_IMAGE_NAME} --build-arg BUILD_BASE_IMAGE_VERSION=${BUILD_BASE_IMAGE_VERSION} --build-arg HELM_VERSION=${HELM_VERSION} -t ${GIT_SERVICE_IMAGE_NAME} -f build/docker/Dockerfile.emco-service .
${DIR}/deploy-docker.sh ${GIT_SERVICE_IMAGE_NAME} ${GIT_SERVICE_IMAGE_VERSION}
${DIR}/deploy-docker.sh ${GIT_SERVICE_IMAGE_NAME} latest