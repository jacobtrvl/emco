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

# Issuing Cluster
IssuingCluster: issuer1
IssuingClusterConfig: /home/vagrant/github.com/cluster-configs/issuing-cluster-config

# Clsuters
# Cluster1: cluster1
# KubeConfig1: /home/vagrant/github.com/cluster-configs/edge01-1-config
# Cluster2: cluster2
# KubeConfig2: /home/vagrant/github.com/cluster-configs/edge02-1-config
Cluster3: cluster3
KubeConfig3: /home/vagrant/github.com/cluster-configs/cluster-1-config
Cluster4: cluster4
KubeConfig4: /home/vagrant/github.com/cluster-configs/cluster-2-config
Cluster5: cluster5
KubeConfig5: /home/vagrant/github.com/cluster-configs/cluster-3-config

# Cluster Label
GroupLabel1: group1
GroupLabel2: group2

# Cert
CertIntent1: cert1
CertIntent2: cert2
CertIntent3: cert3

# Cluster Issuer
ClusterIssuer: foo
Kind: ClusterIssuer
Group: cert-manager.io

# CSR Info
KeySize: 4096
CommonNamePrefix: foo

# Cluster Group
EdgeClusters: edge
EdgeClusters1: edge1
EdgeClusters2: edge2
RsyncPort: 30431
HostIP: 192.168.121.3

# Logical-Cloud
AdminCloud: default
FooLogicalCloud: foo
BarLogicalCloud: bar
Cluster1Ref: cluster1-ref
FooCloud: foo-ns
BarCloud: bar-ns
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
dcm:
  host: $HOST_IP
  port: 30477
NET

# head of cp_prerequisites.yaml
cat << NET > cp_prerequisites.yaml
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
  clusterProvider: {{.ClusterProvider}}
NET
# head of cp_update_cert.yaml
cat << NET > cp_update_cert.yaml
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

# head of cp_enrollment_instantiate.yaml
cat << NET > cp_enrollment_instantiate.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# instantiate
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/ca-certs/{{.CertIntent}}/enrollment/instantiate
NET

# head of cp_enrollment_update.yaml
cat << NET > cp_enrollment_update.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# update
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/ca-certs/{{.CertIntent}}/enrollment/update
NET

# head of cp_enrollment_terminate.yaml
cat << NET > cp_enrollment_terminate.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# terminate
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/ca-certs/{{.CertIntent}}/enrollment/terminate
NET

# head of cp_distribution_instantiate.yaml
cat << NET > cp_distribution_instantiate.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# instantiate
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/ca-certs/{{.CertIntent}}/distribution/instantiate
NET

# head of cp_distribution_update.yaml
cat << NET > cp_distribution_update.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# update
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/ca-certs/{{.CertIntent}}/distribution/update
NET

# head of cp_distribution_terminate.yaml
cat << NET > cp_distribution_terminate.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# terminate
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/ca-certs/{{.CertIntent}}/distribution/terminate
NET

# head of lc_prerequisites.yaml
cat << NET > lc_prerequisites.yaml
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
# create project
version: emco/v2
resourceContext:
  anchor: projects
metadata :
  name: {{.ProjectName}}
---
# add the cert intent
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/ca-certs
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
# add logical cloud part of the cert intent
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/ca-certs/{{.CertIntent}}/logical-clouds
metadata :
  name: {{.LogicalCloud1}}
spec:
  name: {{.LogicalCloud1}}
---

# # add logical cloud part of the cert intent
# version: emco/v2
# resourceContext:
#   anchor: projects/{{.ProjectName}}/ca-certs/{{.CertIntent}}/logical-clouds
# metadata :
#   name: {{.LogicalCloud2}}
# spec:
#   name: {{.LogicalCloud2}}
# ---

# add the cluster part of the cert intent
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/ca-certs/{{.CertIntent}}/logical-clouds/{{.LogicalCloud1}}/clusters
metadata :
  name: {{.EdgeClusters1}}
spec:
  scope: label
  label: {{.ClusterLabel}}
  name: {{.Cluster1}}
  clusterProvider: {{.ClusterProvider}}
---

