// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/common"
	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/notifyclient"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/readynotify"
	yamlV2 "gopkg.in/yaml.v2"
)

const distributionApp = "cert-distribution"

// CertDistributionManager
type CertDistributionManager interface {
	Instantiate(cert, clusterProvider string) error
	Status(cert, clusterProvider string) (CertEnrollmentStatus, error)
	Terminate(cert, clusterProvider string) error
	Update(cert, clusterProvider string) error
	// Delete()
	// Get()
}

type distributionAppContext struct {
	appContext          appcontext.AppContext
	appHandle           interface{}
	caCert              certificate.Cert // CA
	clusterGroups       []ClusterGroup   // clusters those are part of the CA
	clusterProvider     string
	contextID           string
	resorder            []string
	enrollmentContextID string
	clusters            []string
	// resources       []string // name of the CR created for each clusters
	certificateRequests []certificate.CertificateRequest
	resource            distributionResource
}

type CertDistributionStatus struct {
	DeployedStatus  appcontext.StatusValue `json:"deployedStatus,omitempty,inline"`
	ReadyStatus     string                 `json:"readyStatus,omitempty,inline"`
	ReadyCounts     map[string]int         `json:"readyCounts,omitempty,inline"`
	App             string                 `json:"app,omitempty,inline"`
	ClusterProvider string                 `json:"clusterProvider,omitempty"`
	Cluster         string                 `json:"cluster,omitempty"`
	Connectivity    string                 `json:"connectivity,omitempty"`
	Resources       []Resource             `json:"resources,omitempty"`
}

type distributionResource struct {
	handle             interface{}
	cluster            string
	secret             certificate.Secret
	proxyConfig        certificate.ProxyConfig
	clusterIssuer      certificate.ClusterIssuer
	certifiacteRequest certificate.CertificateRequest
}

// CertDistributionClient
type CertDistributionClient struct {
	db DbInfo
}

// NewCertDistributionClient
func NewCertDistributionClient() *CertDistributionClient {
	return &CertDistributionClient{
		db: DbInfo{
			storeName: "resources",
			tagMeta:   "data"}}
}

func (c *CertDistributionClient) Instantiate(cert, clusterProvider string) error {
	// get the ca cert
	caCert, err := getCertificate(cert, clusterProvider)
	if err != nil {
		return err
	}

	appContext := distributionAppContext{
		caCert:          caCert,
		clusterProvider: clusterProvider}

	if err := appContext.validate(); err != nil {
		return err
	}

	// instantiate a new appContext
	ctx := common.Context{
		AppName:    distributionApp,
		ClientName: clientName}

	if err := ctx.InitAppContext(); err != nil {
		return err
	}

	appContext.appContext = ctx.AppContext
	appContext.appHandle = ctx.AppHandle
	appContext.contextID = ctx.ContextID

	// initialize the Instantiation
	if err = appContext.instantiateDistribution(); err != nil {
		return err
	}

	//  invoke Rsync to create the CR resources in the issuing cluster
	if err = appContext.callRsyncInstall(); err != nil {
		return err
	}

	key := StateKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		AppName:         distributionApp}

	if err := common.NewStateClient().UpdateState(key, state.StateEnum.Instantiated, ctx.ContextID, false); err != nil {
		return err
	}

	return nil
}

