#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2021 Intel Corporation

if [ "$#" -lt 1 ] ; then
  echo "enter a destination directory"      
  exit  
fi

mkdir $1
if [ "$?" -ne 0 ] ; then
	echo "directory already exists"
	exit
fi

for f in slb ngfw sdewan
do
	kubectl get pod `kubectl get pod -lapp=$f --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'` -o yaml > $1/$f.yaml
	kubectl exec `kubectl get pod -lapp=$f --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'` -- route -n > $1/$f.route
	kubectl exec `kubectl get pod -lapp=$f --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'` -- ifconfig > $1/n$f.ifconfig
done

kubectl get -n kube-system cm nodus-dynamic-network-pool -o yaml > $1/nodus-dynamic-network-pool.yaml
for n in `kubectl get networks --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'`
do
	kubectl get network $n -o yaml > $1/$n.yaml
done

kubectl -n sfc-head get pod/`kubectl get pods -lsfc=head -n sfc-head  --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' | head -1` -o yaml > $1/nginx-left-live.yaml
kubectl -n sfc-tail get pod/`kubectl get pods -lsfc=tail -n sfc-tail  --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' | head -1` -o yaml > $1/nginx-right-live.yaml

for t in head tail
do
	for p in `kubectl get pods -lsfc=$t -n sfc-$t --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'`
	do
		kubectl -n sfc-$t exec $p -- route -n > $1/$p.route
		kubectl -n sfc-$t exec $p -- ifconfig > $1/$p.ifconfig
		kubectl -n sfc-$t get pod $p -o yaml > $1/$p.yaml
	done
	kubectl get ns sfc-$t -o yaml > $1/sfc-$t.yaml
done

for p in `kubectl get pods -lsfc=head2 -n sfc-head-two --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'`
do
       kubectl -n sfc-head-two exec $p -- route -n > $1/$p.route
       kubectl -n sfc-head-two exec $p -- ifconfig > $1/$p.ifconfig
       kubectl -n sfc-head-two get pod $p -o yaml > $1/$p.yaml
done
kubectl get ns sfc-head-two -o yaml > $1/sfc-head-two.yaml

for p in `kubectl get pods -lsfc=tail2 -n sfc-tail-two --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'`
do
       kubectl -n sfc-tail-two exec $p -- route -n > $1/$p.route
       kubectl -n sfc-tail-two exec $p -- ifconfig > $1/$p.ifconfig
       kubectl -n sfc-tail-two get pod $p -o yaml > $1/$p.yaml
done
kubectl get ns sfc-tail-two -o yaml > $1/sfc-tail-two.yaml

kubectl -n kube-system logs --tail 5000 -l name=nfn-operator  > $1/nfn-operator.log
kubectl -n kube-system logs --tail 5000 -l app=nfn-agent  > $1/nfn-agent.log
kubectl get providernetwork left-pnetwork -o yaml > $1/left-pnetwork-live.yaml
kubectl get providernetwork right-pnetwork -o yaml > $1/right-pnetwork-live.yaml
kubectl get ns sfc-head -o yaml > $1/sfc-head-live.yaml
kubectl get ns sfc-tail -o yaml > $1/sfc-tail-live.yaml
#kubectl -n sfc-head exec `kubectl get pod -lsfc=head -n sfc-head  --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' | head -1` -- traceroute -n -q 1 -I 172.30.22.4 > $1/traceroute.out
