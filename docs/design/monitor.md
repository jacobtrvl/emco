# EMCO Monitor

Monitor is a CRD controller that is expected to run on the Kubernetes clusters managed by EMCO and it helps to track the resources deployed by EMCO. Rsync is an EMCO controller responsible for interfacing with the Kubernetes clusters and all communication is done with Monitor is through Rsync.

## Operation Flow for Monitor

Monitor watches status changes to any of the resources deployed on Kubernetes clusters by EMCO.

When the Monitor starts running on a Kubernetes cluster it creates a list of Kubernetes resources to watch as explained below. During normal operation Monitor filters the watched resources based on the label *emco/deployment-id*. Rsync applies this label to each of the Kubernetes resource that is deployed on a cluster based on deployment intent group and corresponding application context id and the name of the application.

Example label added to each resource deployed by EMCO.

```
    emco/deployment-id: 6643168398470847081-collectd
```

In this example 6643168398470847081 is the appContext Id and collectd is the name of the application.

Rsync also creates a Monitor CR aka *ResourceBundleState* CR, corresponding to each application deployed by Rsync and applies that to the Kubernetes cluster where the application is deployed. This CR is used to capture status of the application resources. This CR also has the same label as the corresponding resources.

Once *ResourceBundleState* CR is applied to the cluster, Monitor starts collecting status of the resources with the corresponding label in the status section of the CR.

Example `ResourceBundleState` CR as applied by Rsync to a Kubernetes cluster. Notice the label 6643168398470847081-collectd.

```

apiVersion: k8splugin.io/v1alpha1
kind: ResourceBundleState
metadata:
  annotations:
  labels:
    emco/deployment-id: 6643168398470847081-collectd
  name: 6643168398470847081-collectd
  namespace: default
spec:
  selector:
    matchLabels:
      emco/deployment-id: 6643168398470847081-collectd

```

Example `ResourceBundleState` CR with status field updated by Monitor based on the application resources applied on the cluster:

```
apiVersion: k8splugin.io/v1alpha1
kind: ResourceBundleState
metadata:
  annotations:
  labels:
    emco/deployment-id: 6643168398470847081-collectd
  name: 6643168398470847081-collectd
  namespace: default
spec:
  selector:
    matchLabels:
      emco/deployment-id: 6643168398470847081-collectd
status:
  configMapStatuses:
  - apiVersion: v1
    kind: ConfigMap
    metadata:
      annotations:
        kubectl.kubernetes.io/last-applied-configuration: ""
      labels:
        app: collectd
        chart: collectd-0.2.0
        emco/deployment-id: 6643168398470847081-collectd
        release: r1
      name: r1-collectd-config
      namespace: default
  daemonSetStatuses:
  - apiVersion: apps/v1
    kind: DaemonSet
    metadata:
      annotations:
        checksum/config: 625466d49cc850e21080be0faf35d5084b8c63b86f1e3a69db9a0888c014037c
        deprecated.daemonset.template.generation: "1"
      labels:
        app: collectd
        chart: collectd-0.2.0
        emco/deployment-id: 6643168398470847081-collectd
        release: r1
      name: r1-collectd
      namespace: default
    spec:
      revisionHistoryLimit: 10
      selector:
        matchLabels:
          app: collectd
          collector: collectd
          release: r1

```

### Types of resources watched by Monitor

Monitor can be configured to watch any of the Kubernetes resources. By default, Monitor watches some common Kubernetes resources as described below:

The types of resources that are watched by the Monitor fall into 3 categories:

* ResourceBundleState CR

Monitor watches for creation and deletion of any CR of this type and starts collecting status of the resources with the corresponding label in the status section of the CR. See the previous section for description of ResourceBundleState CR and example.

* Common Kubernetes resources

Some common Kubernetes resources are handled explicitly by Monitor and the status of these resources are stored in the CR as Kubernetes resource.

  #### Kubernetes resources handled explicitly  by Monitor

  - Deployments
  - Daemonsets
  - StatefulSet
  - Service
  - Job
  - Pods
  - ConfigMap
  - CSR