func (context *distributionAppContext) instantiateDistribution() error {
	fmt.Println("Create AddLevelValue", context.caCert.MetaData.Name, context.clusterProvider)

	meta := map[string]string{}
	meta["cert"] = context.caCert.MetaData.Name
	meta["clusterProvider"] = context.clusterProvider
	lHandle, err := context.appContext.AddLevelValue(context.appHandle, appMeta, meta)
	if err != nil {
		return err
	}

	fmt.Println("AddLevelValue Handle: ", lHandle)

	// At this point we assume we have already validated the clusters and the csr created
	// So we don;t have to loop through each clusters in the ca/cluster group
	// just create the resources
	for _, cluster := range context.clusters {
		// add handle for the issuing cluster
		handle, err := context.appContext.AddCluster(context.appHandle,
			strings.Join([]string{context.clusterProvider, cluster}, "+"))
		if err != nil {
			fmt.Println("Failed to add cluster to the context", context.caCert.MetaData.Name, context.caCert.Spec.IssuingCluster.ClusterProvider, "", context.caCert.Spec.IssuingCluster.Cluster, err)
			return err
		}
		fmt.Println("clusterHandle: ", handle)

		for _, cr := range context.certificateRequests {
			name := certificateRequestName(context.caCert.MetaData.Name, cluster, context.clusterProvider, context.enrollmentContextID)
			if cr.MetaData.Name == name {
				dr := distributionResource{
					handle:             handle,
					cluster:            cluster,
					certifiacteRequest: cr}
				context.resource = dr

				// Create the secret
				if err := context.createSecret(); err != nil {
					return err
				}

				// Create the proxyconfig
				// TODO : need to determine if this is needed for the default Istio CA)
				if err := context.createProxyConfig(); err != nil {
					return err
				}

				// Create the ClusterIssuer
				if err := context.createClusterIssuer(); err != nil {
					return err
				}

				resOrder, err := json.Marshal(map[string][]string{"resorder": context.resorder})
				if err != nil {
					return err
				}

				// add instruction under given handle and type
				iHandle, err := context.appContext.AddInstruction(handle, "resource", "order", string(resOrder))
				if err != nil {
					if er := context.appContext.DeleteCompositeApp(); er != nil {
						logutils.Warn("Failed to delete the compositeApp", logutils.Fields{
							"contextID": context.contextID,
							"Error":     err.Error()})
					}
					return err
				}

				fmt.Println("AddInstruction Handle: ", iHandle)
				break
			}
		}
	}

	return nil
}

func (context *distributionAppContext) createSecret() error {
	s := certificate.CreateSecret()
	// TODO : verify the naming
	s.MetaData.Name = secretName(context.caCert.MetaData.Name, context.resource.cluster, context.clusterProvider, context.contextID)
	// TODO: this needs to be configurable, also this namespace should be available, otherwise resource creation will fail
	s.MetaData.Namespace = "cert-manager"
	s.Data["tls.crt"] = context.resource.certifiacteRequest.Status.Certificate
	s.Data["tls.key"] = context.resource.certifiacteRequest.Status.Certificate // TODo: COnfirm which key
	value, err := yamlV2.Marshal(s)
	if err != nil {
		return err
	}

	fmt.Println(string(value))
	// add resource under app and cluster
	if err := addResource(context.appContext, context.resource.handle, s.ResourceName(), string(value)); err != nil {
		deleteCompositeApp(context.appContext) // TODo: COnfirm if we need to delete this for any single resource add
		return err
	}

	context.resource.secret = certificate.Secret{
		APIVersion: s.APIVersion,
		Kind:       s.Kind,
		MetaData:   s.MetaData,
		Data:       s.Data,
		Type:       s.Type}

	context.resorder = append(context.resorder, s.ResourceName())

	return nil

}

func (context *distributionAppContext) createProxyConfig() error {
	pc := certificate.CreateProxyConfig()
	pc.MetaData.Name = proxyConfigName(context.caCert.MetaData.Name, context.resource.cluster, context.clusterProvider, context.contextID) // TODO : verify the naming
	pc.MetaData.Namespace = "istio-system"                                                                                                 // TODO: this needs to be configurable and confirm this is default in a single tenanant use case
	pc.Spec.EnvironmentVariables["ISTIO_META_CERT_SIGNER"] = "istio-system"                                                                // TODO: this needs to be configurable and confirm this is default in a single tenanant use case

	value, err := yamlV2.Marshal(pc)
	if err != nil {
		return err
	}

	fmt.Println(string(value))
	// add resource under app and cluster
	if err := addResource(context.appContext, context.resource.handle, pc.ResourceName(), string(value)); err != nil {
		deleteCompositeApp(context.appContext) // TODO: COnfirm if we need to delete this for any single resource add
		return err
	}

	context.resorder = append(context.resorder, pc.ResourceName())

	return nil

}
func (context *distributionAppContext) createClusterIssuer() error {
	issuer := certificate.CreateClusterIssuer()
	issuer.MetaData.Name = clusterIssuerName(context.caCert.MetaData.Name, context.resource.cluster, context.clusterProvider, context.contextID) // TODO : verify the naming
	issuer.Spec.CA.SecretName = context.resource.secret.MetaData.Name

	value, err := yamlV2.Marshal(issuer)
	if err != nil {
		return err
	}

	fmt.Println(string(value))

	// add resource under app and cluster
	if err := addResource(context.appContext, context.resource.handle, issuer.ResourceName(), string(value)); err != nil {
		deleteCompositeApp(context.appContext) // TODO: COnfirm if we need to delete this for any single resource add
		return err
	}

	context.resorder = append(context.resorder, issuer.ResourceName())
	return nil
}

