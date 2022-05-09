// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/common"
	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/notifyclient"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/readynotify"
	yamlV2 "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const enrollmentApp = "cert-enrollment"
const clientName = "cacert"
const appMeta = "meta"

// CertEnrollmentManager
type CertEnrollmentManager interface {
	Instantiate(cert, clusterProvider string) error
	Status(cert, clusterProvider string) (CertEnrollmentStatus, error)
	Terminate(cert, clusterProvider string) error
	Update(cert, clusterProvider string) error
	// Delete()
	// Get()
}

type enrollmentAppContext struct {
	appContext      appcontext.AppContext
	appHandle       interface{}
	caCert          certificate.Cert // CA
	clusters        []ClusterGroup   // clusters those are part of the CA
	clusterProvider string
	contextID       string
	resorder        []string
	// resources       []string // name of the CR created for each clusters
}

type CertEnrollmentStatus struct {
	DeployedStatus  appcontext.StatusValue `json:"deployedStatus,omitempty,inline"`
	ReadyStatus     string                 `json:"readyStatus,omitempty,inline"`
	ReadyCounts     map[string]int         `json:"readyCounts,omitempty,inline"`
	App             string                 `json:"app,omitempty,inline"`
	ClusterProvider string                 `json:"clusterProvider,omitempty"`
	Cluster         string                 `json:"cluster,omitempty"`
	Connectivity    string                 `json:"connectivity,omitempty"`
	Resources       []Resource             `json:"resources,omitempty"`
}
type Resource struct {
	Gvk         schema.GroupVersionKind `json:"GVK,omitempty"`
	Name        string                  `json:"name,omitempty"`
	ReadyStatus string                  `json:"readyStatus,omitempty"`
}

// CertEnrollmentClient
type CertEnrollmentClient struct {
	db DbInfo
}

// NewCertEnrollmentClient
func NewCertEnrollmentClient() *CertEnrollmentClient {
	return &CertEnrollmentClient{
		db: DbInfo{
			storeName: "resources",
			tagMeta:   "data",
			tagState:  "stateInfo"}}
}

func (c *CertEnrollmentClient) Instantiate(cert, clusterProvider string) error {
	key := StateKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		AppName:         enrollmentApp}

	s := common.NewStateClient()
	// check the current state of the Instantiation, if any
	if _, err := s.VerifyState(key, common.InstantiateEvent); err != nil {
		return err
	}

	// get the ca cert
	caCert, err := getCertificate(cert, clusterProvider)
	if err != nil {
		return err
	}

	// get all the clusters defined under this CA
	clusters, err := getAllClusterGroup(cert, clusterProvider)
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

	appContext := enrollmentAppContext{
		appContext:      ctx.AppContext,
		appHandle:       ctx.AppHandle,
		caCert:          caCert,   // CA
		clusters:        clusters, // clusters those are part of the CA
		clusterProvider: clusterProvider,
		contextID:       ctx.ContextID}

	// initialize the Instantiation
	if err = appContext.instantiateEnrollment(); err != nil {
		return err
	}

	if err = appContext.callRsyncInstall(); err != nil {
		return err
	}

	if err := s.UpdateState(key, state.StateEnum.Instantiated, ctx.ContextID, false); err != nil {
		return err
	}

	return nil
}

func (c *CertEnrollmentClient) Status(cert, clusterProvider string) (CertEnrollmentStatus, error) {
	key := StateKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		AppName:         enrollmentApp}

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

func (c *CertEnrollmentClient) Terminate(cert, clusterProvider string) error {
	key := StateKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		AppName:         enrollmentApp}

	s := common.NewStateClient()
	// // check the current state of the Instantiation, if any
	contextID, err := s.VerifyState(key, common.TerminateEvent)
	if err != nil {
		return err
	}

	// call resource synchronizer to delete the CSR from the issuing cluster
	if err := terminateEnrollment(cert, clusterProvider, contextID); err != nil {
		return err
	}

	// update the state object for the cert resource
	if err := s.UpdateState(key, state.StateEnum.Terminated, contextID, false); err != nil {
		return err
	}

	return nil
}

