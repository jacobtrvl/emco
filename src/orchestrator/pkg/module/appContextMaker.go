// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"encoding/json"
	"fmt"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	pkgerrors "github.com/pkg/errors"
	gpic "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/gpic"
)

type Instantiator struct {
	project              string
	compositeApp         string
	compAppVersion       string
	deploymentIntent     string
	deploymentIntentGrp DeploymentIntentGroup
}

// MakeAppContext shall make an app context and store the app context into etcd. This shall return contextForCompositeApp
func (i *Instantiator) MakeAppContext() (contextForCompositeApp, error) {

	dcmClusters, namespace, level, err := getLogicalCloudInfo(i.project, i.deploymentIntentGrp.Spec.LogicalCloud)
	if err != nil {
		return contextForCompositeApp{}, err
	}

	cca, err := i.makeAppContextForCompositeApp(namespace, level)
	if err != nil {
		return contextForCompositeApp{}, err
	}

	err = i.storeAppContextIntoRunTimeDB(cca, dcmClusters, namespace)
	if err != nil {
		deleteAppContext(cca.context)
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error in storeAppContextIntoETCd")
	}

	return cca, nil
}

func (i *Instantiator) makeAppContextForCompositeApp(namespace, level string) (contextForCompositeApp, error) {
	context := appcontext.AppContext{}
	ctxval, err := context.InitAppContext()
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error creating AppContext CompositeApp")
	}
	compositeHandle, err := context.CreateCompositeApp()
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error creating CompositeApp handle")
	}
	rName := i.deploymentIntentGrp.Spec.Version //rName is releaseName
	err = context.AddCompositeAppMeta( appcontext.CompositeAppMeta {
		Project: i.project,
		CompositeApp: i.compositeApp,
		Version: i.compAppVersion,
		Release: rName,
		DeploymentIntentGroup: i.deploymentIntent,
		Namespace: namespace,
		Level: level})
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error Adding CompositeAppMeta")
	}

	m, _ := context.GetCompositeAppMeta()
	log.Info(":: The meta data stored in the runtime context :: ", log.Fields{"Project": m.Project, "CompositeApp": m.CompositeApp, "Version": m.Version, "Release": m.Release, "DeploymentIntentGroup": m.DeploymentIntentGroup})

	cca := contextForCompositeApp{context: context, ctxval: ctxval, compositeAppHandle: compositeHandle}
	return cca, nil

}

func (i *Instantiator)storeAppContextIntoRunTimeDB(cxtForCApp contextForCompositeApp, dcmClusters []Cluster, namespace string) error {

	ctx := cxtForCApp.context
	// for recording the app order instruction
	var appOrdInsStr appOrderInstr
	// for recording the app dependency
	var appDepStr appDepInstr
	appDepStr.AppDepMap = make(map[string]string)

	overrideValues := i.deploymentIntentGrp.Spec.OverrideValuesObj
	rName := i.deploymentIntentGrp.Spec.Version //rName is releaseName
	cp := i.deploymentIntentGrp.Spec.Profile
	gIntent, err := findGenericPlacementIntent(i.project, i.compositeApp, i.compAppVersion, i.deploymentIntent)
	if err != nil {
		return err
	}
	log.Info(":: The name of the GenPlacIntent ::", log.Fields{"GenPlmtIntent": gIntent})

	allApps, err := NewAppClient().GetApps(i.project, i.compositeApp, i.compAppVersion)
	if err != nil {
		return pkgerrors.Wrap(err, "Not finding the apps")
	}

	// Check dependency between APPS and check for cyclic dependency
	if !checkDependency(allApps, i.project, i.compositeApp, i.compAppVersion) {
		str := fmt.Sprint("Cyclic Dependency between apps found in composite app:", i.compositeApp)
		log.Error(str, log.Fields{"composite app": i.compositeApp})
		return pkgerrors.New(str)
	}
	for _, eachApp := range allApps {
		appOrdInsStr.Apporder = append(appOrdInsStr.Apporder, eachApp.Metadata.Name)
		appDepStr.AppDepMap[eachApp.Metadata.Name] = "go"

		sortedTemplates, hookList, err := GetSortedTemplateForApp(eachApp.Metadata.Name, i.project, i.compositeApp, i.compAppVersion, rName, cp, namespace, overrideValues)

		if err != nil {
			log.Error("Unable to get the sorted templates for app", log.Fields{"AppName": eachApp.Metadata.Name})
			return pkgerrors.Wrap(err, "Unable to get the sorted templates for app")
		}

		log.Info(":: Resolved all the templates ::", log.Fields{"appName": eachApp.Metadata.Name, "SortedTemplate": sortedTemplates})

		defer cleanTmpfiles(sortedTemplates)
		// Read app dependency, if err continue
		appDep, _ := NewAppDependencyClient().GetAllSpecAppDependency(i.project, i.compositeApp, i.compAppVersion, eachApp.Metadata.Name)

		specData, err := NewAppIntentClient().GetAllIntentsByApp(eachApp.Metadata.Name, i.project, i.compositeApp, i.compAppVersion, gIntent, i.deploymentIntent)
		if err != nil {
			return pkgerrors.Wrap(err, "Unable to get the intents for app")
		}

		// listOfClusters shall have both mandatoryClusters and optionalClusters where the app needs to be installed.
		listOfClusters, err := gpic.IntentResolver(specData.Intent)
		if err != nil {
			return pkgerrors.Wrap(err, "Unable to get the intents resolved for app")
		}

		log.Info(":: listOfClusters ::", log.Fields{"listOfClusters": listOfClusters})
		if listOfClusters.MandatoryClusters == nil && listOfClusters.OptionalClusters == nil {
			log.Error("No compatible clusters have been provided to the Deployment Intent Group", log.Fields{"listOfClusters": listOfClusters})
			return pkgerrors.New("No compatible clusters have been provided to the Deployment Intent Group")
		}

		if err := checkClusters(listOfClusters, dcmClusters); err != nil {
			return err
		}

		//BEGIN: storing into etcd
		// Add an app to the app context
		ah := AppHandler {
			appName: eachApp.Metadata.Name,
			clusters: listOfClusters,
			namespace: namespace,
			ht: sortedTemplates,
			hk: hookList,
			dependency: appDep,
		}
		err = ah.addAppToAppContext(cxtForCApp)
		if err != nil {
			return pkgerrors.Wrap(err, "Error adding app to appContext: ")
		}
		err = ah.verifyResources(cxtForCApp)
		if err != nil {
			return pkgerrors.Wrap(err, "Error while verifying resources in app: ")
		}
	}
	jappOrderInstr, err := json.Marshal(appOrdInsStr)
	if err != nil {
		return pkgerrors.Wrap(err, "Error marshalling app order instruction")
	}

	jappDepInstr, err := json.Marshal(appDepStr.AppDepMap)
	if err != nil {
		return pkgerrors.Wrap(err, "Error marshalling app dependency instruction")
	}
	_, err = ctx.AddInstruction(cxtForCApp.compositeAppHandle, "app", "order", string(jappOrderInstr))
	if err != nil {
		return pkgerrors.Wrap(err, "Error adding app dependency instruction")
	}
	_, err = ctx.AddInstruction(cxtForCApp.compositeAppHandle, "app", "dependency", string(jappDepInstr))
	if err != nil {
		return pkgerrors.Wrap(err, "Error adding app dependency instruction")
	}
	//END: storing into etcd

	return nil
}
