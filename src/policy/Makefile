# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Aarna Networks, Inc.

all: 
	$(MAKE) -C dbhelper all
	$(MAKE) -C sacontroller all
	cp run_policy ../../bin/policy/run_policy 

build: clean test cover
deploy: build

.PHONY: test
test: 
	$(MAKE) -C dbhelper test
	$(MAKE) -C sacontroller test

format:
	$(MAKE) -C dbhelper format
	$(MAKE) -C sacontroller format

clean:
	$(MAKE) -C dbhelper clean
	$(MAKE) -C sacontroller clean

.PHONY: cover
cover:
	$(MAKE) -C dbhelper cover
	$(MAKE) -C sacontroller cover