func (c *CertEnrollmentClient) Update(cert, clusterProvider string) error {
	// get the ca cert
	caCert, err := getCertificate(cert, clusterProvider)
	if err != nil {
		return err
	}

	key := StateKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		AppName:         enrollmentApp}

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
			clusters, err := getAllClusterGroup(cert, clusterProvider)
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

			appContext := enrollmentAppContext{
				appContext:      ctx.AppContext,
				appHandle:       ctx.AppHandle,
				caCert:          caCert,   // CA
				clusters:        clusters, // clusters those are part of the CA
				clusterProvider: clusterProvider,
				contextID:       ctx.ContextID}

			// initialize the Instantiation
			if err = appContext.instantiateEnrollment(); err != nil {
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
			// go retrieveCertificateRequests(client, stream, context.resources)

			go retrieveCertificateRequests(client, stream)

			if err := stream.CloseSend(); err != nil {
				fmt.Println("Failed to close the send stream", appContext.contextID, err)
				return err
			}

		}

	}

	return nil
}

func (context *enrollmentAppContext) instantiateEnrollment() error {
	var log = func(message, cert, clusterProvider, label, name string, err error) {
		fields := make(logutils.Fields)
		fields["Cert"] = cert
		fields["ClusterProvider"] = clusterProvider
		if err != nil {
			fields["Error"] = err.Error()
		}
		if len(label) > 0 {
			fields["Label"] = label
		}
		if len(name) > 0 {
			fields["Name"] = name
		}
		logutils.Error(message, fields)
	}

	fmt.Println("Create AddLevelValue", context.caCert.MetaData.Name, context.clusterProvider)

	meta := map[string]string{}
	meta["cert"] = context.caCert.MetaData.Name
	meta["clusterProvider"] = context.clusterProvider
	lHandle, err := context.appContext.AddLevelValue(context.appHandle, appMeta, meta)
	if err != nil {
		return err
	}

	fmt.Println("AddLevelValue Handle: ", lHandle)

	// add handle for the issuing cluster
	clusterHandle, err := context.appContext.AddCluster(context.appHandle,
		strings.Join([]string{context.caCert.Spec.IssuingCluster.ClusterProvider, context.caCert.Spec.IssuingCluster.Cluster}, "+"))
	if err != nil {
		log("Failed to add cluster to the context", context.caCert.MetaData.Name, context.caCert.Spec.IssuingCluster.ClusterProvider, "", context.caCert.Spec.IssuingCluster.Cluster, err)
		return err
	}

	fmt.Println("clusterHandle: ", clusterHandle)

	cClient := cluster.NewClusterClient()
	// create a CertificateRequest for each cluster
	for _, cluster := range context.clusters {
		// TODO: Confirm if we need to veirfy the cluster exists or not
		switch strings.ToLower(cluster.Spec.Scope) {
		case "name":
			// get cluster by name TODO: Confirm if we need to veirfy the cluster exists or not
			if _, err := cClient.GetCluster(context.clusterProvider, cluster.Spec.Name); err != nil {
				log("Failed to get the clutster by name", context.caCert.MetaData.Name, context.clusterProvider, "", cluster.Spec.Name, err)
				return err
			}

			// 	// get the common name for the cluster specific certificate
			// val, err := cClient.GetClusterKvPairsValue(context.clusterProvider, cluster.Spec.Name,"csrkvpairs","commonname")
			// if err != nil {
			// 	fmt.Println(err.Error())
			// }
			// fmt.Println(val.(string))

			if err := context.createCertificateRequestResource(cluster.Spec.Name, clusterHandle); err != nil {
				return err
			}

		case "label":
			// get clusters by label TODO: Confirm if we need to veirfy the cluster exists or not
			list, err := cClient.GetClustersWithLabel(context.clusterProvider, cluster.Spec.Label)
			if err != nil {
				log("Failed to get clusters by label", context.caCert.MetaData.Name, context.clusterProvider, cluster.Spec.Label, "", err)
				return err
			}

			for _, name := range list {
				if err := context.createCertificateRequestResource(name, clusterHandle); err != nil {
					return err
				}
			}
		}
	}

	resOrder, err := json.Marshal(map[string][]string{"resorder": context.resorder})
	if err != nil {
		return err
	}

	// add instruction under given handle and type
	iHandle, err := context.appContext.AddInstruction(clusterHandle, "resource", "order", string(resOrder))
	if err != nil {
		if er := context.appContext.DeleteCompositeApp(); er != nil {
			logutils.Warn("Failed to delete the compositeApp", logutils.Fields{
				"contextID": context.contextID,
				"Error":     err.Error()})
		}
		return err
	}

	fmt.Println("AddInstruction Handle: ", iHandle)

	//  invoke Rsync to create the CR resources in the issuing cluster
	return nil
}

