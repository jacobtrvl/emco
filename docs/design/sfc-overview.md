```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2019-2021 Intel Corporation
```

# Overview of Service Function Chaining (SFC) - aka Network Chains

EMCO provides two action controllers for handling the input and application of SFC  intents.

- The `sfc` action controller handles the input of intents which will be used to generate an SFC CR which will be applied to clusters along with the applications that comprise the functions in the SFC.
- `sfcclient` action controller handles the intents which will be used to attach client applications to the left or right ends of the SFC.

At this time, SFC is supported on clusters where the Nodus `ovn4nfv` CNI is deployed.

See: https://github.com/akraino-edge-stack/icn-nodus

# SFC Controller

The `sfc` action controller is responsible for taking in via API the intents for network chains.

The `sfc` action controller intents are added to a deployment intent group which is created to deploy `Network Chains` to a set of
clusters.  These intents will be used to create a `NetworkChaining` CR.  The left and right ends of a network chain are connected to one or more `Provider Networks` and/or `Pods`.

`Provider Networks` may be deployed to clusters with EMCO using `ncm` provider network intents.

Other `Deployment Intent Groups` which include `sfclclient` action controller intents that refer to a `Network Chain` may specify that `Pods` in specified application resources  be *connected* to an end of a `Network Chain` via label matching.

## Overview of SFC Intent Creation

To create a deployment intent group that defines a network chain, the following steps are required:

- A functions within the chain need to be included in the Composite Application.
- All functions within the chain need to be placed in the same cluster(s).  The networkchaining CR will only be created in clusters where all functions that are part of the chain are present.
- Define an SFC Intent
- Define a set of SFC Link Intents - which define the sequence of functions in the chain.
- Optionally define SFC Provider Network Intents to connect a PRovider Network to either end of the chain.
- Optionally define SFC Client Selector Intents to connect Pods from other deployment intent groups to either end of the chain.
- At least one SFC Provider Network Intent or SFC Client Selector Intent needs to be provided for each end of the chain.

## SFC Intent

The SFC Intent resource is the parent intent that 'holds' the corresponding SFC Link and PRovider Network and/or Client Selector intents.

URL:

```
/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-chains
```

Body:

```
{
  "metadata": {
    "name": "sfc1",
    "description": "virtual/provider network to virtual/provider network",
    "userData1": "data 1",
    "userData2": "data 2"
  },
  "spec": {
    "chainType": "Routing",
    "namespace": "default",
  }
}
```

### SFC Link Intent

The SFC Link Intent is used to define the links of the chain.  For a given SFC Intent, the set of SFC Link Intents are used
to create the `networkChain` attribute of the Nodus networkchaining CR, which looks like the following:

```
    "networkChain": "net=vnet1,app=slb,net=dync-net1,app=vfw,net=dync-net2,app=sdewan,net=vnet2"
```

URL:

```
/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-chains/{sfc-intent}/links
```

Body:

```
{
  "metadata": {
    "name": "sfc-link1",
  },
  "spec": {
    "leftNet": "vnet1",
    "rightNet": "dync-net1",
    "app": slb
    "linkLabel": app=slb
    "workloadResource": slb
    "resourceType": Deployment
  }
}
```

In the SFC Link Intent `spec` portion, the `app` identifies the application within the EMCO composite application for the given function.  The `resourceType` and `workloadResource` identify the Kind and name of the resource within the `app` that is the function.
The `linkLabel` defines the label that the Nodus SFC functionality will expect to find on a Pod when it instantiates the network chain.  The EMCO sfc controller will ensure that the Pod template of the identified resource is labelled with `linkLabel`.

The `leftNet` and `rightNet` attributes identify labels on Nodus Network resources that stitch together links in the network chain from left to right.  Nodus expects the network labels to be of the form `net=value`.  Only the value portion is required in the EMCO SFC Link intent.  The sfc controller will prepend `net=` to each network label value when it creates the `networkChain` attribute.

To complete the set of SFC Link intents for the example `networkChain` above, the  following two SFC Link intents are also created:

```
{
  "metadata": {
    "name": "sfc-link2",
  },
  "spec": {
    "leftNet": "dync-net1",
    "rightNet": "dync-net2",
    "app": vfw
    "linkLabel": app=vfw
    "workloadResource": vfw
    "resourceType": Deployment
  }
}
```

```
{
  "metadata": {
    "name": "sfc-link3",
  },
  "spec": {
    "leftNet": "dync-net2",
    "rightNet": "vnet2",
    "app": sdewan
    "linkLabel": app=sdewan
    "workloadResource": sdewan
    "resourceType": Deployment
  }
}
```

It is not required that EMCO create the networks that make up the links in the chain.  Nodus uses these network labels to create virtual networks of the same name, if necessary.


### Network Chain Client Selectors

Defines the labels and namespaces that will match with pod template specs to attach the pod to the specified end of the SFC.

URL:

```
/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-chains/{sfc}/client-selectors
```

Body:

```
{
  "metadata": {
    "name": "client1",
    "description": "pod client of an SFC",
    "userData1": "data 1",
    "userData2": "data 2"
  },
  "spec": {
    "chainEnd": "< left | right >",
    "podSelector": {
      "matchLabels": {
	"app": "vbng"
      }
    },
    "namespaceSelector": {
      "matchLabels": {
	"app": "vbng"
      }
    }
  }
}
```

### Network Chain Provider Networks

Defines the provider networks that may be attached to a specified end of the SFC.

URL:

```
/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-chains/{sfc}/provider-networks

```

Body:

```
{
  "metadata": {
    "name": "chain-providernetwork",
    "description": "chain-providernetwork",
    "userData1": "data 1",
    "userData2": "data 2"
  },
  "spec": {
    "chainEnd": "< left | right >",
    "networkName": "provider-network-1",
    "gatewayIp": "172.30.10.3",
 	    "networkRepresentor": {
      "gatewayip": "<ipaddress>",
      "subnet": "<ipaddress>"
    }
  }
}
```

# Network Chain Client Controller

The `sfcclient` controller is responsible for taking in, via API, the intents that will match up applications to specified endpoints of a Network Chain.

When this intent is included in the set of intents of a `Deployment Intent Group` and the `sfcclient` action controller is executed (during instantiation of the deployment intent group), the identified chain will be queried and:

- the set of SFC Client Selector Intents for the specified chain end (`left` or `right`) is found
- for the first SFC Client Selector Intent whose `namespaceSelector` labels match the labels associated with the logical cloud of the current deployment intent group, the `podSelector` labels of the intent are applied to the Pod template of the resource identified in the SFC Client Intent.

## Network Chain Client Intents

URL:

```
/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/sfc-clients
```

Body:

```
{
  "metadata": {
    "name": "chain1 client1",
    "description": "chain1 client1 information",
    "userData1": "blah blah",
    "userData2": "abc xyz"
  },
  "spec": {
    "appName": "app1",
    "workloadResource": "app1Deployment",
    "resourceType": "deployment"
    "chainEnd": "<left | right>",
    "chainName": "chain1",
    "chainCompositeApp": "chain1-CA",
    "chainCompositeAppVersion": "chain1-CA-version",
    "chainDeploymentIntentGroup": "chain1-deployment-intent-group",
  }
}
```

# References

See the following links for more information about network chaining.

- https://github.com/akraino-edge-stack/icn-nodus/tree/master/demo/calico-nodus-secondary-sfc-setup-II
