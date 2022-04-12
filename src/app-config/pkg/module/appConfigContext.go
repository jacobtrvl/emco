// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	rsyncclient "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/installappclient"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	orchmodule "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/utils"
	yaml "gopkg.in/yaml.v3"
)

type acInstantiator struct {
	project          string
	compositeApp     string
	compAppVersion   string
	deploymentIntent string
}

type ContextForAppConfig struct {
	context         appcontext.AppContext
	ctxval          interface{}
	appConfigHandle interface{}
}

const SEPARATOR = "+"

// MakeAppContext shall make an app context and store the app context into etcd. This shall return contextForCompositeApp
func (i *acInstantiator) MakeAppConfigContext(appcfg string) (ContextForAppConfig, error) {

	cca, err := i.makeAppConfigContext()
	if err != nil {
		return ContextForAppConfig{}, err
	}

	err = i.storeAppConfigIntoRunTimeDB(cca, appcfg)
	if err != nil {
		deleteAppContext(cca.context)
		return ContextForAppConfig{}, pkgerrors.Wrap(err, "Error in storeAppConfigContextIntoETCd")
	}
	err = rsyncInstall(cca.ctxval, true)
	if err != nil {
		deleteAppContext(cca.context)
		return ContextForAppConfig{}, pkgerrors.Wrap(err, "Error in rsyncInstall")
	}

	return cca, nil
}

func SyncTerminateAppConfig(appCfgCtx string) error {
	_, err := QueryDBAndSetRsyncInfo()
	if err != nil {
		return pkgerrors.Wrap(err, "Failed to get rsync info")
	}
	appContextID := appCfgCtx
	err = rsyncclient.InvokeUninstallApp(appContextID)
	if err != nil {
		return pkgerrors.Wrap(err, "Failed to uninstall app")
	}
	return nil
}

func (i *acInstantiator) makeAppConfigContext() (ContextForAppConfig, error) {
	context := appcontext.AppContext{}
	ctxval, err := context.InitAppContext()
	if err != nil {
		return ContextForAppConfig{}, pkgerrors.Wrap(err, "Error creating AppContext AppConfig")
	}
	compositeHandle, err := context.CreateCompositeApp()
	if err != nil {
		return ContextForAppConfig{}, pkgerrors.Wrap(err, "Error creating AppConfig handle")
	}
	err = context.AddCompositeAppMeta(appcontext.CompositeAppMeta{
		Project:               i.project,
		CompositeApp:          i.compositeApp,
		Version:               i.compAppVersion,
		DeploymentIntentGroup: i.deploymentIntent,
		Namespace:             "default"})
	if err != nil {
		return ContextForAppConfig{}, pkgerrors.Wrap(err, "Error Adding CompositeAppConfigMeta")
	}

	m, _ := context.GetCompositeAppMeta()
	log.Info(":: The meta data stored in the runtime context :: ", log.Fields{"Project": m.Project, "CompositeApp": m.CompositeApp, "Version": m.Version, "DeploymentIntentGroup": m.DeploymentIntentGroup})

	cca := ContextForAppConfig{context: context, ctxval: ctxval, appConfigHandle: compositeHandle}
	return cca, nil

}

type appOrderInstr struct {
	Apporder []string `json:"apporder"`
}

type appDepInstr struct {
	AppDepMap map[string]string `json:"appdependency"`
}

func (i *acInstantiator) GetAppForAppConfig(appcfg string) (string, error) {
	appConfig, err := NewAppConfigClient().GetAppConfig(appcfg, i.project, i.compositeApp, i.compAppVersion, i.deploymentIntent)
	if err != nil {
		return "", err
	}
	cr, err := NewResTypeClient().GetResType(appConfig.Spec.ResType, i.project, i.compositeApp, i.compAppVersion)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error in getting restype for appconfig "+appConfig.Spec.ResType)
	}
	return cr.Spec.AppName, nil
}

func (i *acInstantiator) GetAppContextIDForParentApp() (string, error) {
	s, err := orchmodule.NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(i.deploymentIntent, i.project, i.compositeApp, i.compAppVersion)
	if err != nil {
		return "", err
	}
	return state.GetLastContextIdFromStateInfo(s), nil
}