func (context *enrollmentAppContext) createCertificateRequestResource(cluster string, clusterHandle interface{}) error {
	// create a CR
	crName := certificateRequestName(context.caCert.MetaData.Name, cluster, context.clusterProvider, context.contextID) // TODO: this needs to be a unique name, check the format
	context.caCert.Spec.CertificateSigningInfo.Subject.Names.CommonName = strings.Join([]string{cluster, context.clusterProvider, context.caCert.MetaData.Name, "ca"}, "-")
	fmt.Println("CommonName: ", context.caCert.Spec.CertificateSigningInfo.Subject.Names.CommonName)
	cr, err := certificate.CreateCertificateRequest(context.caCert, crName)
	if err != nil {
		return err
	}

	value, err := yamlV2.Marshal(cr)
	if err != nil {
		return err
	}

	// add resource under app and cluster
	if err := addResource(context.appContext, clusterHandle, cr.ResourceName(), string(value)); err != nil {
		deleteCompositeApp(context.appContext) // TODO: COnfirm if we need to delete this for any single resource add
		return err
	}

	context.resorder = append(context.resorder, cr.ResourceName())

	return nil
}

func (context *enrollmentAppContext) callRsyncInstall() error {
	var log = func(message, contextID string, err error) {
		fields := make(logutils.Fields)
		fields["ContextID"] = contextID
		fields["Client"] = clientName
		if err != nil {
			fields["Error"] = err.Error()
		}
		logutils.Error(message, fields)
	}

	// invokes the rsync service
	if err := notifyclient.CallRsyncInstall(context.contextID); err != nil {
		log("Failed to invokes the rsync service", context.contextID, err)
		return err
	}

	// subscribe to alerts
	stream, client, err := notifyclient.InvokeReadyNotify(context.contextID, clientName)
	if err != nil {
		log("Failed to subscribe to alerts from the rsync gRPC server", context.contextID, err)
		return err
	}

	// do not wait for the resource synchronization to complete
	// handle the retrieval and verification in a go-routine
	// store the certificate generated by CR in mongo-db
	// update the state of enrollment to instantiated and return the api
	// go retrieveCertificateRequests(client, stream, context.resources)

	go retrieveCertificateRequests(client, stream)

	if err := stream.CloseSend(); err != nil {
		log("Failed to close the send stream", context.contextID, err)
		return err
	}

	return nil
}

// retrieveCertificateRequests
// func retrieveCertificateRequests(client readynotify.ReadyNotifyClient, stream readynotify.ReadyNotify_AlertClient, resources []string) error {
func retrieveCertificateRequests(client readynotify.ReadyNotifyClient, stream readynotify.ReadyNotify_AlertClient) {

	contextID := retrieveAppContext(client, stream)
	certificateRequests, err := checkCertificateRequestStatus(contextID) // check whether all certificates have been issued
	if err != nil {
		return
	}

	// At this point we will have certificateRequests created

	// TODO: Get the state and contextID from the state info
	// validate the context id
	// Why don;t we pass the variables from the top level, why we need to query context?

	processCertificateRequests(contextID, certificateRequests)

	if _, err := client.Unsubscribe(context.Background(), &readynotify.Topic{ClientName: clientName, AppContext: contextID}); err != nil {
		logutils.Error("[ReadyNotify gRPC] Failed to unsubscribe to alerts",
			logutils.Fields{"ContextID": contextID,
				"Error": err.Error()})
	}

}

func retrieveAppContext(client readynotify.ReadyNotifyClient, stream readynotify.ReadyNotify_AlertClient) string {
	var (
		contextID  string
		backOff    int = config.GetConfiguration().BackOff
		maxBackOff int = config.GetConfiguration().MaxBackOff
	)

	// retrieve the appContextID from the stream, wait till we get the notification response
	for {
		resp, err := stream.Recv()
		if err != nil {
			logutils.Error(fmt.Sprintf("Failed to receive ReadyNotify notification, retry after %d seconds.", backOff),
				logutils.Fields{
					"Error": err.Error()})
			// instead of retrying immediately, waits some amount of time between tries
			time.Sleep(time.Duration(backOff) * time.Second)

			if backOff*2 < maxBackOff {
				backOff *= 2
			} else {
				backOff = maxBackOff
			}

			continue
		}

		contextID = resp.AppContext
		logutils.Info("Received ReadyNotify notification alert",
			logutils.Fields{
				"appContextID": contextID})
		// received notification response
		break
	}

	return contextID
}

