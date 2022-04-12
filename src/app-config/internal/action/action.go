package action

import (
	"context"
	"fmt"
	"time"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/app-config/internal/utils"
	"gitlab.com/project-emco/core/emco-base/src/app-config/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	rsyncclient "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/notifyclient"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	readynotifypb "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/readynotify"
)

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

// SEPARATOR used while creating resourceNames to store in etcd
const (
	SEPARATOR      = "+"
	transportError = "rpc error: code = Unavailable desc = transport is closing"
)

// UpdateAppContext is the method which calls the backend logic of this controller.
func UpdateAppContext(intentName, appContextID string) error {
	log.Info("Begin updating app context ", log.Fields{"intent-name": intentName, "appcontext": appContextID})

	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(appContextID)
	if err != nil {
		log.Error("Loading AppContext failed ", log.Fields{"intent-name": intentName, "appcontext": appContextID, "Error": err.Error()})
		return pkgerrors.Errorf("Internal error")
	}

	caMeta, err := ac.GetCompositeAppMeta()
	if err != nil {
		log.Error("Error getting metadata for AppContext ", log.Fields{"intent-name": intentName, "appcontext": appContextID, "Error": err.Error()})
		return pkgerrors.Errorf("Internal error")
	}

	p := caMeta.Project
	ca := caMeta.CompositeApp
	cv := caMeta.Version
	dig := caMeta.DeploymentIntentGroup

	acList, err := module.NewAppConfigClient().GetAllAppConfig(p, ca, cv, dig)
	if err != nil {
		return pkgerrors.Errorf("Error in getting appconfig for composite app %s\n", ca)
	}
	for _, aConfig := range acList {
		cr, err := module.NewResTypeClient().GetResType(aConfig.Spec.ResType, p, ca, cv)
		if err != nil {
			log.Error("Error in getting restype for appconfig", log.Fields{})
			continue
		}
		_, err = cluster.NewClusterClient().GetClustersWithLabel(aConfig.Spec.ClusterInfo.ClusterProvider, aConfig.Spec.ClusterInfo.ClusterLabel)
		if intentName == "Instantiate" {
			stream, client, err := rsyncclient.InvokeReadyNotify(appContextID, aConfig.Metadata.Name)
			if err != nil {
				log.Error("Error in callRsyncReadyNotify", log.Fields{
					"error": err, "appContextID": appContextID,
				})
				return pkgerrors.Wrap(err, "Error in callRsyncReadyNotify")
			}
			acKey := fmt.Sprintf("%s-%s-%s", p, ca, dig)
			acContext, ok := module.AppCfgState[acKey].Data[aConfig.Metadata.Name]
			if !ok {
				return pkgerrors.New("AppCfg state is empty...")
			}
			acContext.Lock.Lock()
			acContext.Client = client
			acContext.Lock.Unlock()

			go processAppConfig(appContextID, aConfig, cr.Spec.AppName, stream, client, acContext)
		} else if intentName == "Terminate" {
			// call unsubscribe

			err = module.TerminateAppConfig(appContextID, aConfig.Metadata.Name, p, ca, cv, dig)

		} else {
			return pkgerrors.Errorf("Unknown appcfg intent name")
		}
	}

	return nil

}

func DeployAppConfig(ac appcontext.AppContext, appContextID string, appcfg module.AppConfig, app string) error {
	caMeta, err := ac.GetCompositeAppMeta()
	if err != nil {
		log.Error("Error getting metadata for AppContext ", log.Fields{"appcontext": appContextID, "Error": err.Error()})
		return pkgerrors.Errorf("Internal error")
	}

	p := caMeta.Project
	ca := caMeta.CompositeApp
	cv := caMeta.Version
	dig := caMeta.DeploymentIntentGroup

	log.Info("Now applying the app config", log.Fields{})
	_, ctxval, err := module.InstantiateAppConfig(appcfg.Metadata.Name, p, ca, cv, dig)
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating appConfig context")
	}
	module.WaitUpdateAppConfigState(ctxval.(string), state.StateEnum.Instantiated, appcfg.Metadata.Name, p, ca, cv, dig)
	if err != nil {
		return pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return nil

}

