#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020-2021 Intel Corporation

test_yaml=$1

source ../scripts/_status.sh

project=$(cat values.yaml | grep ProjectName: | sed -z 's/.*ProjectName: //')
logical_cloud_name=$(cat values.yaml | grep AdminCloud: | sed -z 's/.*AdminCloud: //')
deployment_intent_group_name=$(cat values.yaml | grep DeploymentIntent: | sed -z 's/.*DeploymentIntent: //')
composite_app=$(cat values.yaml | grep CompositeAppGac: | sed -z 's/.*CompositeAppGac: //')

emcoctl --config emco-cfg.yaml apply -f terminate.yaml -v values.yaml
get_deployment_intent_group_delete_status emco-cfg.yaml $project $composite_app $deployment_intent_group_name
emcoctl --config emco-cfg.yaml delete -f ${test_yaml} -v values.yaml
emcoctl --config emco-cfg.yaml delete -f intents.yaml -v values.yaml
emcoctl --config emco-cfg.yaml delete -f apps.yaml -v values.yaml
emcoctl --config emco-cfg.yaml apply -f terminatelc.yaml -v values.yaml
get_logical_cloud_delete_status emco-cfg.yaml $project $logical_cloud_name
emcoctl --config emco-cfg.yaml delete -f logicalclouds.yaml -v values.yaml
emcoctl --config emco-cfg.yaml delete -f clusters.yaml -v values.yaml
emcoctl --config emco-cfg.yaml delete -f projects.yaml -v values.yaml
emcoctl --config emco-cfg.yaml delete -f controllers.yaml -v values.yaml
./setup.sh cleanup
