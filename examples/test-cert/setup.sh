#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

HOST_IP=${HOST_IP:-"oops"}
CLUSTER_1_KUBE_PATH=${KUBE_PATH:-"oops"}
ISSUING_CLUSTER_KUBE_PATH=${ISSUING_CLUSTER_KUBE_PATH:-"oops"}

function create {
# head of values.yaml
cat << NET > values.yaml
ProjectName: proj1
ClusterProvider: provider1
# issuing cluster
IssuingCluster: Issuer1
IssuingClusterConfig: $ISSUING_CLUSTER_KUBE_PATH
# edge clsuters
Cluster1: cluster1
Cluster2: cluster2
Cluster3: cluster3
Cluster4: cluster4
Cluster5: cluster5
ClusterLabel: edge-01
KubeConfig: $CLUSTER_1_KUBE_PATH
# cert intent
CertIntent: cert1
ClusterIssuer: foo
Kind: ClusterIssuer
Group: cert-manager.io
KeySize: 4096
# csr template
CommonName: foo
# clusters
EdgeClusters: edge-01
RsyncPort: 30431
HostIP: $HOST_IP
NET

# head of emco-cfg.yaml
cat << NET > emco-cfg.yaml
cacert:
  host: $HOST_IP
  port: 30436
orchestrator:
  host: $HOST_IP
  port: 30415
  statusPort: 30416
clm:
  host: $HOST_IP
  port: 30461
NET

# head of prerequisites.yaml
cat << NET > prerequisites.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# register rsync controller
version: emco/v2
resourceContext:
  anchor: controllers
metadata :
  name: rsync
spec:
  host:  {{.HostIP}}
  port: {{.RsyncPort}}
---
# create project
version: emco/v2
resourceContext:
  anchor: projects
metadata :
  name: {{.ProjectName}}
---
# create cluster provider
version: emco/v2
resourceContext:
  anchor: cluster-providers
metadata :
  name: {{.ClusterProvider}}
---
# create issuing cluster
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters
metadata :
  name: {{.IssuingCluster}}
file:
  {{.IssuingClusterConfig}}
---
# create edge cluster
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters
metadata :
  name: {{.Cluster1}}
file:
  {{.KubeConfig}}
---
# create edge cluster
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters
metadata :
  name: {{.Cluster2}}
file:
  {{.KubeConfig}}
---
# create edge cluster
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters
metadata :
  name: {{.Cluster3}}
file:
  {{.KubeConfig}}
---
# create edge cluster
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters
metadata :
  name: {{.Cluster4}}
file:
  {{.KubeConfig}}
---
# create edge cluster
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters
metadata :
  name: {{.Cluster5}}
file:
  {{.KubeConfig}}
---
# add cluster label
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters/{{.Cluster1}}/labels
clusterLabel: {{.ClusterLabel}}
---
# add cluster label
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters/{{.Cluster2}}/labels
clusterLabel: {{.ClusterLabel}}
---
# add cluster label
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters/{{.Cluster3}}/labels
clusterLabel: {{.ClusterLabel}}
---
# add cluster label
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters/{{.Cluster4}}/labels
clusterLabel: {{.ClusterLabel}}
---
# add cluster label
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters/{{.Cluster5}}/labels
clusterLabel: {{.ClusterLabel}}
---
# add the cert intent
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/ca-certs
metadata:
  name: {{.CertIntent}}
spec:
  isCA: true
  issuerRef:
    name: {{.ClusterIssuer}}
    kind: {{.Kind}}
    group: {{.Group}}
  duration: "8760h"
  issuingCluster: 
    cluster: {{.IssuingCluster}}
    clusterProvider: {{.ClusterProvider}}
  template: 
    keySize: {{.KeySize}}
    version: 1
    dnsNames: []
    emailAddresses: []
    keyUsages: []
    algorithm: 
      publicKeyAlgorithm: RSA
      signatureAlgorithm: SHA512WithRSA
    subject:
      locale: 
        country: []
        locality: []
        postalCode: []
        province: []
        streetAddress: []
      names:
        commonName: {{.CommonName}}
      organization:
        names: []
        units: []
---
# add the cluster part of the cert intent
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/ca-certs/{{.CertIntent}}/clusters
metadata :
  name: {{.EdgeClusters}}
spec:
  scope: label
  label: {{.ClusterLabel}}
  name: {{.Cluster1}}
NET
# head of update-cert.yaml
cat << NET > update-cert.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# update edge cluster
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters
metadata :
  name: cluster100
file:
  {{.KubeConfig}}
---
# add cluster label
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters/cluster100/labels
clusterLabel: {{.ClusterLabel}}
---
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters
metadata :
  name: cluster200
file:
  {{.KubeConfig}}
---
# add cluster label
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters/cluster200/labels
clusterLabel: {{.ClusterLabel}}
NET

# head of instantiate.yaml
cat << NET > instantiate_enrollment.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# instantiate
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/ca-certs/{{.CertIntent}}/enrollment/instantiate
NET

# head of update.yaml
cat << NET > update_enrollment.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# update
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/ca-certs/{{.CertIntent}}/enrollment/update
NET

# head of terminate.yaml
cat << NET > terminate_enrollment.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# terminate
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/ca-certs/{{.CertIntent}}/enrollment/terminate
NET

# head of instantiate_distribution.yaml
cat << NET > instantiate_distribution.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# instantiate
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/ca-certs/{{.CertIntent}}/distribution/instantiate
NET

# head of update_distribution.yaml
cat << NET > update_distribution.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# update
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/ca-certs/{{.CertIntent}}/distribution/update
NET

# head of terminate_distribution.yaml
cat << NET > terminate_distribution.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# terminate
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/ca-certs/{{.CertIntent}}/distribution/terminate
NET

}

function usage {
    echo "Usage: $0  create|cleanup"
}

function cleanup {
    rm -f *.tar.gz
    rm -f values.yaml
    rm -f emco-cfg.yaml
    rm -f prerequisites.yaml
    rm -f instantiate.yaml
    rm -f update.yaml
    rm -f rollback.yaml
    rm -rf output
}

if [ "$#" -lt 1 ] ; then
    usage
    exit
fi

case "$1" in
    "create" )
        if [ "${HOST_IP}" == "oops" ] || [ "${KUBE_PATH}" == "oops" ] ; then
            echo -e "ERROR - HOST_IP & KUBE_PATH environment variable needs to be set"
        else
            create
        fi
        ;;
    "cleanup" )
        cleanup
    ;;
    *)
        usage ;;
esac
