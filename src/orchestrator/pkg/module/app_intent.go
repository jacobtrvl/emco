// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

/*
This file deals with the backend implementation of
Adding/Querying AppIntents for each application in the composite-app
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	pkgerrors "github.com/pkg/errors"

	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/gpic"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/utils"
)

// AppIntent has two components - metadata, spec
type AppIntent struct {
	MetaData MetaData `json:"metadata,omitempty"`
	Spec     SpecData `json:"spec,omitempty"`
}

// MetaData has - name, description, userdata1, userdata2
type MetaData struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	UserData1   string `json:"userData1,omitempty"`
	UserData2   string `json:"userData2,omitempty"`
}

// SpecData consists of appName and intent
type SpecData struct {
	AppName string           `json:"app,omitempty"`
	Intent  gpic.IntentStruc `json:"intent,omitempty"`
}

// AppIntentManager is an interface which exposes the
// AppIntentManager functionalities
type AppIntentManager interface {
	CreateAppIntent(ctx context.Context, a AppIntent, p string, ca string, v string, i string, digName string, failIfExists bool) (AppIntent, bool, error)
	GetAppIntent(ctx context.Context, ai string, p string, ca string, v string, i string, digName string) (AppIntent, error)
	GetAllIntentsByApp(ctx context.Context, aN, p, ca, v, i, digName string) (SpecData, error)
	GetAllAppIntents(ctx context.Context, p, ca, v, i, digName string) ([]AppIntent, error)
	DeleteAppIntent(ctx context.Context, ai string, p string, ca string, v string, i string, digName string) error
}

// AppIntentQueryKey required for query
type AppIntentQueryKey struct {
	AppName string `json:"app"`
}

// AppIntentKey is used as primary key
type AppIntentKey struct {
	Name                      string `json:"genericAppPlacementIntent"`
	Project                   string `json:"project"`
	CompositeApp              string `json:"compositeApp"`
	Version                   string `json:"compositeAppVersion"`
	Intent                    string `json:"genericPlacementIntent"`
	DeploymentIntentGroupName string `json:"deploymentIntentGroup"`
}

// AppIntentFindByAppKey required for query
type AppIntentFindByAppKey struct {
	Project                   string `json:"project"`
	CompositeApp              string `json:"compositeApp"`
	CompositeAppVersion       string `json:"compositeAppVersion"`
	Intent                    string `json:"genericPlacementIntent"`
	DeploymentIntentGroupName string `json:"deploymentIntentGroup"`
	AppName                   string `json:"app"`
}

// ApplicationsAndClusterInfo type represents the list of
type ApplicationsAndClusterInfo struct {
	ArrayOfAppClusterInfo []AppClusterInfo `json:"applications"`
}

// AppClusterInfo is a type linking the app and the clusters
// on which they need to be installed.
type AppClusterInfo struct {
	Name       string       `json:"name"`
	AllOfArray []gpic.AllOf `json:"allOf,omitempty"`
	AnyOfArray []gpic.AnyOf `json:"anyOf,omitempty"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (ak AppIntentKey) String() string {
	out, err := json.Marshal(ak)
	if err != nil {
		return ""
	}
	return string(out)
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (ai AppIntentFindByAppKey) String() string {
	out, err := json.Marshal(ai)
	if err != nil {
		return ""
	}
	return string(out)
}

type IntentSelectorHandler interface {
	Handle(ctx context.Context, appIntent *AppIntent, digName, project, contextApp, version string) error
}

type intentSelectorHandler struct {
}

func NewIntentSelectorHandler() IntentSelectorHandler {
	return &intentSelectorHandler{}
}

// AppIntentClient implements the AppIntentManager interface
type AppIntentClient struct {
	storeName        string
	tagMetaData      string
	selectorsHandler IntentSelectorHandler
}

// NewAppIntentClient returns an instance of AppIntentClient
func NewAppIntentClient() *AppIntentClient {
	return &AppIntentClient{
		storeName:        "resources",
		tagMetaData:      "data",
		selectorsHandler: NewIntentSelectorHandler(),
	}
}

// CreateAppIntent creates an entry for AppIntent in the db.
// Other input parameters for it - projectName, compositeAppName, version, intentName and deploymentIntentGroupName.
// failIfExists - indicates the request is POST=true or PUT=false
func (c *AppIntentClient) CreateAppIntent(ctx context.Context, a AppIntent, p string, ca string, v string, i string, digName string, failIfExists bool) (AppIntent, bool, error) {
	aiExists := false

	//Check for the AppIntent already exists here.
	res, err := c.GetAppIntent(ctx, a.MetaData.Name, p, ca, v, i, digName)
	if err == nil && !reflect.DeepEqual(res, AppIntent{}) {
		aiExists = true
	}

	if aiExists && failIfExists {
		return AppIntent{}, aiExists, pkgerrors.New("AppIntent already exists")
	}

	err = c.selectorsHandler.Handle(ctx, &a, digName, p, ca, v)
	if err != nil {
		return AppIntent{}, aiExists, err
	}

	akey := AppIntentKey{
		Name:                      a.MetaData.Name,
		Project:                   p,
		CompositeApp:              ca,
		Version:                   v,
		Intent:                    i,
		DeploymentIntentGroupName: digName,
	}

	qkey := AppIntentQueryKey{
		AppName: a.Spec.AppName,
	}

	err = db.DBconn.Insert(ctx, c.storeName, akey, qkey, c.tagMetaData, a)
	if err != nil {
		return AppIntent{}, aiExists, err
	}

	return a, aiExists, nil
}

// GetAppIntent shall take arguments - name of the app intent, name of the project, name of the composite app, version of the composite app,intent name and deploymentIntentGroupName. It shall return the AppIntent
func (c *AppIntentClient) GetAppIntent(ctx context.Context, ai string, p string, ca string, v string, i string, digName string) (AppIntent, error) {

	k := AppIntentKey{
		Name:                      ai,
		Project:                   p,
		CompositeApp:              ca,
		Version:                   v,
		Intent:                    i,
		DeploymentIntentGroupName: digName,
	}

	result, err := db.DBconn.Find(ctx, c.storeName, k, c.tagMetaData)
	if err != nil {
		return AppIntent{}, err
	}

	if len(result) == 0 {
		return AppIntent{}, pkgerrors.New("AppIntent not found")
	}

	if result != nil {
		a := AppIntent{}
		err = db.DBconn.Unmarshal(result[0], &a)
		if err != nil {
			return AppIntent{}, err
		}
		return a, nil

	}
	return AppIntent{}, pkgerrors.New("Unknown Error")
}

/*
GetAllIntentsByApp queries intent by AppName, it takes in parameters AppName, CompositeAppName, CompositeNameVersion,
GenericPlacementIntentName & DeploymentIntentGroupName. Returns SpecData which contains
all the intents for the app.
*/
func (c *AppIntentClient) GetAllIntentsByApp(ctx context.Context, aN, p, ca, v, i, digName string) (SpecData, error) {
	k := AppIntentFindByAppKey{
		Project:                   p,
		CompositeApp:              ca,
		CompositeAppVersion:       v,
		Intent:                    i,
		DeploymentIntentGroupName: digName,
		AppName:                   aN,
	}
	result, err := db.DBconn.Find(ctx, c.storeName, k, c.tagMetaData)
	if err != nil {
		return SpecData{}, err
	}
	if len(result) == 0 {
		return SpecData{}, nil
	}

	var a AppIntent
	err = db.DBconn.Unmarshal(result[0], &a)
	if err != nil {
		return SpecData{}, err
	}
	return a.Spec, nil

}