func (i *acInstantiator) storeAppConfigIntoRunTimeDB(cxtForCApp ContextForAppConfig, appcfg string) error {

	var appOrdInsStr appOrderInstr
	var appDepStr appDepInstr
	appDepStr.AppDepMap = make(map[string]string)

	ct := cxtForCApp.context
	appConfig, err := NewAppConfigClient().GetAppConfig(appcfg, i.project, i.compositeApp, i.compAppVersion, i.deploymentIntent)
	if err != nil {
		return pkgerrors.Wrap(err, "Not finding the appConfig")
	}
	aC, err := NewAppConfigClient().GetAppConfigContent(appcfg, i.project, i.compositeApp, i.compAppVersion, i.deploymentIntent)
	if err != nil {
		return pkgerrors.Wrap(err, "Not finding the appConfig Content")
	}
	appContent, err := base64.StdEncoding.DecodeString(aC.FileContents[0])
	if err != nil {
		return pkgerrors.Wrap(err, "Fail to convert to byte array")
	}
	parentApp, err := i.GetAppForAppConfig(appcfg)
	if err != nil {
		return pkgerrors.Wrap(err, "Cannot find the parent application for appcfg")
	}
	appName := fmt.Sprintf("%s-%s", parentApp, appcfg)
	appHandle, err := ct.AddApp(cxtForCApp.appConfigHandle, appName)
	if err != nil {
		return pkgerrors.Wrap(err, "Cannot create AppConfig Handle")
	}

	var cSpecific, cScope, cProvider, cName, cLabel string
	if strings.ToLower(appConfig.Spec.ClusterSpecific) == "true" && (ClusterInfo{}) != appConfig.Spec.ClusterInfo {
		cSpecific = strings.ToLower(appConfig.Spec.ClusterSpecific)
		cScope = strings.ToLower(appConfig.Spec.ClusterInfo.Scope)
		cProvider = appConfig.Spec.ClusterInfo.ClusterProvider
		cName = appConfig.Spec.ClusterInfo.ClusterName
		cLabel = appConfig.Spec.ClusterInfo.ClusterLabel
	}
	pacID, err := i.GetAppContextIDForParentApp()
	if err != nil {
		return pkgerrors.Wrap(err, "Failed to get the parent app context ID")
	}
	var pac appcontext.AppContext
	_, err = pac.LoadAppContext(pacID)
	if err != nil {
		return pkgerrors.Errorf("Internal error")
	}
	clusters, err := pac.GetClusterNames(parentApp)
	if err != nil {
		return pkgerrors.Errorf("Internal error")
	}

	appOrdInsStr.Apporder = append(appOrdInsStr.Apporder, appName)
	appDepStr.AppDepMap[appName] = "go"
	for _, c := range clusters {
		if cSpecific == "true" && cScope == "label" {
			allow, err := isValidClusterToApplyByLabel(cProvider, c, cLabel)
			if err != nil {
				log.Error("Error ApplyToClusterByLabel", log.Fields{"Provider": cProvider, "ClusterName": cName, "ClusterLabel": cLabel})
				return pkgerrors.Errorf("Internal error")
			}
			if !allow {
				continue
			}
		}
		if cSpecific == "true" && cScope == "name" {
			allow, err := isValidClusterToApplyByName(cProvider, c, cName)
			if err != nil {
				log.Error("Error ApplyClusterByName", log.Fields{"Provider": cProvider, "GivenClusterName": cName, "AutheticatingForCluste": c})
				return pkgerrors.Errorf("Internal error")
			}
			if !allow {
				continue
			}
		}
		p := appConfig.Spec.ClusterInfo.ClusterProvider
		clusterhandle, err := ct.AddCluster(appHandle, c)
		if err != nil {
			return pkgerrors.Wrapf(err, "Error adding Resources to Cluster(provider::%s, name::%s) to AppContext", p, c)
		}
		var yamlStruct utils.YamlStruct
		err = yaml.Unmarshal(appContent, &yamlStruct)
		if err != nil {
			return pkgerrors.Wrap(err, "Cant unmarshal yaml file ..")
		}
		name := yamlStruct.Metadata.Name + SEPARATOR + yamlStruct.Kind
		if name == SEPARATOR {
			log.Warn(":: Ignoring, Unable to render the app config template ::", log.Fields{})
			continue
		}
		var resOrderInstr struct {
			Resorder []string `json:"resorder"`
		}
		resOrderInstr.Resorder = append(resOrderInstr.Resorder, name)
		_, err = ct.AddResource(clusterhandle, name, string(appContent))
		if err != nil {
			return pkgerrors.Wrapf(err, "Error adding resource ::%s to AppContext", name)
		}
		jresOrderInstr, _ := json.Marshal(resOrderInstr)
		_, err = ct.AddInstruction(clusterhandle, "resource", "order", string(jresOrderInstr))
		if err != nil {
			return pkgerrors.Wrapf(err, "Error adding instruction for resource order")
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
	_, err = ct.AddInstruction(cxtForCApp.appConfigHandle, "app", "order", string(jappOrderInstr))
	if err != nil {
		return pkgerrors.Wrap(err, "Error adding app dependency instruction")
	}
	_, err = ct.AddInstruction(cxtForCApp.appConfigHandle, "app", "dependency", string(jappDepInstr))
	if err != nil {
		return pkgerrors.Wrap(err, "Error adding app dependency instruction")
	}
	//END: storing into etcd

	return nil
}

func rsyncInstall(ctxval interface{}, action bool) error {
	_, err := QueryDBAndSetRsyncInfo()
	if err != nil {
		return pkgerrors.Wrap(err, "Failed to get rsync info")
	}
	appContextID := fmt.Sprintf("%v", ctxval)
	if action == true {
		err = rsyncclient.InvokeInstallApp(appContextID)
	} else if action == false {
		err = rsyncclient.InvokeUninstallApp(appContextID)
	} else {
		return pkgerrors.Errorf("Unknown rsync action")
	}
	if err != nil {
		return pkgerrors.Wrap(err, "Failed to install app")
	}

	return nil
}

func QueryDBAndSetRsyncInfo() (rsyncclient.RsyncInfo, error) {
	client := controller.NewControllerClient("resources", "data", "orchestrator")
	vals, _ := client.GetControllers()
	for _, v := range vals {
		if v.Metadata.Name == "rsync" {
			rsyncInfo := rsyncclient.NewRsyncInfo(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
			return rsyncInfo, nil
		}
	}
	return rsyncclient.RsyncInfo{}, pkgerrors.Errorf("queryRsyncInfoInMCODB Failed - Could not get find rsync by name : rsync")
}

func deleteAppContext(ac appcontext.AppContext) {
	err := ac.DeleteCompositeApp()
	if err != nil {
		log.Error(":: Error deleting AppContext ::", log.Fields{})
	}
}
