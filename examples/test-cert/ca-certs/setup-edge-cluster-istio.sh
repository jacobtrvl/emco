#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# this script sets up a cert-manager clusterissuer
# that can be used as the default cert signer
# for istio when installed with cert-manager
# as the external signer.

# additionally, it installs Istio with the Istio
# mesh config set up to use the certificate that is
# created by this script.

# clean up previously created files
sudo rm -r ./istio-system/csr
sudo rm -r ./istio-system/istio
sudo rm -r ./istio-system/root
rm istio-install.yaml

kubectl delete secret -n cert-manager istio-system-ca-secret
kubectl delete clusterIssuer istio-system

echo "========================================================================================="
pushd ./istio-system/
mkdir root csr istio

echo "========= Create istio-system Root CA ========="
openssl genrsa -out root/istio-system-ca-key.pem 4096
openssl req -new -key root/istio-system-ca-key.pem -config root-ca.conf -out root/istio-system-ca.csr
openssl x509 -req -days 3650 -signkey root/istio-system-ca-key.pem -extensions req_ext -extfile root-ca.conf -in root/istio-system-ca.csr -out root/istio-system-ca-cert.pem

echo "========= Create istio-system Root Secret "
cp root/istio-system-ca-cert.pem root/tls.crt
cp root/istio-system-ca-key.pem root/tls.key

# secret
kubectl create secret generic istio-system-ca-secret -n cert-manager --from-file=root/tls.crt --from-file=root/tls.key
sudo sleep 10 # for the secret to be available

echo "========= Create istio-system ClusterIssuer "
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: istio-system
spec:
  ca:
    secretName: istio-system-ca-secret
EOF
popd
echo "=========================================================================================="

echo "-- Wait for clusterissuer"
sudo sleep 5

kubectl get clusterissuer -o wide
kubectl get secret -n cert-manager --sort-by=.metadata.creationTimestamp


# Install Istio with cert-manager configured as the external signer and
# the caCertificate configured using the Secret created above.
cat << EOF > istio-install.yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  meshConfig:
    defaultConfig:
      proxyMetadata:
        ISTIO_META_CERT_SIGNER: istio-system
    caCertificates:
    - pem: |
$(kubectl get secret -n cert-manager istio-system-ca-secret -o jsonpath='{.data.tls\.crt}' |base64 -d | sed -e 's;\(.*\);        \1;g')
      certSigners: 
      - clusterissuers.cert-manager.io/istio-system
  components:
    pilot:
      k8s:
        env:
        - name: CERT_SIGNER_DOMAIN
          value: clusterissuers.cert-manager.io
        - name: EXTERNAL_CA
          value: ISTIOD_RA_KUBERNETES_API
        - name: PILOT_CERT_PROVIDER
          value: k8s.io/clusterissuers.cert-manager.io/istio-system
        overlays:
          # Amend ClusterRole to add permission for istiod to approve certificate signing by custom signer
          - kind: ClusterRole
            name: istiod-clusterrole-istio-system
            patches:
              - path: rules[-1]
                value: |
                  apiGroups:
                  - certificates.k8s.io
                  resourceNames:
                  - clusterissuers.cert-manager.io/*
                  resources:
                  - signers
                  verbs:
                  - approve
  values:
    pilot:
      image: pilot
    global:
      proxy:
        image: proxyv2
EOF