func processAppConfig(appContextID string, appcfg module.AppConfig, app string, stream readynotifypb.ReadyNotify_AlertClient, cl readynotifypb.ReadyNotifyClient, acContext *module.AppCfgContext) error {
	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(appContextID)
	if err != nil {
		log.Error("Error getting AppContext with Id", log.Fields{
			"error": err, "appContextID": appContextID,
		})
		return pkgerrors.Wrapf(err, "Error getting AppContext with Id: %v", appContextID)
	}

	err = processAlertForAppConfig(stream, appContextID, appcfg, app)
	if err != nil {
		log.Error("Unable to process the alert for app config discovery", log.Fields{"err": err.Error()})
		stream.CloseSend()
		return pkgerrors.Wrapf(err, "siv siv Exiting after context cancel: %v", appContextID)

	}

	// call unsubscribe
	_, err = cl.Unsubscribe(context.Background(), &readynotifypb.Topic{ClientName: appcfg.Metadata.Name, AppContext: appContextID})
	if err != nil {
		log.Error("[ReadyNotify gRPC] Failed to unsubscribe to alerts", log.Fields{"err": err, "appContextId": appContextID})
		return err
	}

	// close the stream
	stream.CloseSend()
	// Get the appcontext status value
	acStatus, err := state.GetAppContextStatus(appContextID)
	if err != nil {
		log.Error("Unable to get the parent's app context status", log.Fields{"err": err.Error()})
		return pkgerrors.Wrap(err, "Unable to get the status of the app context")
	}
	acContext.Lock.Lock()
	acContext.Client = nil
	acContext.Lock.Unlock()

	if acStatus.Status == appcontext.AppContextStatusEnum.Instantiated {
		err := DeployAppConfig(ac, appContextID, appcfg, app)
		if err != nil {
			log.Error("Unable to deploy the App Config", log.Fields{"err": err.Error(), "AppContextID": appContextID})
			return pkgerrors.Wrap(err, "Unable to deploy the App Cofnig for the contextID: "+appContextID)
		}

	}
	return nil

}

func processAlertForAppConfig(stream readynotifypb.ReadyNotify_AlertClient, appContextID string, appcfg module.AppConfig, app string) error {

	for {

		appReady := false
		// Now check whether the parent app context has been "Instantiated".

		acStatus, err := state.GetAppContextStatus(appContextID)
		if err != nil {
			log.Warn("[ReadyNotify gRPC] Unable to get status of app context", log.Fields{"err": err, "appCID": appContextID})
			time.Sleep(1 * time.Second)
			continue
		}

		if acStatus.Status == appcontext.AppContextStatusEnum.Instantiated {
			log.Info("Parent's app context is 'Instantiated'. Checking for app deployment", log.Fields{"appContextID": appContextID})
			// Now check the status of the app deployed
			condition, err := utils.CheckDeploymentStatus(appContextID, app)
			if err != nil {
				log.Error("Unable to check the deployment status of the appconfig app", log.Fields{"err": err.Error(), "App": app})
				return pkgerrors.Wrap(err, "Unable to check the deployment status of the appconfig app")
			}
			if condition {
				// Server App has been successfully deployed
				log.Info("AppConfig App has been successfully deployed", log.Fields{"app": app})

				// Check for the loadbalancer external IP
				var ac appcontext.AppContext
				_, err := ac.LoadAppContext(appContextID)
				if err != nil {
					log.Error("Error loading AppContext", log.Fields{
						"error": err,
					})
					return pkgerrors.Wrap(err, "Error loading AppContext")
				}

				// Get the clusters in the appcontext for this app
				clusters, err := ac.GetClusterNames(app)
				if err != nil {
					log.Error("Unable to get the cluster names",
						log.Fields{"AppName": app, "Error": err})
					return pkgerrors.Wrap(err, "Unable to get the cluster names")
				}
				for _, cluster := range clusters {
					_, err := utils.GetClusterResources(appContextID, app, cluster)
					if err != nil {
						log.Error("Unable to get the cluster resources",
							log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
						return pkgerrors.Wrap(err, "Unable to get the cluster resources")
					}
				}
				appReady = true
				log.Info("Leaving the process loop for deploying the app config...", log.Fields{})
				return nil
			} else {
				log.Info("AppConfig App has not been deployed yet", log.Fields{"appconfigApp": app})
			}
		} else if acStatus.Status == appcontext.AppContextStatusEnum.Instantiating ||
			acStatus.Status == appcontext.AppContextStatusEnum.Created {
			log.Info("Parent's app ctxt is still 'Instantiating' state", log.Fields{"appContextID": appContextID})
		} else { // If the parent's appContext is "Terminating/Terminated/InstantiateFailed/TerminateFailed"
			log.Error("Parent's app context is not in 'Instantiated' state", log.Fields{"appContextID": appContextID})
			return pkgerrors.New("Parent's app context is not in 'Instantiated' state")
		}

		if !appReady {
			// Here the code gets blocked until the load balancer external IP is obtained
			if stream != nil {
				resp, err := stream.Recv()
				if err != nil {
					log.Error("[ReadyNotify gRPC] Failed to receive notification", log.Fields{"err": err.Error()})
					if err.Error() == transportError {
						log.Error("appcfg: transport error waiting  5 seconds ...", log.Fields{})
						time.Sleep(5 * time.Second)
						continue
					}
					return pkgerrors.Wrap(err, "appcfg [Process] Failed to receive notification")
				}

				appContextID = resp.AppContext
				// log.Info("[ReadyNotify gRPC] Received alert from rsync", log.Fields{"appContextId": appContextID, "err": err})
			} else {
				// if stream is nil then poll until the load balancer IP is obtained
				time.Sleep(5 * time.Second)
				continue
			}

		}

	}

}
