#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020-2021 Intel Corporation

./setup.sh create
emcoctl --config emco-cfg.yaml apply -f controllers.yaml -v values.yaml
emcoctl --config emco-cfg.yaml apply -f projects.yaml -v values.yaml
emcoctl --config emco-cfg.yaml apply -f clusters.yaml -v values.yaml
emcoctl --config emco-cfg.yaml apply -f logicalclouds.yaml -v values.yaml
emcoctl --config emco-cfg.yaml apply -f apps.yaml -v values.yaml
emcoctl --config emco-cfg.yaml apply -f intents.yaml -v values.yaml
sleep 10
emcoctl --config emco-cfg.yaml apply -f instantiate.yaml -v values.yaml
