# HPA examples

## Summary

Here are some examples of HPA usages in EMCO in which the apps are deployed to
remote clusters that meet the apps requirements.

There are currently two composite apps: http and virtual firewall. The http app is
composed from a client and a server app. It is used to demonstrate a couple of use
cases where allocatable and non-allocatable resources are requested in several
combinations of remote clusters.

The vFW is compsed from three apps: packetgen, firewall and sink. Packetgen
and firewall use fd.io's VPP to implement features. And VPP requires hugepages
during operations. Thus the worker nodes in the target clusters need to be
enabled with hugepages. EMCO checks if the target clusters have nodes with
hugepages labelled before deploying the apps.

## Deployment
### Preparation
Edit the setup.sh file and replace the KUBE_PATH and HOST_IP variables with the correct
the paths to the kube configs of the remote clusters and the host IP.

### Labelling nodes in the target clusters using NFD
Nodes with specific features can be labelled using NFD (Node feature discovery). Please
refer to [this guide][1] to enable NFD in the clusters. For example, a custom label of
hugepages can be applied to the nodes by adding a file with following content to:
```
$ sudo vi /etc/kubernetes/node-feature-discovery/features.d/extra-features
/hugepages
```
We can do the same with other features.

### Deployment steps
Generate the files containing basic deployment information
```
$ ./setup.sh create
```
Register remote clusters
```
$ emcoctl --config emco-cfg.yaml apply -f prerequisites.yaml -v values.yaml
```
Deploy the http app. Replace the hpa-example.yaml with other file names for other use cases (hpa-test-simple.yaml, hpa-test-advanced.yaml or hpa-test-advanced-composite.yaml)
```
$ emcoctl --config emco-cfg.yaml apply -f hpa-example.yaml -v values.yaml
```
For the vFW app, the deployment will be done in two steps
```
$ emcoctl --config emco-cfg.yaml apply -f hpa-vfw-networks.yaml -v values.yaml
$ emcoctl --config emco-cfg.yaml apply -f hpa-vfw-apps.yaml -v values.yaml
```

To demonstrate the concepts of HPA, the firewall and packetgen apps are intended to be
deployed to any of the clusters which support hugepages, while the sink app is intended
to be deployed to all clusters. In this case, if cluster1 and cluster2 both support 
hugepages are labelled correctly, firewall and packetgen apps are deployed to only one of
them. If only one cluster supports hugepages and labelled correctly, they will be deployed
to it.
To deploy vFW to all clusters, please change the intent of firewall and packetgen
apps to allOf (instead of anyOf). In this case, all clusters need to have hugepages and
be labelled correctly. Otherwise the deployment will fail.

### Uninstallation
The deployment can be removed in the reversed order

In case of the vFW app
```
$ emcoctl --config emco-cfg.yaml delete -f hpa-vfw-apps.yaml -v values.yaml
$ emcoctl --config emco-cfg.yaml delete -f hpa-vfw-networks.yaml -v values.yaml
```
For the http app
```
$ emcoctl --config emco-cfg.yaml delete -f hpa-example.yaml -v values.yaml
```
And then
```
$ emcoctl --config emco-cfg.yaml delete -f prerequisites.yaml -v values.yaml
```
Finally,
```
$ ./setup.sh cleanup
```

[1]: https://kubernetes-sigs.github.io/node-feature-discovery/stable/get-started/index.html
