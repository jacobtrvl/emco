#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020-2021 Intel Corporation
source ../../scripts/_status.sh

project=$(cat values.yaml | grep ProjectName: | sed -z 's/.*ProjectName: //')
logical_cloud_name=$(cat values.yaml | grep LogicalCloud: | sed -z 's/.*LogicalCloud: //')
deployment_intent_group_name=$(cat values.yaml | grep DeploymentIntentGroup: | sed -z 's/.*DeploymentIntentGroup: //')
composite_app=$(cat values.yaml | grep CompositeApp: | sed -z 's/.*CompositeApp: //')

emcoctl --config emco-cfg.yaml apply -f terminate.yaml -v values.yaml
get_deployment_intent_group_delete_status emco-cfg.yaml $project $composite_app $deployment_intent_group_name
emcoctl --config emco-cfg.yaml delete -f intents.yaml -v values.yaml
emcoctl --config emco-cfg.yaml delete -f apps.yaml -v values.yaml
emcoctl --config emco-cfg.yaml apply -f terminatelc.yaml -v values.yaml
get_logical_cloud_delete_status emco-cfg.yaml $project $logical_cloud_name
emcoctl --config emco-cfg.yaml delete -f logicalclouds.yaml -v values.yaml
emcoctl --config emco-cfg.yaml delete -f clusters.yaml -v values.yaml
emcoctl --config emco-cfg.yaml delete -f projects.yaml -v values.yaml
emcoctl --config emco-cfg.yaml delete -f controllers.yaml -v values.yaml
./setup.sh cleanup