func processCertificateRequests(contextID string, certificateRequests []certificate.CertificateRequest) {
	// TODO:-context id can be retrieved from the state as well
	var appContext appcontext.AppContext
	ac, err := appContext.LoadAppContext(contextID)
	if err != nil {
		logutils.Error("Failed to load appContext",
			logutils.Fields{
				"ContextID": contextID,
				"Errpr":     err.Error()})
		return
	}
	fmt.Println("LoadAppContext: ", ac)

	appHandle, err := appContext.GetAppHandle(enrollmentApp)
	if err != nil {
		logutils.Error("Failed to load the app handle",
			logutils.Fields{
				"App":       enrollmentApp,
				"ContextID": contextID,
				"Errpr":     err.Error()})
		return

	}

	fmt.Println("GetAppHandle: ", appHandle)

	lHandle, err := appContext.GetLevelHandle(appHandle, appMeta)
	if err != nil {
		logutils.Error("Failed to load the app level handle",
			logutils.Fields{
				"App":       enrollmentApp,
				"ContextID": contextID,
				"Level":     appMeta,
				"Errpr":     err.Error()})
		return
	}

	fmt.Println("GetLevelHandle: ", lHandle)

	// get the value of 'appMeta' handle
	val, err := appContext.GetValue(lHandle)
	if err != nil {
		logutils.Error("Failed to load the app level handle",
			logutils.Fields{
				"App":       enrollmentApp,
				"ContextID": contextID,
				"Handle":    lHandle,
				"Level":     appMeta,
				"Errpr":     err.Error()})
		return
	}

	meta, ok := val.(map[string]interface{})
	if !ok {
		logutils.Error("Failed to load the app level handle",
			logutils.Fields{
				"App":       enrollmentApp,
				"ContextID": contextID,
				"Handle":    lHandle,
				"Level":     appMeta,
				"Errpr":     err.Error()})
		return
	}
	appMeta := struct {
		Cert            string
		ClusterProvider string
	}{
		Cert:            meta["cert"].(string),
		ClusterProvider: meta["clusterProvider"].(string),
	}

	fmt.Println("appMeta: ", appMeta)

	// get all the clusters defined under this CA
	clusters, err := getAllClusterGroup(appMeta.Cert, appMeta.ClusterProvider)
	if err != nil {
		logutils.Error("Failed to get clusters using the appMeta",
			logutils.Fields{
				"App":       enrollmentApp,
				"AppMeta":   appMeta,
				"ContextID": contextID,
				"Handle":    lHandle,
				"Level":     appMeta,
				"Errpr":     err.Error()})
		return
	}

	cClient := cluster.NewClusterClient()
	for _, cluster := range clusters {
		// TODO: Confirm if we need to veirfy the cluster exists or not
		switch strings.ToLower(cluster.Spec.Scope) {
		case "name":
			// get cluster by name TODO: Confirm if we need to veirfy the cluster exists or not
			if _, err := cClient.GetCluster(appMeta.ClusterProvider, cluster.Spec.Name); err != nil {
				fmt.Println("Failed to get the clutster by name", appMeta.Cert, appMeta.ClusterProvider, "", cluster.Spec.Name, err)
				return
			}

			if err := saveCertificate(appMeta.Cert, cluster.Spec.Name, cluster.MetaData.Name, appMeta.ClusterProvider, contextID, certificateRequests); err != nil {
				return
			}

		case "label":
			// get clusters by label TODO: Confirm if we need to veirfy the cluster exists or not
			list, err := cClient.GetClustersWithLabel(appMeta.ClusterProvider, cluster.Spec.Label)
			if err != nil {
				fmt.Println("Failed to get clusters by label", appMeta.Cert, appMeta.ClusterProvider, cluster.Spec.Label, "", err)
				return
			}

			for _, name := range list {
				if err := saveCertificate(appMeta.Cert, name, cluster.MetaData.Name, appMeta.ClusterProvider, contextID, certificateRequests); err != nil {
					return
				}
			}
		}
	}
}