func (context *distributionAppContext) callRsyncInstall() error {
	// var log = func(message, contextID string, err error) {
	// 	fields := make(logutils.Fields)
	// 	fields["ContextID"] = contextID
	// 	fields["Client"] = clientName
	// 	if err != nil {
	// 		fields["Error"] = err.Error()
	// 	}
	// 	logutils.Error(message, fields)
	// }

	// invokes the rsync service
	if err := notifyclient.CallRsyncInstall(context.contextID); err != nil {
		fmt.Println("Failed to invokes the rsync service", context.contextID, err)
		return err
	}

	// subscribe to alerts
	stream, client, err := notifyclient.InvokeReadyNotify(context.contextID, clientName)
	if err != nil {
		fmt.Println("Failed to subscribe to alerts from the rsync gRPC server", context.contextID, err)
		return err
	}

	// do not wait for the resource synchronization to complete
	// handle the retrieval and verification in a go-routine
	// store the certificate generated by CR in mongo-db
	// update the state of enrollment to instantiated and return the api

	go retrieveDistributionStatus(client, stream)

	if err := stream.CloseSend(); err != nil {
		fmt.Println("Failed to close the send stream", context.contextID, err)
		return err
	}

	return nil
}

func (context *distributionAppContext) validate() error {
	key := StateKey{
		Cert:            context.caCert.MetaData.Name,
		ClusterProvider: context.clusterProvider,
		AppName:         distributionApp}

	s := common.NewStateClient()
	// check the current state of the distribution, if any
	if _, err := s.VerifyState(key, common.InstantiateEvent); err != nil {
		return err
	}

	if err := context.verifyCertEnrollmentState(); err != nil {
		return err
	}

	// Validate the clusters list in the Cacert and the total resources created
	// This is to verify any pending update is required for enrollemnt

	readyCount, err := context.validateCertEnrollmentStatus()
	if err != nil {
		return err
	}

	csrs, err := checkCertificateRequestStatus(context.enrollmentContextID)
	if err != nil {
		return err
	}

	if readyCount != len(csrs) {
		return errors.New("Enrollment is not ready")
	}

	context.certificateRequests = csrs
	if err := context.validateCertificateRequest(); err != nil {
		return err
	}

	return nil
}

func (context *distributionAppContext) verifyCertEnrollmentState() (err error) {
	// get the cert enrollemnt instantiation state
	key := StateKey{
		Cert:            context.caCert.MetaData.Name,
		ClusterProvider: context.clusterProvider,
		AppName:         enrollmentApp}

	stateInfo, err := common.NewStateClient().GetState(key)
	if err != nil {
		return err
	}

	contextID := state.GetLastContextIdFromStateInfo(stateInfo)
	if len(contextID) == 0 {
		return errors.New("Enrollment is not completed")
	}

	status, err := state.GetAppContextStatus(contextID)
	if err != nil {
		return err
	}

	if status.Status != appcontext.AppContextStatusEnum.Instantiated &&
		status.Status != appcontext.AppContextStatusEnum.Updated {
		return errors.New("Enrollment is not completed")
	}

	context.enrollmentContextID = contextID
	return nil
}

func (context *distributionAppContext) validateCertEnrollmentStatus() (readyCount int, err error) {
	//  verify the status of the enrollemnt
	certEnrollmentStatus, err := NewCertEnrollmentClient().Status(context.caCert.MetaData.Name, context.clusterProvider)
	if err != nil {
		return readyCount, err
	}

	if strings.ToLower(string(certEnrollmentStatus.DeployedStatus)) != "instantiated" {
		return readyCount, errors.New("Enrollment is not ready")
	}
	if strings.ToLower(certEnrollmentStatus.ReadyStatus) != "ready" {
		return readyCount, errors.New("Enrollment is not ready")
	}
	if strings.ToLower(certEnrollmentStatus.Connectivity) != "available" {
		return readyCount, errors.New("Enrollment is not ready")
	}

	for _, resource := range certEnrollmentStatus.Resources {
		if strings.ToLower(resource.ReadyStatus) != "ready" {
			return readyCount, errors.New("Enrollment is not ready")
		}
	}

	return certEnrollmentStatus.ReadyCounts["Ready"], nil
}

