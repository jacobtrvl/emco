#!/bin/bash

# delete test apps
kubectl delete -f /home/vagrant/github.com/istio/ca-certs/files/httpbin.yaml -n foo
kubectl delete -f /home/vagrant/github.com/istio/ca-certs/files/httpbin.yaml -n bar
kubectl delete -f /home/vagrant/github.com/istio/ca-certs/files/sleep.yaml -n foo
kubectl delete -f /home/vagrant/github.com/istio/ca-certs/files/sleep.yaml -n bar

# delete proxyconfigs
kubectl delete -f /home/vagrant/github.com/istio/ca-certs/files/proxyconfig-bar.yaml
kubectl delete -f /home/vagrant/github.com/istio/ca-certs/files/proxyconfig-foo.yaml

#  delete namespaces
kubectl delete ns bar
kubectl delete ns foo

# uninstall istio
istioctl x uninstall --purge -y
