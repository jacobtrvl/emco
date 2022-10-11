# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation
export GO111MODULE=on
export EMCOBUILDROOT=$(shell pwd)
export CONFIG := $(wildcard config/*.txt)

# build the containers and deploy with helm charts here

all: clean build-docker-workflowclient

compile-workflowclient:
	cd ../../../src/tools/workflowclient && make all

build-docker-workflowclient: compile-workflowclient
	@echo "Building Workflowclient container image."
	docker build --build-arg EMCODOCKERREPO=${EMCODOCKERREPO} --build-arg MAINDOCKERREPO=${MAINDOCKERREPO} --build-arg SERVICE_BASE_IMAGE_NAME=workflow-client --build-arg SERVICE_BASE_IMAGE_VERSION=latest --rm -f docker/Dockerfile.workflowclient -t workflow-client ../../../bin/workflowclient
	docker tag workflow-client:latest ${MAINDOCKERREPO}workflow-client:latest
	docker push ${MAINDOCKERREPO}workflow-client:latest

clean-workflowclient:
	cd ../../../src/tools/workflowclient && make clean

package:
	helm package helm -d ..

clean: clean-workflowclient