func (context *distributionAppContext) validateCertificateRequest() error {
	// get all the clusters defined under this CA
	clusterGroups, err := getAllClusterGroup(context.caCert.MetaData.Name, context.clusterProvider)
	if err != nil {
		return err
	}
	cClient := cluster.NewClusterClient()
	context.clusters = []string{}
	for _, clusterGroup := range clusterGroups {
		available := false
		// TODO: Confirm if we need to veirfy the cluster exists or not
		switch strings.ToLower(clusterGroup.Spec.Scope) {
		case "name":
			context.clusters = append(context.clusters, clusterGroup.Spec.Name)
			crName := certificateRequestName(context.caCert.MetaData.Name, clusterGroup.Spec.Name, context.clusterProvider, context.enrollmentContextID) // TODO: this needs to be a unique name, check the format
			for _, csr := range context.certificateRequests {
				if csr.MetaData.Name == crName {
					if err := certificate.ValidateCertificateRequest(csr); err != nil {
						fmt.Println("Failed to delete clusters certificate request", context.caCert.MetaData.Name, clusterGroup.Spec.Name, clusterGroup.MetaData.Name, context.clusterProvider, err.Error())
						return err
					}
					available = true
					break
				}
			}

			// TODO : verify the logic here
			if !available {
				return errors.New("certificate request is not ready for ")

			}

		case "label":
			// get clusters by label TODO: Confirm if we need to veirfy the cluster exists or not
			list, err := cClient.GetClustersWithLabel(context.clusterProvider, clusterGroup.Spec.Label)
			if err != nil {
				fmt.Println("Failed to get clusters by label", context.caCert.MetaData.Name, context.clusterProvider, clusterGroup.Spec.Label, "", err)
				return err
			}
			context.clusters = append(context.clusters, list...)
			for _, name := range list {
				crName := certificateRequestName(context.caCert.MetaData.Name, name, context.clusterProvider, context.enrollmentContextID) // TODO: this needs to be a unique name, check the format
				for _, cr := range context.certificateRequests {
					{
						if cr.MetaData.Name == crName {
							if err := certificate.ValidateCertificateRequest(cr); err != nil {
								fmt.Println("Failed to delete clusters certificate request", context.caCert.MetaData.Name, clusterGroup.Spec.Name, clusterGroup.MetaData.Name, context.clusterProvider, err.Error())
								return err
							}
							available = true
							break
						}

					}

				}
				// TODO : verify the logic here
				if !available {
					return errors.New("certificate request is not ready for ")

				}
			}

		}

	}

	context.clusterGroups = clusterGroups
	return nil
}

func retrieveDistributionStatus(client readynotify.ReadyNotifyClient, stream readynotify.ReadyNotify_AlertClient) {
	contextID := retrieveAppContext(client, stream)
	_, err := checkDistributionStatus(contextID) // check whether all certificates have been issued
	if err != nil {
		return
	}

	// At this point we will have certificateRequests created

	// TODO: Get the state and contextID from the state info
	// validate the context id
	// Why don;t we pass the variables from the top level, why we need to query context?

	// processCertificateRequests(contextID, certificateRequests)

	if _, err := client.Unsubscribe(context.Background(), &readynotify.Topic{ClientName: clientName, AppContext: contextID}); err != nil {
		logutils.Error("[ReadyNotify gRPC] Failed to unsubscribe to alerts",
			logutils.Fields{"ContextID": contextID,
				"Error": err.Error()})
	}

}

