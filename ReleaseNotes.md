# Release Notes
This document provides high level features, fixes, and known issues and limitations of the Edge Multi-Cluster Orchestrator (EMCO) project.

- [Release Notes](#release-notes)
	- [Compatibility](#compatibility)
	- [EMCO 22.09](#emco-2209)
		- [Changelog](#changelog)
		- [Known issues](#known-issues)
	- [EMCO 22.06](#emco-2206)
		- [Changelog](#changelog-1)
		- [Known issues](#known-issues-1)
	- [EMCO 22.03.1](#emco-22031)
		- [Changelog](#changelog-2)
		- [Known issues](#known-issues-2)
	- [EMCO 22.03](#emco-2203)
		- [Changelog](#changelog-3)
		- [Known issues](#known-issues-3)
	- [EMCO 21.12](#emco-2112)
		- [Changelog](#changelog-4)
		- [Known issues](#known-issues-4)
	- [LFN Seed Code](#lfn-seed-code)
		- [Changelog](#changelog-5)
		- [Known issues](#known-issues-5)
	- [EMCO 21.03.05](#emco-210305)
		- [Changelog](#changelog-6)
		- [Known issues](#known-issues-6)
	- [EMCO 21.03](#emco-2103)
		- [Changelog](#changelog-7)
		- [Known issues](#known-issues-7)
	- [EMCO 20.12](#emco-2012)
		- [Changelog](#changelog-8)
		- [Known issues](#known-issues-8)

## Compatibility

The following table outlines EMCO's compatibility with other software projects, such as Kubernetes. It represents the versions that the team has decided to support and test, or use as part of the build process. Other versions not mentioned may also be compatible, but the project makes no promises or guarantees for such.

| EMCO         | Kubernetes (EMCO)  | Kubernetes (edge)  | Helm    | Go (compile) | Alpine (containerize)
| ------------ | ------------------ | ------------------ | ------- | ------------ | ---------------------
| **22.09**    | 1.18.x - 1.24.x    | 1.21.x - 1.24.x    | 3.8.2   | 1.17.x       | 3.12
| **22.06**    | 1.18.x - 1.23.x    | 1.21.x - 1.23.x    | 3.8.2   | 1.17.x       | 3.12
| **22.03.1**  | 1.18.x - 1.23.x    | 1.21.x - 1.23.x    | 3.8.2   | 1.17.x       | 3.12
| **22.03**    | 1.18.x - 1.23.x    | 1.21.x - 1.23.x    | 3.5.2   | 1.17.x       | 3.12
| **21.12**    | 1.18.x - 1.19.x    | 1.18.x - 1.19.x    | 3.5.2   | 1.14.x       | 3.12

Kubernetes (EMCO) represents the Kubernetes versions where the EMCO services themselves can run (with the exception of Monitor, which runs on the edge clusters).

Kubernetes (edge) represents the Kubernetes versions that the edge clusters orchestrated by EMCO must be running, for the successful deployment of Logical Clouds and Composite Apps, as well as running the Monitor service.


------------------------------

## EMCO 22.09

**Released**: 2022-09-30

### Changelog
- Expand Google Anthos GitOps support to Standard and Privileged Logical Clouds (i.e. we can deploy apps over namespaces and use RBAC).
- Add support for tracing and basic metrics to the clm, dcm, orchestrator and rsync services. Details may be found in [EMCO Observability](docs/developer/observability.md).
- Add a new error package to enforce centralized error handling across all the services in emco.
- Proof of concepts for issuing workload certificates using intel SGX. The ca-cert controller supports the intermediate ca-cert enrollment and distribution with Intel SGX capabilities.
- EMCO now comes with a local Git server, powered by Gitea. Scripts for installation and setup added.
- Core Git support added for GitOps. Through these new interfaces, any git-based repository can be accessed (not just GitHub).
- Add Intent APIs for the deploy worker in the Temporal Action Controller (tac).
- Various bugfixes, technical debt, code quality, documentation and other improvements are included in this release.
- All changes merged as part of 22.09 can be seen on [GitLab](https://gitlab.com/project-emco/core/emco-base/-/merge_requests?scope=all&state=merged&milestone_title=22.09).

### Known issues
- Open Issues confirmed as affecting EMCO 22.09 can be found at [EMCO Issues](https://gitlab.com/project-emco/core/emco-base/-/issues?sort=created_date&state=opened&label_name[]=affects:22.09).
- Other Open Issues with the label "Bug" can also be found at [EMCO Issues](https://gitlab.com/project-emco/core/emco-base/-/issues?sort=created_date&state=opened&label_name[]=Bug) although those may not be accurate as not all of them will have been triaged.
- The status query with `status=ready` parameter (i.e. show status of resources in the edge clusters), will show resources that are not handled by `monitor` as `NotPresent`. See [#149](https://gitlab.com/project-emco/core/emco-base/-/issues/149).
- The EMCO Monitor isn't currently able to watch resources outside of its own namespace, as such status querying of a standard/privileged Logical Cloud will report certain resources as not ready even when they are ready. See [#159](https://gitlab.com/project-emco/core/emco-base/-/issues/159).
- EMCO supports Referential Integrity, although in some cases it's still possible to execute operations that shouldn't be allowed after a prior failure. As such, if using `emcoctl`, we recommend using the `-s` flag in every command so that the client will stop executing after the first non-successful API call.

------------------------------

## EMCO 22.06

**Released**: 2022-06-30

### Changelog
- Adds the ca-certs controller which allows for the distribution of Intermediate CA Certificates with common root CA for the Istio service mesh in a set of clusters.  The default Istio CA can be updated and/or multiple CAs can be distributed for different namespaces (e.g. Logical Clouds).
- Add support for Google Anthos as one of the GitOps-based cluster backend for EMCO (Administrative Logical Clouds).
- Logical Cloud support added for Fluxv2 and Azure Arc based clusters.
- Orchestrator interface for action controllers extended to include post install, post update, pre terminate and post terminate grpc calls.
- Adds Temporal Action Controller which allows lifecycle management hooks allowing execution of temporal workflows before or after install/update/termination.
- Adds SDEWAN sub controller to DTC.
- Add support for Kubernetes strategic merge patch in generic action controller (GAC).
- Access point API is enabled for fine grain authorization.
- Adds support for external access to cluster microservice using user certs and keys in Istio sub controller (ITS).
- Adds support for Azure-Arc enabled Kubernetes cluster with FluxV2.
- Various bugfixes, technical debt, code quality, documentation and other improvements are included in this release.
- All changes merged as part of 22.06 can be seen on [GitLab](https://gitlab.com/project-emco/core/emco-base/-/merge_requests?scope=all&state=merged&milestone_title=22.06).

### Known issues
- Open Issues confirmed as affecting EMCO 22.06 can be found at [EMCO Issues](https://gitlab.com/project-emco/core/emco-base/-/issues?sort=created_date&state=opened&label_name[]=affects:22.06).
- Other Open Issues with the label "Bug" can also be found at [EMCO Issues](https://gitlab.com/project-emco/core/emco-base/-/issues?sort=created_date&state=opened&label_name[]=Bug) although those may not be accurate as not all of them will have been triaged.
- Google Anthos can be used with Administrative Logical Clouds, but not yet for Standard/Privileged Clouds.
- The status query with `status=ready` parameter (i.e. show status of resources in the edge clusters), will show resources that are not handled by `monitor` as `NotPresent`. See [#149](https://gitlab.com/project-emco/core/emco-base/-/issues/149).
- The EMCO Monitor isn't currently able to watch resources outside of its own namespace, as such status querying of a standard/privileged Logical Cloud will report certain resources as not ready even when they are ready. See [#159](https://gitlab.com/project-emco/core/emco-base/-/issues/159).
- EMCO now supports Referential Integrity, although in some cases it's still possible to execute operations that shouldn't be allowed after a prior failure. As such, if using `emcoctl`, we recommend using the `-s` flag in every command so that the client will stop executing after the first non-successful API call.

------------------------------

## EMCO 22.03.1

**Released**: 2022-05-19

### Changelog
- Updated all main EMCO Helm charts (emcoBase/* and monitor/) to apiVersion: v2
- Updated EMCO Helm charts and the AIO Helm Package to correctly support parameterization of the EMCO version (container image tag) as well as the GitLab Container Registry.
- Packaged and released the EMCO Helm charts publicly on the GitLab Package Registry
- Updated documentation to reflect the new method of installing EMCO (via the public Helm charts)
- Multiple fixes to the Helm charts which improve Kubernetes compatibility, including with OpenShift
- Misc fixes and improvements in/to the Helm charts

### Known issues
- All known issues of 22.03 also apply to 22.03.1
- Refer to [README.md](README.md), [BUILD.md](docs/BUILD.md) and [Tutorial_Helm.md](Tutorial_Helm.md) for additional issues related to deploying EMCO with Helm.

------------------------------

## EMCO 22.03

**Released**: 2022-03-31

### Changelog
- EMCO has been extended with Temporal workflow support, with the addition of the EMCO Workflow Manager, `workflowmgr`, with APIs to manage Temporal workflow intents. An example migrate workflow has been added to the [EMCO ecosystem temporal-migrate-workflow](https://gitlab.com/project-emco/ecosystem/temporal-migrate-workflow) repository.
- GitOps support has been enhanced with Flux v2 and Azure Arc plugins. Status handling support has been added for GitOps-based clusters. The `monitor` has been updated to support GitOps-based clusters.
- The EMCO Distributed Traffic Controller, `dtc`, now includes additional APIs for Client Access control to manage client authorizations.
- The EMCO Monitor, `monitor`, has been enhanced to be capable of supporting an extensible list of Kubernetes resources.
- The EMCO Generic Action Controller, `gac`, now has enhanced ConfigMap and Secret support to handle more use cases.
- The EMCO Distributed Cloud Manager, `dcm`, has a number of enhancements:
  - Logical Clouds may be updated after instantiation via the `/update` operation, reflecting updated attributes and resources such as additional Cluster References
  - Support for Logical Cloud status queries is enhanced to support all relevant status query parameters
  - Support for Logical Cloud status notification is enhanced to support relevant status notification parameters
- Status querying has been enhanced:
  - Query response attributes have been made consistent with status notification framework. As a result some query parameters and attributes are deprecated. See [Status changes](docs/design/Resource_Lifecycle_and_Status.md#summary-of-changes-due-to-deprecated-type-parameter)
  - The status query shows accurate results after the Deployement Intent Group has had `update` and `rollback` operations performed.
  - The status notification framework has had internal improvements made to minimize the number of status queries required and eliminate sending of duplicate notifications.
- EMCO has been updated to support deploying to Kubernetes v1.23 clusters. **Important Note**: Edge clusters must be running at least Kubernetes v1.21 in order to support EMCO capabilities, like standard/privileged Logical Clouds.
- A critical issue in the EMCO Distributed Cloud Manager has been resolved, where the wrong private key was used to generate a Logical Cloud. As such, existng EMCO users using standard/privileged Logical Clouds should migrate to 22.03 as soon as possible.
- Various bugfixes, technical debt, code quality, documentation and other improvements are included in this release.
- All changes merged as part of 22.03 can be seen on [GitLab](https://gitlab.com/project-emco/core/emco-base/-/merge_requests?scope=all&state=merged&milestone_title=22.03).

### Known issues
- Open Issues confirmed as affecting EMCO 22.03 can be found at [EMCO Issues](https://gitlab.com/project-emco/core/emco-base/-/issues?sort=created_date&state=opened&label_name[]=Affects:22.03).
- Other Open Issues with the label "Bug" can also be found at [EMCO Issues](https://gitlab.com/project-emco/core/emco-base/-/issues?sort=created_date&state=opened&label_name[]=Bug) although those may not be accurate as not all of them will have been triaged.
- The status query with `status=ready` parameter (i.e. show status of resources in the edge clusters), will show resources that are not handled by `monitor` as `NotPresent`. See [#149](https://gitlab.com/project-emco/core/emco-base/-/issues/149).
- The EMCO Monitor isn't currently able to watch resources outside of its own namespace, as such status querying of a standard/privileged Logical Cloud will report certain resources as not ready even when they are ready. See [#159](https://gitlab.com/project-emco/core/emco-base/-/issues/159).
- EMCO now supports Referential Integrity, although in some cases it's still possible to execute operations that shouldn't be allowed after a prior failure. As such, if using `emcoctl`, we recommend using the `-s` flag in every command so that the client will stop executing after the first non-successful API call. See [#98](https://gitlab.com/project-emco/core/emco-base/-/issues/98).

------------------------------

## EMCO 21.12

**Released**: 2021-12-31

### Changelog
- EMCO's Multi Mesh Istio DTC Sub-Controller. Complex applications would have front-end microservice and backend microservices.  Backend microservices may be distributed across multiple clusters.  This feature enables connectivity among the microservices in different clusters by automating the ISTIO configuration. Also, it automates the configuration of the ingress proxy to expose the frontend microservice to external users and route the traffic from the external users to the frontend microservice. As a test case, we will use EMCO to deploy Google's Online Boutique which consists of a 10-tier microservices application. This application is a web-based e-commerce app where users can browse items, add them to the cart, and purchase them. Google uses this application to demonstrate the use of technologies like Kubernetes/GKE and Istio. The feature will ensure that EMCO can deploy these microservices properly and can set up a communication channel between these microservices seamlessly through integration with Istio.
- Adds framework to support GitOps. Rsync has been modified to support multiple plugins based on how the edge clusters are managed. Cluster API's are modified to take authentication information for Git, AzureArc, GoogleAnthos etc. In the current release only plugin that is supported is Kubernetes and other plugins will be added in subsequent releases.
- Adds support for inter-dependency between applications in a composite application. A new API was added for the app to specify its dependencies. This API also allows "optional" specification of delay after the dependent app is "Ready"/"Deployed". An app is deployed after its dependent apps are properly deployed or are in a "ready" state as specified in the API.
- Adds support for Helm charts that include Helm Hooks. Helm provides a hook mechanism to allow the users to specify additional action/logic at certain points in an application's life cycle. The Helm hooks that are suuported in this release are pre-install, post-install, pre-delete and post-delete Helm Hooks.
- Enhances EMCO SFC feature to align with latest Nodus CNI SFC enhancements - e.g. dynamic virtual networks, etc.  See [SFC Overview](docs/design/sfc-overview.md)
- Referential Integrity. Add support for schema registration for new controllers. Referential integrity for EMCO data resource means the following rules must be satisfied:
  1. When a data resource is created/ updated, any resources referenced by this data resource must already exist.
  2. The data resource can't be deleted when referenced by another resource.
EMCO framework allows the dynamic addition of new controllers to the EMCO control plane (for example, you can add a new GPU capacity-based placement controller). EMCO will support referential integrity associated with any new controller. EMCO will automatically build the cross-reference relationship associated with the new data resources created by the new controller and enforce the referential integrity associated with the new data resources.
See [ Referential Integrity ](docs/developer/ReferentialIntegrity.md)
- Added support for Logical Cloud labels, translating into K8s namespace labels on the instantiated clusters.
- Implemented Status API endpoint for Logical Clouds, to check instantiation/termination progress and result.
- Implemented Stop API endpoint for Logical Clouds, to allow users to stop a pending/stuck instantiation/termination and allow for graceful removal of resources.
- General API improvements including the inclusion of JSON schema validation for DCM, bringing it up to par with the other main microservices.
- General improvements and bugfixes.
- (experimental) Introduces a status notification framework with support for orchestrator, dcm and ncm.  See [Status Notification](docs/design/Status_Notification.md)

### Known issues
Known issues were only captured in GitLab as part of this release.

------------------------------

## LFN Seed Code

**Released**: 2021-09-28

### Changelog
- The Seed Code is the first launch of EMCO in the Linux Foundation Networking umbrella, and it's an improved version of [21.03.05](#210305)
- Updates to the code required to move repository from [OpenNess] (https://github.com/smart-edge-open/EMCO) to project-emco GitLab.
- Introduced HPA (Hardware Platform Aware) placement and action controllers.
- The EMCO APIs have been modified to use camel case naming of attributes.  For example, the an attribute named 'cluster-provider' changes to 'clusterProvider'.  Also, in support of referential integrity, attribute names have been renamed in some cases to be consistent across all usages.  For example, 'appname', 'app-name', 'appName' all become 'app'.  See [JSON tag changes](JsonTagChanges.md) for a list of attributes names that have changed from earlier releases of EMCO.
- The database referential integrity feature has been added to identify referenced resources when an EMCO resource is created.  On creation, the parent resource is verified to exist and a list of other references are saved in the resource.  On delete, the EMCO database interface will prevent deletion of resources that are referenced by other resources.  See [Referential Integrity](src/orchestrator/pkg/infra/db/ReferentialIntegrity.md) for more information.
- Fix for issue DIG status API does not return complete info about clusters.
- Adds Get Arrays of Cluster with Labels query.
- Misc. document and small bug fixes.

### Known issues
- EMCO API's: Put API's are missing in some cases.
- EMCO API's: Inconsistency in error codes for EMCO API's can exist between the API doc and the actual code returned.
- General: Unit test cases are missing for some EMCO modules/controllers and test case coverage is low in some cases.
- Rsync not cleaning up state for AppContext after all processing is completed. This includes the kubeconfig files read from database and in memory data structures.
- The status of the application on termination is not tracked by rsync correctly.
- Monitor: Monitor currently supports the resources of the following type only:  corev1.Pod, corev1.Service, v1.Job, corev1.ConfigMap, corev1.Secret, appsv1.Deployment, appsv1.DaemonSet, appsv1.StatefulSet, v1beta1.CSR, v1beta1.Ingress
- Monitor Bug: Monitor is not tracking the termination status correctly.
- Emcoctl: Supports http only currently.
- Emcoctl: Add controller URL's to config file. Currently those are hardcoded in the code.
- General: EMCO is currently not able to recover when the external controllers crash or on lose of connectivity. Use grpc keepalives and timeouts will help in detecting the loss of connectivity.
- Monitor: Large applications cause " etcdserver: request is too large" error and monitor is not able to provide status of those applications.
- Rsync,Monitor: The update of cluster status (via 'monitor' and 'rsync') appears to stop working some times.  This has been observed when it was observed that standard logical clouds were not getting fully created (because the status info from the cluster was not updated).  Similar sightings in other scenarios have been reported.  The reproduction sequence is not known.
- SFC: Service Function Chaining currently expects the labels identifying the functions in a network chain to be of the form:  "app=<app name>" where the value matches an "app" name in the composite application.
- SFC: Deploying composite apps which will attach to service function chains using an SFC client intent require a namespace label match.  The labeling of the namespace in the target edge cluster(s) needs to be provided by manual or other means.  The plan is to enhance logical cloud creation to label namespaces.
- SFC: Service Function Chaining intent currently requires both a provider network intent and client selector intent for each end of the chain.  Only one of each of these intents is used (in the event more than one is created).
- General: There is no project-scoped control of privileges, even though DCM can deploy logical clouds of different privilege levels
- General: EMCO currently assumes that the namespace to access each cluster before the creation of any logical cloud is "default". However, there should a way to specify this.
- General: There is no way to modify an instantiated logical cloud (in the sense of augmenting or shrinking it, or updated quotas) - it has to be terminated and then re-instantiated.
- DCM: no status query support yet
- DCM: no JSON validation
- ovnaction:  Assumes that edge clusters will have a network-attachment-definition installed that matches: https://github.com/onap/multicloud-k8s/blob/master/kud/deployment_infra/helm/ovn4nfv-network/templates/ovnnetwork.yaml

------------------------------

## EMCO 21.03.05

**Released**: 2021-07-06

### Changelog
- Support for Application Update added in this release
- New APIs for Migrate, Update and Rollback added
- Supports changes in Helm Chart (Uploading a new Helm chart) using migrate API
- Supports adding and deleting of applications in composite application using migrate API
- Addition/ Deletion of cluster (Generic Placement Intent) based on change of labels, adding/deleting clusters from label
- Any changes to deployed resources (like replica count, labels, port changes etc)
- Rollback to any previous revision
- Application Status support for updates

- Rsync code refactored to support updates, restarts and ability to handle multiple actions simultaneously
- Status queries of Deployment Intent Groups and Cluster Networks have a few changes to align with the `rsync` and application update changes.
- The AppContext instance ID used to track AppContext status is indicated in the top level `states` object by the attribuate `statusctxid`.
- A `readystatus` attribute with values `Unknown`, `Available` or `Retrying` is added at the cluster object level.  Resources will no longer have a `Retrying` status.
- Resource level status simply shows the last action made by `rsync` on the resource.
- A sub-controller framework for DTC was added. New controllers for network policy(nps) and service discovery(sds) as sub-controllers were added to DTC.
- Logical cloud functionality has been enhanced as follows:
- There are three types of logical clouds `admin` (aka level 0), `standard` (aka level 1) and `privileged`.
- Privileged clouds allow multiple user permissions for multiple namespaces to be defined.
- The user permission API has become a distinct API to allow multiple permissions to be specified for a logical cloud.
- A variation of the vFW example has been added that does not require virtual machines in containters (i.e. virtlet).
- Service Function Chaining (SFC) is introduced as a preview feature to demonstrate integration with Nodus CNI network chaining.
- Additional unit test coverage in many packages has been added.

### Known issues
- A delay for deploying SFC CRs is explicitly performed.  This will be replaced with a more generic app / resource dependency mechanism.

------------------------------

## EMCO 21.03

**Released**: 2021-03-31

### Changelog
- Support for Helm v3 charts in composite applications.
- Service Discovery for Deployment Intent Groups. See [Service Discovery Design](docs/developer/service-discovery-design.md).
- `Put` support added to the `emcoctl` tool.
- Simple EMCO deployment Helm charts have been replaced with fuller function Helm charts with sub-charts per EMCO microservice.
- The Cluster Manager ( `clm` ) has been extended to support the invocation of registered plugin controllers when clusters are created or deleted.
- Ability for `rsync` microservice to read (get) Kubernetes resources has been added.
- Additional query parameters added to the Deployment Intent Group status query to allow for querying the list of apps, the clusters by app and resources by app.  See the status query section of [Resource Lifecycle](docs/user/Resource_Lifecycle.md).
- Emcoctl get with token has been fixed.
- Fixes in many microservices to align the data, REST API return codes with the EMCO OpenAPI documentation.
- REST PUT support added for many of the EMCO APIs.
- Additional unit test coverage in many packages has been added.
- Format of the cluster network intent status query response has been simplified to remove inapplicable and redundant `apps` and `clusters` lists.

### Known issues
- If the `monitor` pod is restarted on an edge cluster, the `rsync` connection will fail because it continues to listen on the previous (now removed) connection.
- Username / password authentication is enabled by default for EMCO mongo and etcd services.  If persistence is also enabled, then the same passwords should be used across install cycles.
- REST PUT (update) is not yet supported for `Cluster` resources  and `Deployment Intent Group` resources or sub-resources (i.e. intents) managed by the `orchestrator` microservice.
- A REST GET of a composite application app or app profile without specifying an appropriate Accept header causes the `orchestrator` microservice to panic.
- REST GETs of various intent resources of the Traffic Controller microservice `dtc` return incorrect HTTP return codes (something other than 404) when the parent resources in the URI do not exist.

------------------------------

## EMCO 20.12

**Released**: 2020-12-18

### Changelog
- This is the first release of the Edge Multi-Cluster Orchestrator (EMCO) project.  EMCO supports the automated deployment of  geo-distributed applications to multiple clusters via an intent driven API.
-   EMCO is composed of a number of central microservices:
-   Cluster Manager (clm) : onboard clusters into EMCO
-   Network Configuration Manager (ncm) : define and apply provider and virtual network intents to clusters which required additional network interfaces for workloads, such as Virtual Network Functions.  Support for OVN4NVF networks is present.
-   Distributed Cloud Manager (dcm) : define and instantiate logical clouds which provide a common namespace across a set of clusters to which applications may be deployed
-   Distributed Application Scheduler (orchestrator) : supports creation of composite applications via onboarding of Helm charts and customization and automation of deployment via support for placement and action controllers.
-   OVN Action Controller (ovnaction) : action controller which supports creation of network interface intents which automates the addition of OVN4NFV network interfaces connected to provider or vitual networks  to specified applications during deployment.
-   Traffic Controller (dtc) : action controller which supports creation of network policy intents which will deploy network policy resources to the specified clusters of the application.
-   Generic Action Controller (gac) : action controller which supports creation of intents which allow for the creation of additional Kubernetes resources for some or all of the clusters where an application is deployed.  Also it supports intents to modify Kubernetes objects which are already part of the application.
-   Resource Synchronizer (rsync) : handles instantiation, termination and status collection of the resources prepared by the other EMCO microservices to the remote clusters.
-   EMCO provides a microservice for the remote clusters:
    -   Monitor (monitor): collects and aggregates status of supported Kubernetes resources that have been deployed by EMCO.  EMCO rsync watches for updates and collects the status information.
-   EMCO provides a CLI tool (emcoctl) which may be used to interact with the EMCO REST APIs.
-   Authorization and Authentication may be provided for EMCO by utilizing Istio. See [Emco Integrity Access Management](docs/user/Emco_Integrity_Access_Management.md) for more details.

### Known issues
- EMCO provides a simple Helm chart to deploy EMCO microservices under `deployments/helm/emcoCI`.   This Helm chart supports limited scoped user authentication to the EMCO Mongo and etcd databases.  The comprehensive Helm charts under `deployments/helm/emcoBase` are still a work in progress and will include the authentication and full integration with EMCO microservices in a future release.
- Many of the EMCO microservice REST APIs do not support the PUT API for providing modifications to resources after initial creation.
- The `emcoctl` command line tool does not support a `put` operation at all.
- In some cases, EMCO does not prevent deletion of API resources which are depended on by other resources.  For example, a Cluster resource might be deleted while a Deployment Intent Group is instantiated and has resources deployed to the Cluster.  Until this issue is addressed in the next release, the best method is to ensure that resources are deleted in the reverse order from their creation.
- EMCO does not provide for encryption-at-rest for the database storage of the Mongo and etcd databases. EMCO plans to provide support for encryption of critical database resources in an upcoming release.
- The example virtual firewall composite application needs to be deployed to a Kubernetes cluster which has Multus, OVN4K8S CNI and virtlet support installed.  Refer to [KUD](https://github.com/onap/multicloud-k8s/tree/master/kud) for an example cluster that which supports the requirement needed by the virtual firewall example.
- The monitor microservice is only able to monitor the status of a limited set of Kubernetes resource Kinds:  pod, service, configmap, deployment, secret, deamonset, ingress, jobs, statefulset, csrstatus
- Emcoctl get with token doesn't work. That is because of a bug in the code. Solution to the issue is to remove line 25 from the EMCO/src/emcoctl/cmd/get.go and rebuild emcoctl code.
