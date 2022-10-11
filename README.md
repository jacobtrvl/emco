[//]: # "SPDX-License-Identifier: Apache-2.0"
[//]: # "Copyright (c) 2019-2022 Intel Corporation"

# EMCO (Edge Multi-Cloud Orchestrator)

- [EMCO (Edge Multi-Cloud Orchestrator)](#emco-edge-multi-cloud-orchestrator)
  - [Overview](#overview)
  - [Installation](#installation)
    - [Requirements](#requirements)
    - [Using Helm](#using-helm)
      - [Tuning & Compatibility](#tuning--compatibility)
      - [Known issues](#known-issues)
    - [From source](#from-source)
    - [Client/CLI](#clientcli)

## Overview

The Edge Multi-Cluster Orchestrator (EMCO) is a software framework for
intent-based deployment of cloud-native applications to a set of Kubernetes
clusters, spanning enterprise data centers, multiple cloud service providers
and numerous edge locations. It is architected to be flexible, modular and
highly scalable. It is aimed at various verticals, including telecommunication
service providers.

Refer to [EMCO documentation](docs/design/emco-design.md) for details on EMCO architecture.

## Installation

### Requirements

In general, to install and use EMCO, you will need at least **2 Kubernetes clusters**. One cluster to run the EMCO microservices themselves, and one to many Kubernetes clusters where the applications to be deployed by EMCO will reside.

Additionally, each of the Kubernetes where applications (and [Logical Clouds](docs/design/Logical_Clouds.md)) reside also require that the [EMCO Monitor](docs/design/monitor.md) (also known as EMCO Status Monitoring) service be running. Instructions to deploy Monitor are provided in this readme file.

Refer to the [Release Notes](ReleaseNotes.md) for a tested compatibility table between versions of EMCO and versions of Kubernetes, Helm, and others.


### Using Helm

Using the official EMCO Helm charts for EMCO is the recommended installation/deployment method for first-timers.

This isn't the only way of deploying EMCO using Helm, but it's the only one using pre-built Helm charts. As such, there is no cloning of source code involved.

Before attempting to install, make sure you already have Kubernetes cluster available and that the  `$KUBECONFIG` environment variable is set to its kubeconfig file path.

**Add EMCO's official Helm repository:**
```
helm repo add emco https://gitlab.com/api/v4/projects/29353813/packages/helm/stable
helm repo update
```

Here's how to list the public Helm charts:
```
helm search repo
```
```
NAME                    CHART VERSION   APP VERSION     DESCRIPTION
emco/emco               1.0.1                           EMCO All-In-One Package
emco/emco-db            1.0.0                           EMCO Database Package
emco/emco-services      1.0.1                           EMCO Services Package
emco/emco-tools         1.0.0                           EMCO Tools Package
emco/monitor            1.0.0                           EMCO Status Monitoring
```

EMCO Helm charts don't contain an app version since the specific EMCO version can be specified using `global.emcoTag` (for the main EMCO services) or simply `emcoTag` (for EMCO Status Monitoring). More on this below.

**Install EMCO on the Kubernetes cluster (on the `emco` namespace):**
```
kubectl create namespace emco
helm install emco -n emco emco/emco \
  --set global.db.emcoPassword=SETPASS \
  --set global.db.rootPassword=SETPASS \
  --set global.contextdb.rootPassword=SETPASS \
  --set global.emcoTag=latest
```

Replace `SETPASS` with the your choice of passwords for MongoDB and etcd, respectively. You may also choose to not set custom passwords, in which case they will be randomized. If you choose random passwords, make sure to check the [Known issues](#known-issues).

Replace `latest` with the version of EMCO you wish to deploy, such as `22.06`. If you don't set this field, the `latest` EMCO container images will be installed. Typically, the `latest` tag is updated once a day.

*Keep in mind that different EMCO versions require different Helm chart version in order to unlock to full set of features. For example, installing `emco` chart `1.0.0` while setting the `emcoTag` to `22.06` will not enable the newly-introduced action controllers in 22.06. Equivalently, setting installing `emco` chart `1.0.1` while setting the `emcoTag` to `22.03.1` will result in some pods coming back as `ImagePullBackOff` since EMCO 22.03.1 doesn't contain all of the expected action controllers. See issue #240.*

Release `22.06` introduces an optional database encryption feature, which is not enabled by default.  To enable it, set the following flags to the `helm install` command:

- `--set global.enableMongoSecret=true`  Enable the encryption feature
- `--set global.db.dataSecret=<secret value>` (optionally) set the value for the secret which is used to generate the key.  If not provided, helm will autogenerate a key.

Installation should take a handful of minutes to complete, as the multiple pods will be brought up, including etcd and mongo, and initialization takes place.
The temporary restarts and probe failures you may witness are expected and relate to the initialization of mongo, etcd, and instantiating the [Referential Integrity](docs/developer/ReferentialIntegrity.md) schema.

Confirm that EMCO services are up and running:

```
kubectl get pods -A
```
```
NAMESPACE     NAME                                  READY   STATUS    RESTARTS       AGE
emco          emco-clm-5c6745b964-4w8mp             1/1     Running   3 (86s ago)    2m35s
emco          emco-dcm-b459f6f45-2flnn              1/1     Running   3 (85s ago)    2m35s
emco          emco-dtc-85477bc696-g6tvw             1/1     Running   3 (77s ago)    2m35s
emco          emco-emco-mongo-0                     1/1     Running   0              2m35s
emco          emco-etcd-0                           1/1     Running   0              2m35s
emco          emco-fluentd-0                        1/1     Running   0              2m35s
emco          emco-fluentd-7q5xf                    1/1     Running   3 (2m8s ago)   2m35s
emco          emco-gac-5ffbf484b7-nqs2s             1/1     Running   3 (76s ago)    2m35s
emco          emco-hpa-ac-76574fdf47-8bndh          1/1     Running   3 (81s ago)    2m35s
emco          emco-hpa-plc-5b9c85ddd7-th7bd         1/1     Running   3 (83s ago)    2m35s
emco          emco-its-74675669f7-vvprf             1/1     Running   3 (73s ago)    2m35s
emco          emco-ncm-56bbffbc67-svgvf             1/1     Running   3 (89s ago)    2m34s
emco          emco-nps-7d6c99959f-jng29             1/1     Running   3 (87s ago)    2m34s
emco          emco-orchestrator-cbd5f4cdf-pbnq6     1/1     Running   3 (74s ago)    2m34s
emco          emco-ovnaction-767b49f466-lvbwx       1/1     Running   3 (90s ago)    2m34s
emco          emco-rsync-67bd699cdb-jx4qc           1/1     Running   3 (75s ago)    2m34s
emco          emco-sds-58b48f74f5-hrmwn             1/1     Running   2 (98s ago)    2m33s
emco          emco-sfc-7fcbb94fb7-hzz45             1/1     Running   3 (91s ago)    2m33s
emco          emco-sfcclient-685b99c45b-hqghd       1/1     Running   2 (88s ago)    2m33s
emco          emco-workflowmgr-86c9cd8fbf-b9t95     1/1     Running   1 (2m1s ago)   2m33s
kube-system   coredns-64897985d-gp6d4               1/1     Running   0              3d23h
kube-system   coredns-64897985d-xwcbx               1/1     Running   0              3d23h
kube-system   etcd-frostcanyon                      1/1     Running   1              3d23h
kube-system   kube-apiserver-frostcanyon            1/1     Running   1              3d23h
kube-system   kube-controller-manager-frostcanyon   1/1     Running   0              3d23h
kube-system   kube-flannel-ds-nvgds                 1/1     Running   0              3d23h
kube-system   kube-proxy-9gplv                      1/1     Running   0              3d23h
kube-system   kube-scheduler-frostcanyon            1/1     Running   1              3d23h
```

You should see similar output with all emco-* services running and 1/1 ready. The output above is from a single-node (all-in-one) deployment of Kubernetes 1.23.6.

**Install Monitor:**

The EMCO Monitor service needs to be deployed in each of the clusters that will be used as a target of applications and Logical Clouds by EMCO.

Set your `KUBECONFIG` (or take equivalent actions) according to each of the clusters you want to use as an EMCO target edge cluster, and install Monitor:
```
kubectl create namespace emco
helm install monitor -n emco emco/monitor \
  --set emcoTag=latest
```

Replace `latest` with the version of the EMCO Monitor you wish to deploy. This version must match the version of EMCO installed in the main EMCO cluster, or expect *unexpected* behavior. If you don't set this field, the `latest` EMCO Monitor container image will be installed. Typically, the `latest` tag is updated once a day.
> Notice that for Monitor, the image tag version is specified with `emcoTag` unlike in the main EMCO cluster, where it's specified with `global.emcoTag`.


#### Tuning & Compatibility

In some distributions of Kubernetes, additional modifications may be needed.

For example, in order to deploy EMCO with OpenShift, after creating the `emco` namespace, you will need the following rolebinding:
```
kubectl -n emco create rolebinding \
  system:openshift:scc:privileged-emco \
  --clusterrole=system:openshift:scc:privileged \
  --group=system:serviceaccounts:emco
```

And the `global.securityContext.privileged=true` flag while installing via Helm, as such:
```
helm install emco -n emco emco/emco \
  --set global.securityContext.privileged=true \
  --set global.db.emcoPassword=SETPASS \
  --set global.db.rootPassword=SETPASS \
  --set global.contextdb.rootPassword=SETPASS \
  --set global.emcoTag=latest
```

#### Known issues

**`unable to authenticate using mechanism \"SCRAM-SHA-256\"`**
If your EMCO pods are not getting ready and logs show the error above, there is an authentication problem between them and MongoDB. Usually, this is only seen when running EMCO with persistence enabled (which is the default), for the 2nd+ time. The MongoDB data store won't be deleted after the 1st deployment, which means that unless the EMCO microservices are configured to have the exact same authentication method/credentials as the 1st deployment, then they won't be able to authenticate to the (existing) MongoDB data store. Usually one will hit this issue installing EMCO via Helm without setting a fixed password, which defaults to using random passwords. The random password will be ignored by MongoDB on the 2nd+ deployment, since an existing data store already exists (via the persistent volume).

See the [EMCO Helm Tutorial](deployments/helm/Tutorial_Helm.md) for additional insight into this and additional issues.

### From source

If you wish to build and deploy EMCO from source and/or customize/build local Helm charts, check [EMCO Build & Deploy](docs/design/emco-design.md).

### Client/CLI


When it comes to installing the EMCO Client (a command-line interface), `emcoctl`, there are currently 3 tested options:

**From a downloadable release binary:**
A pre-built executable binary for EMCO 22.06 (linux/amd64) is hosted on GitLab and ready to use:
```
wget https://gitlab.com/project-emco/core/emco-base/-/package_files/46736477/download -O emcoctl-linux-amd64
wget https://gitlab.com/project-emco/core/emco-base/-/package_files/46736528/download -O emcoctl-linux-amd64.sha256
sha256sum -c emcoctl-linux-amd64.sha256
sudo cp emcoctl-linux-amd64 /usr/local/bin/emcoctl
sudo chmod +x /usr/local/bin/emcoctl
```

**From source using Go directly:**
```
git clone git@gitlab.com:project-emco/core/emco-base.git
cd emco-base
MODS=tools/emcoctl make all
```

**From source with the help of Docker:**
```
git clone git@gitlab.com:project-emco/core/emco-base.git
cd emco-base
MODS=tools/emcoctl make deploy-compile
```

Either of the *source* commands above will place the `emcoctl` executable binary in `./emco-base/bin/emcoctl/emcoctl`.
To make it easy to call `emcoctl`, load it into your `$PATH` or copy into `/usr/local/bin`.

Example for updating the `$PATH`:
```
echo 'export PATH=/home/user/emco-base/bin/emcoctl:$PATH' >> /home/user/.profile
source /home/user/.profile
```

**Test that `emcoctl` works:**
```
# emcoctl help
Warning: No Configuration File found. Using defaults
Emcoctl - CLI for EMCO

Usage:
  emco [flags]
  emco [command]

Available Commands:
  apply       apply(Post) the resources from input file or url(without body) from command line
  completion  Generate the autocompletion script for the specified shell
  delete      Delete the resources from input file or url(without body) from command line
  get         Get the resources from input file or url from command line
  help        Help about any command
  update      update(Put) the resources from input file or url(without body) from command line
  watch       Watch for resource status notifications from input file or url from command line

Flags:
      --config string   config file (default is $HOME/.emco.yaml)
  -h, --help            help for emco

Use "emco [command] --help" for more information about a command.
```