func checkDistributionStatus(contextID string) ([]distributionResource, error) {
	var distributionResources []distributionResource
	var appContext appcontext.AppContext

	// load the appContext
	_, err := appContext.LoadAppContext(contextID)

	// TODO: Confirm return err or false? Shoudl we retry in case of appContext errors
	if err != nil {
		fmt.Println("Failed to load the appContext", "", "", contextID, "", err)
		return []distributionResource{}, err
	}

	// get the app instruction for 'order'
	appsOrder, err := appContext.GetAppInstruction("order")
	if err != nil {
		fmt.Println("Failed to get the app instruction for the 'order' instruction type", "", "", contextID, "", err)
		return []distributionResource{}, err
	}

	var appList map[string][]string
	json.Unmarshal([]byte(appsOrder.(string)), &appList)

	distributionResources = []distributionResource{}

	for _, app := range appList["apporder"] {
		//  get all the clusters associated with the app
		clusters, err := appContext.GetClusterNames(app)
		if err != nil {
			fmt.Println("Failed to list clusters", app, "", contextID, "", err)
			return []distributionResource{}, err
		}

		for _, cluster := range clusters {
			// get the resources
			resources, err := appContext.GetResourceNames(app, cluster)
			if err != nil {
				fmt.Println("Failed to get the resources", app, cluster, contextID, "", err)
				return []distributionResource{}, err
			}

			fmt.Println("GetResourceNames: ", resources)

			// get the cluster handle
			cHandle, err := appContext.GetClusterHandle(app, cluster)
			if err != nil {
				fmt.Println("Failed to get the cluster handle", app, cluster, contextID, "", err)
				return []distributionResource{}, err
			}

			// get the the cluster status handle
			sHandle, err := appContext.GetLevelHandle(cHandle, "status")
			if err != nil {
				fmt.Println("Failed to get the handle of 'status'", app, cluster, contextID, "", err)
				return []distributionResource{}, err
			}

			// get the status of the cetificaterequests resource creation
			// wait for the resources to be created and available in the monitor resource bundle state
			// retry if the resources satatus are not available
			//  TODO: Confirm max retrying
			// statusReady := false
			// for !statusReady {
			// 	// get the value of 'status' handle
			val, err := appContext.GetValue(sHandle)
			if err != nil {
				fmt.Println("Failed to get the value of 'status' handle", app, cluster, contextID, "", err)
				continue
			}

			fmt.Println("Status Value: ", val.(string))

			s := struct {
				Ready            bool `json:"ready"`
				ResourceCount    int  `json:"resourceCount"`
				ResourceStatuses []struct {
					Group     string `json:"Group"`
					Version   string `json:"Version"`
					Kind      string `json:"Kind"`
					Name      string `json:"Name"`
					Namespace string `json:"Namespace"`
					Res       string `json:"Res"`
				} `json:"resourceStatuses"`
			}{}

			if err := json.Unmarshal([]byte(val.(string)), &s); err != nil {
				fmt.Println("Failed to unmarshal cluster status", app, cluster, contextID, val.(string), err)
				return []distributionResource{}, err
			}

			fmt.Println("Resource status: ", s)

			// if len(s.ResourceStatuses) == 0 {
			// 	continue
			// }

			// for each resource make sure the certificate request is created and the status is available
			// for _, resource := range resources {
			// 	for _, rStatus := range s.ResourceStatuses {
			// 		secret := certificate.CreateSecret()
			// 		pc := certificate.CreateProxyConfig()
			// 		issuer := certificate.CreateClusterIssuer()

			// 		if secret.Kind == rStatus.Kind && secret.APIVersion == rStatus.Version {
			// 			secret.MetaData.Name= rStatus.Name

			// 		}
			// 		crsName := certificateRequestResourceName(rStatus.Name, rStatus.Kind)
			// 		if crsName == resource {
			// 			if len(rStatus.Res) == 0 {
			// 				logutils.Warn(fmt.Sprintf("Cluster status does not contain the certificate details for %s", rStatus.Name),
			// 					logutils.Fields{})
			// 				break
			// 			}

			// 			data, err := base64.StdEncoding.DecodeString(rStatus.Res)
			// 			if err != nil {
			// 				fmt.Println("Failed to decode cluster status response", app, cluster, contextID, rStatus.Res, err)
			// 				return []distributionResource{}, err
			// 			}

			// 			status := distributionResource{}
			// 			if err := json.Unmarshal(data, &status); err != nil {
			// 				fmt.Println("Failed to unmarshal cluster status", app, cluster, contextID, string(data), err)
			// 				return []distributionResource{}, err
			// 			}

			// 			distributionResources = append(distributionResources, status)
			// 			break
			// 		}
			// 	}
			// }

			// TODO : verify the monitor bundle state should only have the resources created for the specific app context
			// the number of CR resoreces in the statuses should be equal to the number of CR resources created
			// no need to capture the resources created and validate against the statuses

			// if len(resources) == len(certificateRequestStatuses) {
			// 	logutils.Info(fmt.Sprintf("Cluster status contains the certificate for App: %s, ClusterGroup: %s and ContextID: %s", app, cluster, contextID),
			// 		logutils.Fields{})
			// 	// At this point we assume we have the certificate requests created and the status is available
			// 	// no need to retry , break the loop
			// 	statusReady = true
			// }
			// }

		}
	}

	return distributionResources, nil

}

