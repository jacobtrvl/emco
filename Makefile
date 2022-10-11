# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation
export GO111MODULE=on
export EMCOBUILDROOT=$(shell pwd)
export CONFIG := $(wildcard config/*.txt)

# inject all variables defined in $(CONFIG) file
export GO_VERSION := $(shell cat $(CONFIG) | grep 'GO_VERSION' | cut -d'=' -f2)
export HELM_VERSION := $(shell cat $(CONFIG) | grep 'HELM_VERSION' | cut -d'=' -f2)
export BUILD_BASE_IMAGE_NAME := $(shell cat $(CONFIG) | grep 'BUILD_BASE_IMAGE_NAME' | cut -d'=' -f2)
export BUILD_BASE_IMAGE_VERSION := $(shell cat $(CONFIG) | grep 'BUILD_BASE_IMAGE_VERSION' | cut -d'=' -f2)
export SERVICE_BASE_IMAGE_NAME := $(shell cat $(CONFIG) | grep 'SERVICE_BASE_IMAGE_NAME' | cut -d'=' -f2)
export SERVICE_BASE_IMAGE_VERSION := $(shell cat $(CONFIG) | grep 'SERVICE_BASE_IMAGE_VERSION' | cut -d'=' -f2)
export EMCODOCKERREPO_CONFIG := $(shell cat $(CONFIG) | grep 'EMCODOCKERREPO' | cut -d'=' -f2)
export MAINDOCKERREPO_CONFIG := $(shell cat $(CONFIG) | grep 'MAINDOCKERREPO' | cut -d'=' -f2)
export GIT_SERVICE_IMAGE_NAME := $(shell cat $(CONFIG) | grep 'GIT_SERVICE_IMAGE_NAME' | cut -d'=' -f2)
export GIT_SERVICE_IMAGE_VERSION := $(shell cat $(CONFIG) | grep 'GIT_SERVICE_IMAGE_VERSION' | cut -d'=' -f2)

# docker registry URL defined in environment take precedence over ones defined in the $(CONFIG) file:
ifndef EMCODOCKERREPO
	export EMCODOCKERREPO := ${EMCODOCKERREPO_CONFIG}
endif
ifndef MAINDOCKERREPO
	export MAINDOCKERREPO := ${MAINDOCKERREPO_CONFIG}
endif

ifndef MODS
MODS=clm dcm dtc nps sds its genericactioncontroller monitor ncm orchestrator ovnaction rsync tools/emcoctl sfc sfcclient hpa-plc hpa-ac workflowmgr ca-certs swc tac
endif

noparam:
	@echo "Please pick a make a target, such as: all, build-base, deploy, develop, test, clean-all."

all: check-env compile build-containers

check-env:
	@echo "Check for environment parameters"
ifndef EMCODOCKERREPO
	$(error EMCODOCKERREPO env variable needs to be set)
endif

ifndef COMMITID
export COMMITID=$(shell git show -s --format=%h)
endif

ifndef BRANCH
export BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
endif

ifndef TAG
  ifeq ($(BUILD_CAUSE),RELEASE)
    ifneq ($(EMCOSRV_RELEASE_TAG),)
      # remove "v" from e.g. "vXX.YY"
      export TAG=${EMCOSRV_RELEASE_TAG:v%=%}
    else
      export TAG=$(shell git tag --points-at HEAD | awk 'NR==1 {print $1}')
      ifeq ($(TAG),)
        export TAG=${BRANCH}-daily-`date +"%m%d%y"`
      endif
    endif
  else ifeq ($(BUILD_CAUSE),DEV_TEST)
    export TAG=${USER}-latest
  else ifeq ($(BUILD_CAUSE),TIMERTRIGGER)
    export TAG=${BRANCH}-daily-`date +"%m%d%y"`-${COMMITID}
  else
    export TAG=latest
  endif
endif

ifeq ($(BUILD_BASE),true)
	export BASEDOCKERREPO=${EMCODOCKERREPO}
else
	export BASEDOCKERREPO=registry.gitlab.com/project-emco/core/emco-base/
endif

clean-all:
	@echo "Cleaning artifacts"
	@for m in $(MODS); do \
	    $(MAKE) -C ./src/$$m clean; \
	 done
	@rm -rf bin
	@echo "    Done."

clean:
	@echo "Cleaning artifacts"
	@for m in $(MODS); do \
	    $(MAKE) -C ./src/$$m clean; \
	 done
	@echo "    Done."

pre-compile: clean
	@echo "Setting up pre-requisites"
	@for m in $(MODS); do \
	    mkdir -p bin/$$m;  \
	    ARGS=""; CJ="src/$$m/config.json"; JS="src/$$m/json-schemas"; RS="src/$$m/ref-schemas"; \
	    [ -f $$CJ ] && ARGS="$$ARGS $$CJ"; \
	    [ -d $$JS ] && ARGS="$$ARGS $$JS"; \
	    [ -d $$RS ] && ARGS="$$ARGS $$RS"; \
	    [ -z "$$ARGS" ] || cp -r $$ARGS bin/$$m; \
	 done
	@echo "    Done."

compile: pre-compile
	@echo "Building artifacts"
	@for m in $(MODS); do \
	    $(MAKE) -C ./src/$$m all || exit 1; \
	 done
	@echo "    Done."

deploy-compile: check-env
	@echo "Building microservices within Docker build container"
	docker run --rm --user `id -u`:`id -g` --env MODS="${MODS}" --env GO111MODULE --env XDG_CACHE_HOME=/tmp/.cache --env BRANCH=${BRANCH} --env TAG=${TAG} --env HTTP_PROXY=${HTTP_PROXY} --env HTTPS_PROXY=${HTTPS_PROXY} --env GOPATH=/repo/bin -v `pwd`:/repo ${BASEDOCKERREPO}${BUILD_BASE_IMAGE_NAME}:${BUILD_BASE_IMAGE_VERSION} /bin/sh -c "cd /repo; git config --global --add safe.directory /repo; make compile"
	@echo "    Done."

# Modules that follow naming conventions are done in a loop, rest later
build-containers:
	@echo "Packaging microservices "
	@export ARGS="--build-arg BASEDOCKERREPO=${BASEDOCKERREPO} --build-arg EMCODOCKERREPO=${EMCODOCKERREPO} --build-arg MAINDOCKERREPO=${MAINDOCKERREPO} --build-arg SERVICE_BASE_IMAGE_NAME=${SERVICE_BASE_IMAGE_NAME} --build-arg SERVICE_BASE_IMAGE_VERSION=${SERVICE_BASE_IMAGE_VERSION} --build-arg GIT_SERVICE_IMAGE_NAME=${GIT_SERVICE_IMAGE_NAME} --build-arg GIT_SERVICE_IMAGE_VERSION=${GIT_SERVICE_IMAGE_VERSION} --build-arg EMCO_SHA=${COMMITID} --build-arg EMCO_VERSION=${TAG}"; \
	 for m in $(MODS); do \
	    case $$m in \
	      "tools/emcoctl") continue;; \
	      "ovnaction") d="ovn"; n=$$d;; \
	      "genericactioncontroller") d="gac"; n=$$d;; \
	      "orchestrator") d=$$m; n="orch";; \
	      *) d=$$m; n=$$m;; \
	    esac; \
	    echo "Packaging $$m"; \
	    docker build $$ARGS --rm -t emco-$$n -f ./build/docker/Dockerfile.$$d ./bin/$$m || exit 1; \
	 done
	@echo "    Done."

deploy: check-env deploy-compile build-containers
	@echo "Creating helm charts. Pushing microservices to registry & copying docker-compose files if BUILD_CAUSE set to DEV_TEST"
	@docker run --env USER=${USER} --env BASEDOCKERREPO=${BASEDOCKERREPO} --env EMCODOCKERREPO=${EMCODOCKERREPO} --env MAINDOCKERREPO=${MAINDOCKERREPO} --env BUILD_CAUSE=${BUILD_CAUSE} --env BRANCH=${BRANCH} --env TAG=${TAG} --env EMCOSRV_RELEASE_TAG=${EMCOSRV_RELEASE_TAG} --rm --user `id -u`:`id -g` --env GO111MODULE --env XDG_CACHE_HOME=/tmp/.cache -v `pwd`:/repo ${BASEDOCKERREPO}${BUILD_BASE_IMAGE_NAME}:${BUILD_BASE_IMAGE_VERSION} /bin/sh -c "cd /repo/scripts ; bash deploy_emco.sh"
	@MODS=`echo ${MODS} | sed 's/ovnaction/ovn/;s/genericactioncontroller/gac/;s/orchestrator/orch/;'` ./scripts/push_to_registry.sh
	@echo "    Done."

test:
	@TESTFAILED=""; \
	for m in $(MODS); do \
	  STATUS=0; \
	  echo Running test cases for $$m; \
	  $(MAKE) -C ./src/$$m test || STATUS=$$?; \
	  if [ $$STATUS != 0 ]; then \
	    echo "One or more test case(s) of $$m failed"; \
	    TESTFAILED="$$TESTFAILED$$m,"; \
	  else \
            echo "Test case(s) for $$m executed successfully"; \
	  fi \
	done; \
	if [ ! -z "$$TESTFAILED" -a "$$TESTFAILED" != " " ]; then \
	    echo "One or more test case(s) of $$TESTFAILED failed"; \
	    exit 1; \
	fi

tidy:
	@echo "Cleaning up dependencies"
	@for m in $(MODS); do \
	    cd src/$$m; go mod tidy; cd - > /dev/null; \
	 done
	@echo "    Done."

build-base:
	@echo "Building emco-build-base image and pushing to registry"
	export BASEDOCKERREPO=${EMCODOCKERREPO} && ./scripts/build-base-images.sh

develop-compile: check-env
	@echo "Building microservices for development within Docker build container with GOPATH set"
	docker run --rm --user `id -u`:`id -g` --env MODS="${MODS}" --env GO111MODULE --env XDG_CACHE_HOME=/tmp/.cache --env BRANCH=${BRANCH} --env TAG=${TAG} --env HTTP_PROXY=${HTTP_PROXY} --env HTTPS_PROXY=${HTTPS_PROXY} --env GOPATH=/repo/bin -v `pwd`:/repo ${BASEDOCKERREPO}${BUILD_BASE_IMAGE_NAME}:${BUILD_BASE_IMAGE_VERSION} /bin/sh -c "cd /repo; make compile"
	@echo "    Done."

develop: develop-compile build-containers
	@MODS=`echo ${MODS} | sed 's/ovnaction/ovn/;s/genericactioncontroller/gac/;s/orchestrator/orch/;'` ./scripts/push_to_registry.sh
