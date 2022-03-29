#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2021 Intel Corporation

function delay {
	echo "waiting for $1 $2"

	for i in $(eval echo {$1..1})
	do
		echo -en "\r$i    "
		kubectl get networkchaining -o yaml | grep state:
		sleep 1
	done
	echo "    "
}

kubectl apply -f namespace-left.yaml
kubectl apply -f namespace-right.yaml
delay 5  "before applying networks"
kubectl apply -f sfc-virtual-network.yaml
delay 1 "before applying slb"
kubectl apply -f slb-multiple-network.yaml
delay 1 "before applying ngfw"
kubectl apply -f ngfw.yaml
delay 1 "before applying sdewan"
kubectl apply -f sdewan-multiple-network.yaml
delay 1 "before applying sfc"
kubectl apply -f nginx-left-deployment.yaml
delay 1 "before applying nginx right"
kubectl apply -f nginx-right-deployment.yaml
delay 1 "before applying nginx left"
kubectl apply -f sfc-with-virtual-and-provider-network.yaml
delay 20 "before applying nginx left"
