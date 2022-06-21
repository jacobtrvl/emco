#!/bin/bash

# update /home/vagrant/github.com/istio/ca-certs/files/istio-certs.yaml with the respecive signers
# kubectl get clusterissuers foobar -o jsonpath='{.spec.ca.secretName}' | xargs kubectl get secret -n cert-manager -o jsonpath='{.data.tls\.crt}' |base64 -d 
# kubectl get clusterissuers foo -o jsonpath='{.spec.ca.secretName}' | xargs kubectl get secret -n cert-manager -o jsonpath='{.data.tls\.crt}' |base64 -d 
# kubectl get clusterissuers bar -o jsonpath='{.spec.ca.secretName}' | xargs kubectl get secret -n cert-manager -o jsonpath='{.data.tls\.crt}' |base64 -d 


echo "============================================== Install Istio ================================================="
echo "=============================================================================================================="
istioctl install -f /home/vagrant/github.com/istio/ca-certs/files/istio-certs.yaml -y -s tag=latest -s hub=registry.fi.intel.com/staging
sudo sleep 15

echo "=============================================================================================================="
echo "============================================== create namespaces ============================================="
echo "=============================================================================================================="
kubectl create ns foo
kubectl apply -f /home/vagrant/github.com/istio/ca-certs/files/proxyconfig-foo.yaml
kubectl create ns bar
kubectl apply -f /home/vagrant/github.com/istio/ca-certs/files/proxyconfig-bar.yaml
sudo sleep 5
kubectl label ns foo istio-injection=enabled
kubectl label ns bar istio-injection=enabled

echo "=============================================================================================================="
echo "============================================= install workloads =============================================="
echo "=============================================================================================================="
kubectl apply -f /home/vagrant/github.com/istio/ca-certs/files/httpbin.yaml -n foo
kubectl apply -f /home/vagrant/github.com/istio/ca-certs/files/sleep.yaml -n foo
kubectl apply -f /home/vagrant/github.com/istio/ca-certs/files/httpbin.yaml -n bar
kubectl apply -f /home/vagrant/github.com/istio/ca-certs/files/sleep.yaml -n bar
sudo sleep 10 # for the pods to be available
echo "=============================================================================================================="
echo "============================= verify network cnnectivity inside namespace ===================================="
echo "=============================================================================================================="
kubectl exec -it "$(kubectl get pod -n foo -l app=sleep -o jsonpath={.items..metadata.name})" -n foo -c sleep -- curl http://httpbin.foo:8000/html
echo "=============================================================================================================="
kubectl exec -it "$(kubectl get pod -n bar -l app=sleep -o jsonpath={.items..metadata.name})" -n bar -c sleep -- curl http://httpbin.bar:8000/html

echo "=============================================================================================================="
echo "============================= verify network cnnectivity across namespace ===================================="
echo "=============================================================================================================="
kubectl exec -it "$(kubectl get pod -n bar -l app=sleep -o jsonpath={.items..metadata.name})" -n bar -c sleep -- curl http://httpbin.foo:8000/html
echo "=============================================================================================================="
kubectl exec -it "$(kubectl get pod -n foo -l app=sleep -o jsonpath={.items..metadata.name})" -n foo -c sleep -- curl http://httpbin.bar:8000/html