# # add the cluster part of the cert intent
# version: emco/v2
# resourceContext:
#   anchor: projects/{{.ProjectName}}/ca-certs/{{.CertIntent}}/logical-clouds/{{.LogicalCloud1}}/clusters
# metadata :
#   name: {{.EdgeClusters2}}
# spec:
#   scope: label
#   label: {{.ClusterLabel}}
#   name: {{.Cluster2}}
#   clusterProvider: {{.ClusterProvider}}
# ---

# # add the cluster part of the cert intent
# version: emco/v2
# resourceContext:
#   anchor: projects/{{.ProjectName}}/ca-certs/{{.CertIntent}}/logical-clouds/{{.LogicalCloud2}}/clusters
# metadata :
#   name: {{.EdgeClusters1}}
# spec:
#   scope: label
#   label: {{.ClusterLabel}}
#   name: {{.Cluster1}}
#   clusterProvider: {{.ClusterProvider}}
# ---
# # add the cluster part of the cert intent
# version: emco/v2
# resourceContext:
#   anchor: projects/{{.ProjectName}}/ca-certs/{{.CertIntent}}/logical-clouds/{{.LogicalCloud2}}/clusters
# metadata :
#   name: {{.EdgeClusters2}}
# spec:
#   scope: label
#   label: {{.ClusterLabel}}
#   name: {{.Cluster2}}
#   clusterProvider: {{.ClusterProvider}}

NET

# head of lc_cert_update.yaml
cat << NET > lc_update_cert.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# update edge cluster
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/ca-certs/{{.CertIntent}}/logical-clouds/{{.LogicalCloud2}}/clusters
metadata :
  name: cluster100
file:
  {{.KubeConfig}}
---
# add cluster label
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/ca-certs/{{.CertIntent}}/logical-clouds/{{.LogicalCloud2}}/clusters/cluster100/labels
clusterLabel: {{.ClusterLabel}}
---
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/ca-certs/{{.CertIntent}}/logical-clouds/{{.LogicalCloud2}}/clusters
metadata :
  name: cluster200
file:
  {{.KubeConfig}}
---
# add cluster label
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/ca-certs/{{.CertIntent}}/logical-clouds/{{.LogicalCloud2}}/clusters/cluster200/labels
clusterLabel: {{.ClusterLabel}}
NET

# head of setup_logical_cloud.yaml
cat << NET > setup_logical_cloud.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

# create project
version: emco/v2
resourceContext:
  anchor: projects
metadata :
  name: {{.ProjectName}}
---
#create admin logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds
metadata:
  name: {{.AdminCloud}}
spec:
  level: "0"
---
#add cluster reference to logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.AdminCloud}}/cluster-references
metadata:
  name: lc-cl-1
spec:
  clusterProvider: {{.ClusterProvider}}
  cluster: {{.IssuingCluster}}
  loadbalancerIp: "0.0.0.0"
---
#create foo logical cloud without admin permissions
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds
metadata:
  name: {{.FooLogicalCloud}}
spec:
  namespace: {{.FooCloud}}
---
#create cluster quotas
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.FooLogicalCloud}}/cluster-quotas
metadata:
    name: foo-quota
spec:
    requests.ephemeral-storage: '500'
    limits.ephemeral-storage: '500'
    persistentvolumeclaims: '500'
    pods: '500'
    configmaps: '1000'
    replicationcontrollers: '500'
    resourcequotas: '500'
    services: '500'
    services.loadbalancers: '500'
    services.nodeports: '500'
    secrets: '500'
    count/replicationcontrollers: '500'
    count/deployments.apps: '500'
    count/replicasets.apps: '500'
    count/statefulsets.apps: '500'
    count/jobs.batch: '500'
    count/cronjobs.batch: '500'
    count/deployments.extensions: '500'
---
#create foo logical cloud without admin permissions
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.FooLogicalCloud}}/user-permissions
metadata:
  name: foo-permission
spec:
    namespace: {{.FooCloud}}
    apiGroups:
    - ""
    - "apps"
    - "k8splugin.io"
    resources:
    - secrets
    - pods
    - configmaps
    - services
    - deployments
    - resourcebundlestates
    verbs:
    - get
    - watch
    - list
    - create
    - delete