The example below is showing how a configmap resource is stored in the ResourceBundle CR.

```
apiVersion: k8splugin.io/v1alpha1
kind: ResourceBundleState
metadata:
  annotations:
  labels:
    emco/deployment-id: 6643168398470847081-collectd
  name: 6643168398470847081-collectd
  namespace: default
spec:
  selector:
    matchLabels:
      emco/deployment-id: 6643168398470847081-collectd
status:
  configMapStatuses:
  - apiVersion: v1
    kind: ConfigMap
    metadata:
      annotations:
        kubectl.kubernetes.io/last-applied-configuration: ""
      labels:
        app: collectd
        chart: collectd-0.2.0
        emco/deployment-id: 6643168398470847081-collectd
        release: r1
      name: r1-collectd-config
      namespace: default
```

* All other Kubernetes resources

The Monitor can monitor any type of Kubernetes resource. These kinds of resources are captured in the ResourceBundleState CR in the format as shown in the example below. These resources are stored under *resourceStatuses* field in the CR status. The Kubernetes resource is converted to a byte array and stored in the *res* field as shown below.

```
apiVersion: k8splugin.io/v1alpha1
kind: ResourceBundleState
metadata:
  annotations:
  labels:
    emco/deployment-id: 6643168398470847081-collectd
  name: 6643168398470847081-collectd
  namespace: default
spec:
  selector:
    matchLabels:
      emco/deployment-id: 6643168398470847081-collectd
status:
  resourceStatuses:
  - group: rbac.authorization.k8s.io
    kind: Role
    name: pod-reader
    namespace: default
    res: eyJhcGlWZXJzaW9uIjoicmJhYy5hdXRob3JpemF0aW9uLms4cy5pby92MSIsImtpbmQiOiJSb2xlIiwibWV0YWRhdGEiOnsiYW5ub3RhdGlvbnMiOns
ia3ViZWN0bC5rdWJlcm5ldGVzLmlvL2xhc3QtYXBwbGllZC1jb25maWd1cmF0aW9uIjoiIn0sImNyZWF0aW9uVGltZXN0YW1wIjoiMjAyMi0wMi0yNFQwMToxMTo
1M1oiLCJsYWJlbHMiOnsiZW1jby9kZXBsb3ltZW50LWlkIjoiNjY0MzE2ODM5ODQ3MDg0NzA4MS1jb2xsZWN0ZCJ9LCJtYW5hZ2VkRmllbGRzIjpbXSwibmFtZSI
6InBvZC1yZWFkZXIiLCJuYW1lc3BhY2UiOiJkZWZhdWx0IiwicmVzb3VyY2VWZXJzaW9uIjoiOTU1ODI0NyIsInVpZCI6ImIzNmU0ZjcwLTM3NjYtNDk4MC1iNWJ
iLTQyN2FhMThjZTBiYiJ9LCJydWxlcyI6W3siYXBpR3JvdXBzIjpbIiJdLCJyZXNvdXJjZXMiOlsicG9kcyJdLCJ2ZXJicyI6WyJnZXQiLCJ3YXRjaCIsImxpc3Q
iXX1dfQ==
    version: v1

```

Note: Use `kubectl api-resources` command to find the group, version, kind and resource fields for the resource that requires monitoring.

## GitOps support

The Monitor can be configured to store the ResourceBundleState CR in a git location. This details about the git location are provides by creating secret as shown below.

```
apiVersion: v1
kind: Secret
metadata:
  name: {{ .Release.Name }}-git-monitor
  namespace: {{ .Release.Namespace }}
type: Opaque
data:
  username: abc
  token: XXXX
  repo: test
  clustername: provider1+cluster1

```

The secret can be enabled during helm install like below:

```
 helm  install  --set git.token=$GITHUB_TOKEN --set git.repo=SREPO --set git.username=$OWNER --set git.clustername="provider1flux+cluster2" --set git.enabled=true  -n emco monitor .

```
Refer to the GitOps document for more details on GitOps support in EMCO: https://gitlab.com/project-emco/core/emco-base/-/blob/main/docs/design/gitops_support.md