func saveCertificate(cert, cluster, clusterGroup, clusterProvider, contextID string, certificateRequests []certificate.CertificateRequest) error {
	crName := certificateRequestName(cert, cluster, clusterProvider, contextID)
	certReady := false
	for _, cr := range certificateRequests {
		if cr.MetaData.Name == crName {
			cc := certificate.NewCertificateClient()
			if err := cc.SaveClusterProviderCertRequest(cert, cluster, clusterGroup, clusterProvider, cr); err != nil {
				return err
			}
			certReady = true
			break
		}
	}

	if !certReady {
		fmt.Println("certificateRequest is not ready for: ", crName)
	}

	return nil
}

// checkCertificateRequestStatus checks whether the LC from the provided appcontext has had all cluster certificates issued
// func checkCertificateRequestStatus(contextID string, resources []string) bool {
func checkCertificateRequestStatus(contextID string) ([]certificate.CertificateRequest, error) {
	var log = func(message, app, cluster, contextID, status string, err error) {
		fields := make(logutils.Fields)
		fields["ContextID"] = contextID
		if len(app) > 0 {
			fields["App"] = app
		}
		if len(cluster) > 0 {
			fields["Cluster"] = cluster
		}
		if len(status) > 0 {
			fields["Status"] = status
		}
		if err != nil {
			fields["Error"] = err.Error()
		}
		logutils.Error(message, fields)
	}

	var certificateRequestStatuses []certificate.CertificateRequest
	var appContext appcontext.AppContext
	// load the appContext
	_, err := appContext.LoadAppContext(contextID)

	// TODO: Confirm return err or false? Shoudl we retry in case of appContext errors
	if err != nil {
		log("Failed to load the appContext", "", "", contextID, "", err)
		return []certificate.CertificateRequest{}, err
	}

	// get the app instruction for 'order'
	appsOrder, err := appContext.GetAppInstruction("order")
	if err != nil {
		log("Failed to get the app instruction for the 'order' instruction type", "", "", contextID, "", err)
		return []certificate.CertificateRequest{}, err
	}

	var appList map[string][]string
	json.Unmarshal([]byte(appsOrder.(string)), &appList)

	for _, app := range appList["apporder"] {
		//  get all the clusters associated with the app
		clusters, err := appContext.GetClusterNames(app)
		if err != nil {
			log("Failed to list clusters", app, "", contextID, "", err)
			return []certificate.CertificateRequest{}, err
		}

		for _, cluster := range clusters {
			// get the resources
			resources, err := appContext.GetResourceNames(app, cluster)
			if err != nil {
				log("Failed to get the resources", app, cluster, contextID, "", err)
				return []certificate.CertificateRequest{}, err
			}

			fmt.Println("GetResourceNames: ", resources)

			// get the cluster handle
			cHandle, err := appContext.GetClusterHandle(app, cluster)
			if err != nil {
				log("Failed to get the cluster handle", app, cluster, contextID, "", err)
				return []certificate.CertificateRequest{}, err
			}

			// get the the cluster status handle
			sHandle, err := appContext.GetLevelHandle(cHandle, "status")
			if err != nil {
				log("Failed to get the handle of 'status'", app, cluster, contextID, "", err)
				return []certificate.CertificateRequest{}, err
			}

			// get the status of the cetificaterequests resource creation
			// wait for the resources to be created and available in the monitor resource bundle state
			// retry if the resources satatus are not available
			//  TODO: Confirm max retrying
			statusReady := false
			for !statusReady {
				// get the value of 'status' handle
				val, err := appContext.GetValue(sHandle)
				if err != nil {
					log("Failed to get the value of 'status' handle", app, cluster, contextID, "", err)
					continue
				}

				fmt.Println(val.(string))

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
					log("Failed to unmarshal cluster status", app, cluster, contextID, val.(string), err)
					return []certificate.CertificateRequest{}, err
				}

				if len(s.ResourceStatuses) == 0 {
					continue
				}

				certificateRequestStatuses = []certificate.CertificateRequest{}
				// for each resource make sure the certificate request is created and the status is available
				for _, resource := range resources {
					for _, rStatus := range s.ResourceStatuses {
						crsName := certificateRequestResourceName(rStatus.Name, rStatus.Kind)
						if crsName == resource {
							if len(rStatus.Res) == 0 {
								logutils.Warn(fmt.Sprintf("Cluster status does not contain the certificate details for %s", rStatus.Name),
									logutils.Fields{})
								break
							}

							data, err := base64.StdEncoding.DecodeString(rStatus.Res)
							if err != nil {
								log("Failed to decode cluster status response", app, cluster, contextID, rStatus.Res, err)
								return []certificate.CertificateRequest{}, err
							}
							fmt.Println(string(data))

							status := certificate.CertificateRequest{}
							if err := json.Unmarshal(data, &status); err != nil {
								log("Failed to unmarshal cluster status", app, cluster, contextID, string(data), err)
								return []certificate.CertificateRequest{}, err
							}

							certificateRequestStatuses = append(certificateRequestStatuses, status)
							break
						}
					}
				}

				// TODO : verify the monitor bundle state should only have the resources created for the specific app context
				// the number of CR resoreces in the statuses should be equal to the number of CR resources created
				// no need to capture the resources created and validate against the statuses

				if len(resources) == len(certificateRequestStatuses) {
					logutils.Info(fmt.Sprintf("Cluster status contains the certificate for App: %s, ClusterGroup: %s and ContextID: %s", app, cluster, contextID),
						logutils.Fields{})
					// At this point we assume we have the certificate requests created and the status is available
					// no need to retry , break the loop
					statusReady = true
				}
			}

		}
	}

	return certificateRequestStatuses, nil
}

