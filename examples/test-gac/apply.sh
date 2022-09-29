#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

test_yaml=$1

./setup.sh create
source ../scripts/_status.sh
project=$(cat values.yaml | grep ProjectName: | sed -z 's/.*ProjectName: //')
logical_cloud_name=$(cat values.yaml | grep AdminCloud: | sed -z 's/.*AdminCloud: //')
deployment_intent_group_name=$(cat values.yaml | grep DeploymentIntent: | sed -z 's/.*DeploymentIntent: //')
composite_app=$(cat values.yaml | grep CompositeAppGac: | sed -z 's/.*CompositeAppGac: //')
emcoctl --config emco-cfg.yaml apply -f controllers.yaml -v values.yaml
emcoctl --config emco-cfg.yaml apply -f projects.yaml -v values.yaml
emcoctl --config emco-cfg.yaml apply -f clusters.yaml -v values.yaml
emcoctl --config emco-cfg.yaml apply -f logicalclouds.yaml -v values.yaml
emcoctl --config emco-cfg.yaml apply -f instantiatelc.yaml -v values.yaml
get_logical_cloud_apply_status emco-cfg.yaml $project $logical_cloud_name
emcoctl --config emco-cfg.yaml apply -f apps.yaml -v values.yaml
emcoctl --config emco-cfg.yaml apply -f intents.yaml -v values.yaml
emcoctl --config emco-cfg.yaml apply -f ${test_yaml} -v values.yaml
emcoctl --config emco-cfg.yaml apply -f instantiate.yaml -v values.yaml
get_deployment_intent_group_apply_status emco-cfg.yaml $project $composite_app $deployment_intent_group_name