func (c *CertDistributionClient) Status(cert, clusterProvider string) (CertEnrollmentStatus, error) {
	key := StateKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		AppName:         distributionApp}

	// get the current state of the
	stateInfo, err := common.NewStateClient().GetState(key)
	if err != nil {
		return CertEnrollmentStatus{}, err
	}

	status, err := status.PrepareCertEnrollmentStatusResult(stateInfo, "ready")
	if err != nil {
		fmt.Println(err.Error())
	}

	certEnrollmentStatus := CertEnrollmentStatus{
		DeployedStatus: status.DeployedStatus,
		ReadyStatus:    status.ReadyStatus,
		ReadyCounts:    status.ReadyCounts}

	for _, app := range status.Apps {
		certEnrollmentStatus.App = app.Name
	}

	for _, cluster := range status.Apps[0].Clusters {
		certEnrollmentStatus.ClusterProvider = cluster.ClusterProvider
		certEnrollmentStatus.Cluster = cluster.Cluster
		certEnrollmentStatus.Connectivity = cluster.Connectivity
	}

	for _, resource := range status.Apps[0].Clusters[0].Resources {
		r := Resource{
			Gvk:         resource.Gvk,
			Name:        resource.Name,
			ReadyStatus: resource.ReadyStatus,
		}
		certEnrollmentStatus.Resources = append(certEnrollmentStatus.Resources, r)
	}

	return certEnrollmentStatus, nil
}

func (c *CertDistributionClient) Terminate(cert, clusterProvider string) error {
	key := StateKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		AppName:         distributionApp}

	s := common.NewStateClient()
	// check the current state of the Instantiation, if any
	contextID, err := s.VerifyState(key, common.TerminateEvent)
	if err != nil {
		return err
	}

	// call resource synchronizer to delete the CSR from the issuing cluster
	if err := terminateDistribution(cert, clusterProvider, contextID); err != nil {
		return err
	}

	// update the state object for the cert resource
	if err := s.UpdateState(key, state.StateEnum.Terminated, contextID, false); err != nil {
		return err
	}

	return nil
}