/*
GetAllAppIntents takes in paramaters ProjectName, CompositeAppName, CompositeNameVersion
and GenericPlacementIntentName,DeploymentIntentGroupName. Returns an array of AppIntents
*/
func (c *AppIntentClient) GetAllAppIntents(ctx context.Context, p, ca, v, i, digName string) ([]AppIntent, error) {
	k := AppIntentKey{
		Name:                      "",
		Project:                   p,
		CompositeApp:              ca,
		Version:                   v,
		Intent:                    i,
		DeploymentIntentGroupName: digName,
	}
	result, err := db.DBconn.Find(ctx, c.storeName, k, c.tagMetaData)
	if err != nil {
		return []AppIntent{}, err
	}

	var appIntents []AppIntent

	if len(result) != 0 {
		for i := range result {
			aI := AppIntent{}
			err = db.DBconn.Unmarshal(result[i], &aI)
			if err != nil {
				return []AppIntent{}, err
			}
			appIntents = append(appIntents, aI)
		}
	}

	return appIntents, err
}

// DeleteAppIntent delete an AppIntent
func (c *AppIntentClient) DeleteAppIntent(ctx context.Context, ai string, p string, ca string, v string, i string, digName string) error {
	k := AppIntentKey{
		Name:                      ai,
		Project:                   p,
		CompositeApp:              ca,
		Version:                   v,
		Intent:                    i,
		DeploymentIntentGroupName: digName,
	}

	err := db.DBconn.Remove(ctx, c.storeName, k)
	return err

}

