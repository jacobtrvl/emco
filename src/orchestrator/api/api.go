// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"github.com/gorilla/mux"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	controller "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
)

var moduleClient *moduleLib.Client

// NewRouter creates a router that registers the various urls that are supported
func NewRouter(projectClient moduleLib.ProjectManager,
	compositeAppClient moduleLib.CompositeAppManager,
	appClient moduleLib.AppManager,
	ControllerClient controller.ControllerManager,
	genericPlacementIntentClient moduleLib.GenericPlacementIntentManager,
	appIntentClient moduleLib.AppIntentManager,
	deploymentIntentGrpClient moduleLib.DeploymentIntentGroupManager,
	intentClient moduleLib.IntentManager,
	compositeProfileClient moduleLib.CompositeProfileManager,
	appProfileClient moduleLib.AppProfileManager,
	instantiationClient moduleLib.InstantiationManager) *mux.Router {

	router := mux.NewRouter().PathPrefix("/v2").Subrouter()

	moduleClient = moduleLib.NewClient()

	//setting routes for project
	if projectClient == nil {
		projectClient = moduleClient.Project

	}
	projHandler := projectHandler{
		client: projectClient,
	}
	if ControllerClient == nil {
		ControllerClient = moduleClient.Controller
	}
	controlHandler := controllerHandler{
		client: ControllerClient,
	}
	router.HandleFunc("/projects", projHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project}", projHandler.updateHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}", projHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects", projHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}", projHandler.deleteHandler).Methods("DELETE")

	//setting routes for compositeApp
	if compositeAppClient == nil {
		compositeAppClient = moduleClient.CompositeApp
	}
	compAppHandler := compositeAppHandler{
		client: compositeAppClient,
	}
	router.HandleFunc("/projects/{project}/composite-apps", compAppHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}", compAppHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps", compAppHandler.getAllCompositeAppsHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}", compAppHandler.deleteHandler).Methods("DELETE")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}", compAppHandler.updateHandler).Methods("PUT")

	if appClient == nil {
		appClient = moduleClient.App
	}
	appHandler := appHandler{
		client: appClient,
	}

	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/apps", appHandler.createAppHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/apps/{app}", appHandler.getAppHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/apps/{app}", appHandler.updateAppHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/apps", appHandler.getAppHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/apps/{app}", appHandler.deleteAppHandler).Methods("DELETE")

	if compositeProfileClient == nil {
		compositeProfileClient = moduleClient.CompositeProfile
	}
	compProfilepHandler := compositeProfileHandler{
		client: compositeProfileClient,
	}
	if appProfileClient == nil {
		appProfileClient = moduleClient.AppProfile
	}
	appProfileHandler := appProfileHandler{
		client: appProfileClient,
	}

	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles", compProfilepHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles", compProfilepHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}", compProfilepHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}", compProfilepHandler.updateHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}", compProfilepHandler.deleteHandler).Methods("DELETE")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles", appProfileHandler.createAppProfileHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles", appProfileHandler.getAppProfileHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles", appProfileHandler.getAppProfileHandler).Queries("app", "{app}")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles/{appProfile}", appProfileHandler.getAppProfileHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles/{appProfile}", appProfileHandler.updateAppProfileHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles/{appProfile}", appProfileHandler.deleteAppProfileHandler).Methods("DELETE")

	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles", appProfileHandler.createAppProfileHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles", appProfileHandler.getAppProfileHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles", appProfileHandler.getAppProfileHandler).Queries("app", "{app}")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles/{appProfile}", appProfileHandler.getAppProfileHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles/{appProfile}", appProfileHandler.deleteAppProfileHandler).Methods("DELETE")

	router.HandleFunc("/controllers", controlHandler.createHandler).Methods("POST")
	router.HandleFunc("/controllers", controlHandler.getHandler).Methods("GET")
	router.HandleFunc("/controllers/{controller}", controlHandler.putHandler).Methods("PUT")
	router.HandleFunc("/controllers/{controller}", controlHandler.getHandler).Methods("GET")
	router.HandleFunc("/controllers/{controller}", controlHandler.deleteHandler).Methods("DELETE")

	//setting routes for genericPlacementIntent
	if genericPlacementIntentClient == nil {
		genericPlacementIntentClient = moduleClient.GenericPlacementIntent
	}

	genericPlacementIntentHandler := genericPlacementIntentHandler{
		client: genericPlacementIntentClient,
	}
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents", genericPlacementIntentHandler.createGenericPlacementIntentHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}", genericPlacementIntentHandler.getGenericPlacementHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents", genericPlacementIntentHandler.getAllGenericPlacementIntentsHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}", genericPlacementIntentHandler.deleteGenericPlacementHandler).Methods("DELETE")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}", genericPlacementIntentHandler.putGenericPlacementHandler).Methods("PUT")

	//setting routes for AppIntent
	if appIntentClient == nil {
		appIntentClient = moduleClient.AppIntent
	}

	appIntentHandler := appIntentHandler{
		client: appIntentClient,
	}

	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents", appIntentHandler.createAppIntentHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents/{genericAppPlacementIntent}", appIntentHandler.getAppIntentHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents", appIntentHandler.getAllAppIntentsHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents/", appIntentHandler.getAllIntentsByAppHandler).Queries("app", "{app}")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents/{genericAppPlacementIntent}", appIntentHandler.deleteAppIntentHandler).Methods("DELETE")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents/{genericAppPlacementIntent}", appIntentHandler.putAppIntentHandler).Methods("PUT")

	//setting routes for deploymentIntentGroup
	if deploymentIntentGrpClient == nil {
		deploymentIntentGrpClient = moduleClient.DeploymentIntentGroup
	}

	deploymentIntentGrpHandler := deploymentIntentGroupHandler{
		client: deploymentIntentGrpClient,
	}
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups", deploymentIntentGrpHandler.createDeploymentIntentGroupHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}", deploymentIntentGrpHandler.getDeploymentIntentGroupHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups", deploymentIntentGrpHandler.getAllDeploymentIntentGroupsHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}", deploymentIntentGrpHandler.deleteDeploymentIntentGroupHandler).Methods("DELETE")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}", deploymentIntentGrpHandler.putDeploymentIntentGroupHandler).Methods("PUT")

	// setting routes for AddingIntents
	if intentClient == nil {
		intentClient = moduleClient.Intent
	}

	intentHandler := intentHandler{
		client: intentClient,
	}

	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents", intentHandler.addIntentHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents/{groupIntent}", intentHandler.getIntentHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents", intentHandler.getAllIntentsHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents/", intentHandler.getIntentByNameHandler).Queries("intent", "{intent}")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents/{groupIntent}", intentHandler.deleteIntentHandler).Methods("DELETE")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents/{groupIntent}", intentHandler.putIntentHandler).Methods("PUT")

	// setting routes for Instantiation
	if instantiationClient == nil {
		instantiationClient = moduleClient.Instantiation
	}

	instantiationHandler := instantiationHandler{
		client: instantiationClient,
	}

	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/approve", instantiationHandler.approveHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/terminate", instantiationHandler.terminateHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/instantiate", instantiationHandler.instantiateHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/stop", instantiationHandler.stopHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/status", instantiationHandler.statusHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/status",
		instantiationHandler.statusHandler).Queries("instance", "{instance}", "type", "{type}", "output", "{output}", "app", "{app}", "cluster", "{cluster}", "resource", "{resource}", "apps", "{apps}", "clusters", "{clusters}", "resources", "{resources}")

	// setting routes for Update
	updateHandler := updateHandler{
		client: instantiationClient,
	}

	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/migrate", updateHandler.migrateHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/update", updateHandler.updateHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/rollback", updateHandler.rollbackHandler).Methods("POST")
	return router
}
