# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation
export GO111MODULE=on
export EMCOBUILDROOT=$(shell pwd)
export CONFIG := $(wildcard config/*.txt)

all: build-docker-worker 

build-docker-worker: compile-worker
	@echo "Building worker container"
	docker build --rm -f build/docker/Dockerfile.worker -t example-worker ./bin/worker

compile-worker:
	@echo "Compiling worker with app"
	@mkdir -p bin/worker
	@cd src/worker; go build -o ../../bin/worker/worker main.go

clean-worker:
	rm -rf bin

clean: clean-worker

test:
	@echo "No tests yet"

tidy:
	@echo "No dependencies to clean"