func (c *AppIntentClient) CloneAppIntents(ctx context.Context, p string, ca string, v string, i string, di string, tDi string) ([]AppIntent, error) {
	intents, err := c.GetAllAppIntents(ctx, p, ca, v, i, di)
	if err != nil {
		return nil, err
	}

	for _, intent := range intents {
		if _, _, err := c.CreateAppIntent(ctx, intent, p, ca, v, i, tDi, true); err != nil {
			return nil, err
		}
	}

	return intents, nil
}

func (i *intentSelectorHandler) Handle(ctx context.Context, appIntent *AppIntent, digName, project, contextApp, version string) error {

	dig, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(ctx, digName, project, contextApp, version)
	if err != nil {
		return err
	}

	lcCluster, err := NewLogicalCloudClient().Get(ctx, project, dig.Spec.LogicalCloud)
	if err != nil {
		return err
	}

	dcmClusters, err := NewClusterClient().GetAllClusters(ctx, project, lcCluster.MetaData.Name)
	if err != nil {
		return err
	}

	appIntent.Spec.Intent.Selector = gpic.NameClusterSelector
	if i.isLabelSelected(appIntent) {
		appIntent.Spec.Intent.Selector = gpic.LabelClusterSelector
	}

	allOfSelector, err := i.handleAllOfSelectors(appIntent.Spec.Intent.AllOfArray, dcmClusters)
	if err != nil {
		return err
	}

	anyOfSelector, err := i.handleAnyOfSelectors(appIntent.Spec.Intent.AnyOfArray, dcmClusters)
	if err != nil {
		return err
	}

	appIntent.Spec.Intent.AllOfArray = allOfSelector
	appIntent.Spec.Intent.AnyOfArray = anyOfSelector

	return nil
}

func (i *intentSelectorHandler) handleAllOfSelectors(selectorList []gpic.AllOf, allowedClusters []common.Cluster) ([]gpic.AllOf, error) {
	var anyOfList []gpic.AnyOf
	err := utils.ConvertType(selectorList, &anyOfList)
	if err != nil {
		return nil, err
	}

	anyOfSelectedList, err := i.handleAnyOfSelectors(anyOfList, allowedClusters)
	if err != nil {
		return nil, err
	}

	var allOfSelectedList []gpic.AllOf
	err = utils.ConvertType(anyOfSelectedList, &allOfSelectedList)
	if err != nil {
		return nil, err
	}

	return allOfSelectedList, nil
}

func (i *intentSelectorHandler) handleAnyOfSelectors(selectorList []gpic.AnyOf, allowedClusters []common.Cluster) ([]gpic.AnyOf, error) {
	var selectedList []gpic.AnyOf

	if selectorList == nil || len(selectorList) == 0 {
		return selectedList, nil
	}

	for _, selector := range selectorList {
		if selector.ProviderName == "" {
			return nil, fmt.Errorf("\"clusterProvider\" is required")
		}

		if selector.ClusterName == "" && selector.ClusterLabelName == "" {
			return nil, fmt.Errorf("no \"clusterName\" or \"clusterLabel\" found")
		}

		if selector.ClusterName != "" {
			if !utils.ContainCluster(selector.ProviderName, selector.ClusterName, allowedClusters) {
				return nil, fmt.Errorf("cluster \"%s\" is not part of DIG's logical cluster", selector.ClusterName)
			}

			selectedList = append(selectedList, selector)
			continue
		}

		// Provided clusterLabel but missing clusterName
		// need to add clusterName referential integrity
		clusters, err := cluster.NewClusterClient().GetClustersWithLabel(context.Background(),
			selector.ProviderName, selector.ClusterLabelName)
		if err != nil {
			return nil, err
		}

		if clusters == nil || len(clusters) == 0 {
			return nil, fmt.Errorf("no cluster found in cluster provider \"%s\" with label \"%s\"",
				selector.ProviderName, selector.ClusterLabelName)
		}

		found := false
		for _, clusterName := range clusters {
			if !utils.ContainCluster(selector.ProviderName, clusterName, allowedClusters) {
				continue
			}

			found = true
			selectedList = append(selectedList, gpic.AnyOf{
				ProviderName:     selector.ProviderName,
				ClusterName:      clusterName,
				ClusterLabelName: selector.ClusterLabelName,
			})
		}

		if !found {
			return nil, fmt.Errorf("no cluster with label \"%s\" found in DIG logical cluster", selector.ClusterLabelName)
		}

	}

	return selectedList, nil
}

func (i *intentSelectorHandler) isLabelSelected(appIntent *AppIntent) bool {
	for _, selector := range appIntent.Spec.Intent.AllOfArray {
		if selector.ClusterLabelName != "" {
			return true
		}
	}

	for _, selector := range appIntent.Spec.Intent.AnyOfArray {
		if selector.ClusterLabelName != "" {
			return true
		}
	}

	return false
}
