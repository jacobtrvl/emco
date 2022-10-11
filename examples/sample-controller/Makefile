# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation
export GO111MODULE=on
export SERVICE_BASE_IMAGE_NAME=alpine
export SERVICE_BASE_IMAGE_VERSION=3.12

ifndef COMMITID
export COMMITID=$(shell git show -s --format=%h)
endif

ifndef TAG
export TAG=latest
endif

all: compile build-containers

clean:
	$(MAKE) -C ./src/sample clean

pre-compile: clean
	mkdir -p bin/sample
	ARGS=""; CJ="src/sample/config.json"; JS="src/sample/json-schemas"; RS="src/sample/ref-schemas"; \
	    [ -f $$CJ ] && ARGS="$$ARGS $$CJ"; \
	    [ -d $$JS ] && ARGS="$$ARGS $$JS"; \
	    [ -d $$RS ] && ARGS="$$ARGS $$RS"; \
	    [ -z "$$ARGS" ] || cp -r $$ARGS bin/sample

compile: pre-compile
	$(MAKE) -C ./src/sample all

build-containers:
	@export ARGS="--build-arg MAINDOCKERREPO=${MAINDOCKERREPO} --build-arg SERVICE_BASE_IMAGE_NAME=${SERVICE_BASE_IMAGE_NAME} --build-arg SERVICE_BASE_IMAGE_VERSION=${SERVICE_BASE_IMAGE_VERSION} --build-arg EMCO_SHA=${COMMITID} --build-arg EMCO_VERSION=${TAG}"; \
	docker build $$ARGS --rm -t emco-sample:${TAG} -f ./build/docker/Dockerfile.sample ./bin/sample

run-containers:
	docker run --name emco-sample --network="emconet" --rm -p 9024-9025:9024-9025/tcp emco-sample:${TAG}

test:
	$(MAKE) -C ./src/sample test