// terminateEnrollment
func terminateEnrollment(cert, clusterProvider, contextID string) error {
	if err := notifyclient.CallRsyncUninstall(contextID); err != nil {
		return err
	}

	// get all the clusters defined under this CA
	clusters, err := getAllClusterGroup(cert, clusterProvider)
	if err != nil {
		return err
	}

	cClient := cluster.NewClusterClient()
	certClient := certificate.NewCertificateClient()
	for _, cluster := range clusters {
		// TODO: Confirm if we need to veirfy the cluster exists or not
		switch strings.ToLower(cluster.Spec.Scope) {
		case "name":
			cr, err := certClient.GetClusterProviderCertRequest(cert, cluster.Spec.Name, cluster.MetaData.Name, clusterProvider)
			if err != nil &&
				strings.Compare(err.Error(), "CertificateRequest not found") == 0 {
				continue
			}
			// all other errors, returns
			if err != nil {
				fmt.Println("Failed to get clusters certificate request", cert, cluster.Spec.Name, cluster.MetaData.Name, clusterProvider, err.Error())
				return err
			}

			if err := certClient.DeletelusterProviderCertRequest(cert, cluster.Spec.Name, cluster.MetaData.Name, clusterProvider, cr.MetaData.Name); err != nil {
				fmt.Println("Failed to delete clusters certificate request", cert, cluster.Spec.Name, cluster.MetaData.Name, clusterProvider, err.Error())

				return err
			}

		case "label":
			// get clusters by label TODO: Confirm if we need to veirfy the cluster exists or not
			list, err := cClient.GetClustersWithLabel(clusterProvider, cluster.Spec.Label)
			if err != nil {
				fmt.Println("Failed to get clusters by label", cert, clusterProvider, cluster.Spec.Label, "", err)
				return err
			}

			for _, name := range list {
				cr, err := certClient.GetClusterProviderCertRequest(cert, name, cluster.MetaData.Name, clusterProvider)
				if err != nil &&
					strings.Compare(err.Error(), "CertificateRequest not found") == 0 {
					continue
				}
				// all other errors, returns
				if err != nil {
					fmt.Println("Failed to get clusters certificate request", cert, name, cluster.MetaData.Name, clusterProvider, err.Error())
					return err
				}

				if err := certClient.DeletelusterProviderCertRequest(cert, name, cluster.MetaData.Name, clusterProvider, cr.MetaData.Name); err != nil {
					fmt.Println("Failed to delete clusters certificate request", cert, cluster.Spec.Name, cluster.MetaData.Name, clusterProvider, err.Error())
					return err
				}
			}
		}
	}

	return nil
}
