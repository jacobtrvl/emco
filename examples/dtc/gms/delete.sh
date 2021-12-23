#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020-2021 Intel Corporation

emcoctl --config emco-cfg.yaml apply -f terminate.yaml -v values.yaml
sleep 10
emcoctl --config emco-cfg.yaml delete -f intents.yaml -v values.yaml
emcoctl --config emco-cfg.yaml delete -f apps.yaml -v values.yaml
emcoctl --config emco-cfg.yaml delete -f logicalclouds.yaml -v values.yaml
emcoctl --config emco-cfg.yaml delete -f clusters.yaml -v values.yaml
emcoctl --config emco-cfg.yaml delete -f projects.yaml -v values.yaml
emcoctl --config emco-cfg.yaml delete -f controllers.yaml -v values.yaml
./setup.sh cleanup