func (c *CertDistributionClient) Update(cert, clusterProvider string) error {
	// get the ca cert
	caCert, err := getCertificate(cert, clusterProvider)
	if err != nil {
		return err
	}

	key := StateKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		AppName:         distributionApp}

	s := common.NewStateClient()
	// get the current state of the instantiation
	stateInfo, err := s.GetState(key)
	if err != nil {
		return err
	}

	contextID := state.GetLastContextIdFromStateInfo(stateInfo)
	if len(contextID) > 0 {
		// get the existing app context
		status, err := state.GetAppContextStatus(contextID)
		if err != nil {
			return err
		}
		if status.Status == appcontext.AppContextStatusEnum.Instantiated {
			// get all the clusters defined under this CA
			clusterGroups, err := getAllClusterGroup(cert, clusterProvider)
			if err != nil {
				return err
			}

			// instantiate a new appContext
			ctx := common.Context{
				AppName:    enrollmentApp,
				ClientName: clientName}
			if err := ctx.InitAppContext(); err != nil {
				return err
			}

			appContext := distributionAppContext{
				appContext:      ctx.AppContext,
				appHandle:       ctx.AppHandle,
				caCert:          caCert,        // CA
				clusterGroups:   clusterGroups, // clusters those are part of the CA
				clusterProvider: clusterProvider,
				contextID:       ctx.ContextID}

			// initialize the Instantiation
			if err = appContext.instantiateDistribution(); err != nil {
				return err
			}

			if err := state.UpdateAppContextStatusContextID(appContext.contextID, contextID); err != nil {
				return err
			}

			if err := notifyclient.CallRsyncUpdate(contextID, appContext.contextID); err != nil {
				return err
			}

			// update the state object for the cert resource
			if err := s.UpdateState(key, state.StateEnum.Updated, appContext.contextID, false); err != nil {
				return err
			}

			// subscribe to alerts
			stream, client, err := notifyclient.InvokeReadyNotify(appContext.contextID, clientName)
			if err != nil {
				fmt.Println("Failed to subscribe to alerts from the rsync gRPC server", appContext.contextID, err)
				return err
			}

			// do not wait for the resource synchronization to complete
			// handle the retrieval and verification in a go-routine
			// store the certificate generated by CR in mongo-db
			// update the state of enrollment to instantiated and return the api

			go retrieveDistributionStatus(client, stream)

			if err := stream.CloseSend(); err != nil {
				fmt.Println("Failed to close the send stream", appContext.contextID, err)
				return err
			}

		}

	}

	return nil
}

// func (c *CertDistributionClient) Delete() {

// }

// func (c *CertDistributionClient) Get() {

// }

// terminateEnrollment
func terminateDistribution(cert, clusterProvider, contextID string) error {
	if err := notifyclient.CallRsyncUninstall(contextID); err != nil {
		return err
	}

	// get all the clusters defined under this CA
	clusters, err := getAllClusterGroup(cert, clusterProvider)
	if err != nil {
		return err
	}

	// cClient := cluster.NewClusterClient()
	// certClient := certificate.NewCertificateClient()
	for _, cluster := range clusters {
		// TODO: Confirm if we need to veirfy the cluster exists or not
		switch strings.ToLower(cluster.Spec.Scope) {
		case "name":
			// cr, err := certClient.GetClusterProviderCertRequest(cert, cluster.Spec.Name, cluster.MetaData.Name, clusterProvider)
			// if err != nil &&
			// 	strings.Compare(err.Error(), "CertificateRequest not found") == 0 {
			// 	continue
			// }
			// // all other errors, returns
			// if err != nil {
			// 	fmt.Println("Failed to get clusters certificate request", cert, cluster.Spec.Name, cluster.MetaData.Name, clusterProvider, err.Error())
			// 	return err
			// }

			// if err := certClient.DeletelusterProviderCertRequest(cert, cluster.Spec.Name, cluster.MetaData.Name, clusterProvider, cr.MetaData.Name); err != nil {
			// 	fmt.Println("Failed to delete clusters certificate request", cert, cluster.Spec.Name, cluster.MetaData.Name, clusterProvider, err.Error())

			// 	return err
			// }

		case "label":
			// // get clusters by label TODO: Confirm if we need to veirfy the cluster exists or not
			// list, err := cClient.GetClustersWithLabel(clusterProvider, cluster.Spec.Label)
			// if err != nil {
			// 	fmt.Println("Failed to get clusters by label", cert, clusterProvider, cluster.Spec.Label, "", err)
			// 	return err
			// }

			// for _, name := range list {
			// 	cr, err := certClient.GetClusterProviderCertRequest(cert, name, cluster.MetaData.Name, clusterProvider)
			// 	if err != nil &&
			// 		strings.Compare(err.Error(), "CertificateRequest not found") == 0 {
			// 		continue
			// 	}
			// 	// all other errors, returns
			// 	if err != nil {
			// 		fmt.Println("Failed to get clusters certificate request", cert, name, cluster.MetaData.Name, clusterProvider, err.Error())
			// 		return err
			// 	}

			// 	if err := certClient.DeletelusterProviderCertRequest(cert, name, cluster.MetaData.Name, clusterProvider, cr.MetaData.Name); err != nil {
			// 		fmt.Println("Failed to delete clusters certificate request", cert, cluster.Spec.Name, cluster.MetaData.Name, clusterProvider, err.Error())
			// 		return err
			// 	}
			// }
		}
	}

	return nil
}
