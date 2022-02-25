#### SPDX-License-Identifier: Apache-2.0
#### Copyright (c) 2021 Intel Corporation

# Running EMCO testcases with emcoctl

This folder contains an example which uses EMCO to deploy
Service Function Chaining (SFC).  The SFC functionality is
provided by edge clusters which use the Nodus CNI (aka OVN4NFV)
https://github.com/akraino-edge-stack/icn-nodus

EMCO provides two action controllers

1. sfc controller - which takes SFC intents 
as part of an EMCO composite application to deploy the
service functions to create an SFC.

2. sfc client controller - takes SFC client intents to
as part of EMCO composite applications which will connect
to either end of the SFC.

## Setup required

1. An Edge cluster which support Nodus SFC must be prepared.
   - As an example, see instructions found here https://github.com/akraino-edge-stack/icn-nodus/tree/master/demo/calico-nodus-secondary-sfc-setup-II
     - Note that the yaml file used to deploy the Nodus CNI ( https://github.com/akraino-edge-stack/icn-nodus/blob/master/deploy/ovn4nfv-k8s-plugin-sfc-setup-II.yaml ) contains a `ConfigMap` resource called `ovn-controller-network` which is used to define a pool of subnets which are used by Nodus to create virtual networks to stitch together the clients and functions of a network chain.
   - Ensure the EMCO `monitor` is installed in the edge cluster.
   - Apply the `NetworkAttachmentDefinition` used by EMCO to interact with Nodus in clusters (for applying `Networks`, `ProviderNetworks` and `NetworkChaining` CRs).

```
$ cat ovn-networkobj.yaml
apiVersion: k8s.cni.cncf.io/v1
kind: NetworkAttachmentDefinition
metadata:
  name: ovn-networkobj
  namespace: default
spec:
  config: '{ "cniVersion": "0.3.1", "name": "ovn4nfv-k8s-plugin", "type": "ovn4nfvk8s-cni"
    }'
```

## Description of SFC example using EMCO

This EMCO SFC example is based off of the example in the Nodus repository:
https://github.com/akraino-edge-stack/icn-nodus/tree/master/demo/calico-nodus-secondary-sfc-setup-II#demo-setup
and illustrates how SFC can be deployed and used with EMCO.

Note: that testing so far has just focused on the attachment and connection of the left and right nginx applications.
Minimal testing with provider networking attachments to the SFC when deployed with EMCO has occurred.


### SFC Composite Application

An EMCO composite application is created for deploying the SFC.  It is comprised
of the example service functions (SLB, NGFW and SDEWAN) and a set of SFC intents  which define
the following items:

- the structure of the chain - i.e. the sequence of the service functions and connecting
  networks
- the namespace and pod label client selector information used to connect client pods to and end of
  the SFC
- the provider networks that attach to an end of the SFC

When this SFC composite application is deployed by EMCO, the SFC action controller will
create the appropriate Nodus networkchaining CR and deploy it to the edge cluster(s) along
with the service functions.

### SFC Client Composite Applications

Two EMCO composite applications have been created to attach to the ends of the SFC.

Each of these client composite applications have the following:

- a  logical cloud which defines the Kubernetes namespace to be used - each client uses a different logical cloud/namespace (though not required by Nodus SFC)
- the client workload - which in this case - deploys a replicaset of pods
- an SFC client intent which identifies the SFC and which end of the chain to connect with

In addition, a second deployment intent group is defined for each composite applictions (for a total of four
client deployment intent groups).  This allows for testing of multiple Pod clients on each end of the chain.

## Running the SFC example

All scripts and files for the SFC demo are located within this directory and sub-directories.

### Setup the demo

```
./setup.sh cleanup
./setup.sh create  OR ./setup.sh create npn
emcoctl --config emco-cfg.yaml apply -f prerequisites.yaml -v values.yaml
emcoctl --config emco-cfg.yaml apply -f sfc-logical-clouds.yaml -v values.yaml
```

These steps clean up any previous demo executions and create the emco-cfg.yaml, values.yaml
and the helm charts used by the demo.

Note:  the optional `npn` argument is used to prepare the function helm charts without
provider network interface annotations on the `slb` and `sdewan` functions.  A further
improvement to the example will be to optionally apply these annotations using EMCO
provider network intents using the `ovnaction` controller.

### Testing SFC with Provider Networks Intents

1. Run `setup.sh create` to use the function helm charts with provider network annotations.
2. Install the `prerequistes.yaml` and `sfc-logical-clouds.yaml` as shown above.
3. Install the following:

```
emcoctl --config emco-cfg.yaml apply -f sfc-networks.yaml -v values.yaml
emcoctl --config emco-cfg.yaml apply -f sfc-ca.yaml -v values.yaml
```

The file `sfc-networks.yaml` creates the provider networks in the cluster for each
end of the chain.  This example can be further enhanced to create these networks using
EMCO provider network intents using the `ncm` controller instead of deploying via
a composite application and deployment intent group.

The file `sfc-ca.yaml` defines one SFC Provider Netowrk Intent and one SFC Client
Selector Intent for each end of the chain.


### Testing SFC without Provider Networks Intents

1. Run `setup.sh create npn` to use the function helm charts without provider network annotations.
2. Install the `prerequistes.yaml` and `sfc-logical-clouds.yaml` as shown above.
3. Install the following to deploy the functions and networkchaining CR:

```
emcoctl --config emco-cfg.yaml apply -f sfc-ca-npn.yaml -v values.yaml
```

The file `sfc-ca-npn.yaml` defines two different SFC Client Selector Intents for each
end of the chain, so it will be possible to install one or two clients on each end.

### Deploy the SFC Client applications

Deploy the client applications as follows:

```
emcoctl --config emco-cfg.yaml apply -f sfc-left-client.yaml -v values.yaml
emcoctl --config emco-cfg.yaml apply -f sfc-right-client.yaml -v values.yaml
```

The above left and right clients will work for both of the SFC variants above (with and without
provider networks).

For the the `sfc-ca-npn.yaml` chain, a second client can be installed on either end using the following:

```
emcoctl --config emco-cfg.yaml apply -f sfc-left-client-2.yaml -v values.yaml
emcoctl --config emco-cfg.yaml apply -f sfc-right-client-2.yaml -v values.yaml
```

Note:  these second client YAML files just define an additional deployment intent group using the
same underlying composite applications.

### Testing the SFC

After the SFC composite applications have been deployed as described above, the SFC can be tested by
running a traceroute command from a left side client pod to a right side client pod as follows:

```
kubectl -n sfc-head exec `kubectl get pods -lsfc=head -n sfc-head  --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' | head -1` -- traceroute -n -q 1 -I 172.30.22.4

traceroute to 172.30.22.4 (172.30.22.4), 30 hops max, 46 byte packets
 1  172.30.11.3  2.298 ms
 2  172.30.33.3  1.433 ms
 3  172.30.44.2  0.669 ms
 4  172.30.22.4  0.731 ms

```

Note:  the target IP address of clients on the right side may differ depending on which subnet Nodus assigns to the right end virual network.  Use a command like one of the following
to identify the IP address:

```
kubectl -n sfc-tail exec `kubectl get pods -lsfc=tail -n sfc-tail --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'` -- ifconfig sn0

If the second set of clients have been deployed:
kubectl -n sfc-tail-two exec `kubectl get pods -lsfc=tail2 -n sfc-tail-two --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'` -- ifconfig sn0
```

## How to clean up the edge cluster

Delete the emcoctl demo files in reverse order.
The `-w 3` option is useful when deleting - causing `emcoctl` to pause for several seconds after  `terminate` operations - allowing the termination operation to complete
before continuing to delete the resource.

```
emcoctl --config emco-cfg.yaml delete -f sfc-right-client.yaml -v values.yaml -w 3
emcoctl --config emco-cfg.yaml delete -f sfc-left-client.yaml -v values.yaml -w 3
emcoctl --config emco-cfg.yaml delete -f sfc-ca.yaml -v values.yaml -w 3
emcoctl --config emco-cfg.yaml delete -f sfc-networks.yaml -v values.yaml -w 3
emcoctl --config emco-cfg.yaml delete -f sfc-logical-clouds.yaml -v values.yaml -w 3
emcoctl --config emco-cfg.yaml delete -f prerequisites.yaml -v values.yaml -w 3
```


## More advanced clean up

It may be helpful to clean up and reinstall the Nodus CNI.  The following steps may help accomplish this.

Note: in the examples below, the Nodus repository has been pulling into `/home/vagrant/git`.

```
export ovnport=`ip a | grep ': ovn4nfv0' | cut -d ':' -f 2`
kubectl -n kube-system exec $(kubectl get pods -lapp=ovn-controller -n kube-system  --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}') -it -- ovs-vsctl del-port br-int $ovnport
kubectl -n kube-system exec $(kubectl get pods -lapp=ovn-control-plane -n kube-system  --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}') -it -- ovn-nbctl lsp-del $ovnport

kubectl delete -f /home/vagrant/git/icn-nodus/deploy/ovn4nfv-k8s-plugin-sfc-setup-II.yaml
# wait for above to completely terminate
kubectl delete -f /home/vagrant/git/icn-nodus/deploy/ovn-daemonset.yaml
# wait for above to completely terminate

# after above two are completely terminated, re-apply them
kubectl apply -f /home/vagrant/git/icn-nodus/deploy/ovn-daemonset.yaml
# wait for ovn to come up
kubectl apply -f /home/vagrant/git/icn-nodus/deploy/ovn4nfv-k8s-plugin-sfc-setup-II.yaml
```

## Debugging

The `collect-info.sh` script can be used to collect logs and information about the SFC demo.

The SFC demo can also be deployed manually (not using EMCO) on the target edge cluster using the instructions
here https://github.com/akraino-edge-stack/icn-nodus/tree/master/demo/calico-nodus-secondary-sfc-setup-II
if needed to verify the edge cluster has been set up correctly.  A copy of the demo SFC resource yaml files
are included in the kud/sfc/manualSFC directory.