---
#add cluster reference to foo logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.FooLogicalCloud}}/cluster-references
metadata:
  name: lc-cl-foo
spec:
  clusterProvider: {{.ClusterProvider}}
  cluster: {{.IssuingCluster}}
  loadbalancerIp: "0.0.0.0"
---
#create bar logical cloud without admin permissions
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds
metadata:
  name: {{.BarLogicalCloud}}
spec:
  namespace: {{.BarCloud}}
---
#create cluster quotas
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.BarLogicalCloud}}/cluster-quotas
metadata:
    name: bar-quota
spec:
    requests.ephemeral-storage: '500'
    limits.ephemeral-storage: '500'
    persistentvolumeclaims: '500'
    pods: '500'
    configmaps: '1000'
    replicationcontrollers: '500'
    resourcequotas: '500'
    services: '500'
    services.loadbalancers: '500'
    services.nodeports: '500'
    secrets: '500'
    count/replicationcontrollers: '500'
    count/deployments.apps: '500'
    count/replicasets.apps: '500'
    count/statefulsets.apps: '500'
    count/jobs.batch: '500'
    count/cronjobs.batch: '500'
    count/deployments.extensions: '500'
---
#create foo logical cloud without admin permissions
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.BarLogicalCloud}}/user-permissions
metadata:
  name: bar-permission
spec:
    namespace: {{.BarCloud}}
    apiGroups:
    - ""
    - "apps"
    - "k8splugin.io"
    resources:
    - secrets
    - pods
    - configmaps
    - services
    - deployments
    - resourcebundlestates
    verbs:
    - get
    - watch
    - list
    - create
    - delete
---
#add cluster reference to foo logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.BarLogicalCloud}}/cluster-references
metadata:
  name: lc-cl-bar
spec:
  clusterProvider: {{.ClusterProvider}}
  cluster: {{.IssuingCluster}}
  loadbalancerIp: "0.0.0.0"
---
#instantiate Admin logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.AdminCloud}}/instantiate
---
#instantiate Foo logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.FooLogicalCloud}}/instantiate
---
#instantiate bar logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.BarLogicalCloud}}/instantiate
NET

# head of lc_enrollment_instantiate.yaml
cat << NET > lc_enrollment_instantiate.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

#instantiate logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/ca-certs/{{.CertIntent}}/enrollment/instantiate
NET

# head of lc_enrollment_update.yaml
cat << NET > lc_enrollment_update.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

#instantiate logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/ca-certs/{{.CertIntent}}/enrollment/update
NET

# head of lc_enrollment_terminate.yaml
cat << NET > lc_enrollment_terminate.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

#instantiate logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/ca-certs/{{.CertIntent}}/enrollment/terminate
NET

# head of lc_distribution_instantiate.yaml
cat << NET > lc_distribution_instantiate.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

#instantiate logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/ca-certs/{{.CertIntent}}/distribution/instantiate
NET

# head of lc_distribution_update.yaml
cat << NET > lc_distribution_update.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

#instantiate logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/ca-certs/{{.CertIntent}}/distribution/update
NET

# head of lc_distribution_terminate.yaml
cat << NET > lc_distribution_terminate.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

#instantiate logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/ca-certs/{{.CertIntent}}/distribution/terminate
NET


}

function usage {
    echo "Usage: $0  create|cleanup"
}

function cleanup {
     rm -f values.yaml
    rm -f emco-cfg.yaml

    rm -f cp_prerequisites.yaml
    rm -f cp_enrollment_instantiate.yaml
    rm -f cp_enrollment_update.yaml
    rm -f cp_enrollment_terminate.yaml
    rm -f cp_distribution_instantiate.yaml
    rm -f cp_distribution_update.yaml
    rm -f cp_distribution_terminate.yaml

    rm -f cp_update_cert.yaml

    rm -f setup_logical_cloud.yaml

    rm -f lc_prerequisites.yaml
    rm -f lc_enrollment_instantiate.yaml
    rm -f lc_enrollment_update.yaml
    rm -f lc_enrollment_terminate.yaml
    rm -f lc_distribution_instantiate.yaml
    rm -f lc_distribution_update.yaml
    rm -f lc_distribution_terminate.yaml

    rm -f lc_update_cert.yaml
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
