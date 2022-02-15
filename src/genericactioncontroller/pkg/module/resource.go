package module

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"encoding/json"
	"reflect"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// Resource consists of metadata and Spec
type Resource struct {
	Metadata Metadata     `json:"metadata"`
	Spec     ResourceSpec `json:"spec"`
}

// ResourceSpec consists of AppName, NewObject, ResourceGVK
type ResourceSpec struct {
	AppName     string      `json:"app"`
	NewObject   string      `json:"newObject"`
	ResourceGVK ResourceGVK `json:"resourceGVK,omitempty"`
}

// ResourceGVK consists of ApiVersion, Kind, Name
type ResourceGVK struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

// ResourceContent contains the content of resource template
type ResourceContent struct {
	Content string `json:"filecontent"`
}

// ResourceKey consists of resourceName, ProjectName, CompAppName, CompAppVersion,
// DeploymentIntentgroupName, GenericK8sIntentName
type ResourceKey struct {
	Resource              string `json:"genericResource"`
	Project               string `json:"project"`
	CompositeApp          string `json:"compositeApp"`
	CompositeAppVersion   string `json:"compositeAppVersion"`
	DeploymentIntentGroup string `json:"deploymentIntentGroup"`
	GenericK8sIntent      string `json:"genericK8sIntent"`
}

// ResourceManager is an interface that exposes resource related functionalities
type ResourceManager interface {
	CreateResource(res Resource, resContent ResourceContent,
		project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string,
		failIfExists bool) (Resource, bool, error)
	DeleteResource(resource, project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string) error
	GetAllResources(project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string) ([]Resource, error)
	GetResource(resource, project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string) (Resource, error)
	GetResourceContent(resource, project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string) (ResourceContent, error)
}

type clientDbInfo struct {
	storeName  string // name of the mongodb collection to use for client documents
	tagMeta    string // attribute key name for the json data of a client document
	tagContent string // attribute key name for the file data of a client document
}

// ResourceClient implements the resourceManager
type ResourceClient struct {
	db clientDbInfo
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (k ResourceKey) String() string {
	out, err := json.Marshal(k)
	if err != nil {
		return ""
	}
	return string(out)
}

// NewResourceClient returns an instance of the ResourceClient
// which implements the Manager
func NewResourceClient() *ResourceClient {
	return &ResourceClient{
		db: clientDbInfo{
			storeName:  "resources",
			tagMeta:    "data",
			tagContent: "resourcecontent",
		},
	}
}

// CreateResource creates a resource
func (rc *ResourceClient) CreateResource(res Resource, content ResourceContent,
	project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string,
	failIfExists bool) (Resource, bool, error) {

	rExists := false
	key := ResourceKey{
		Resource:              res.Metadata.Name,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   compositeAppVersion,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      genericK8sIntent,
	}

	r, err := rc.GetResource(res.Metadata.Name, project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent)
	if err == nil &&
		!reflect.DeepEqual(r, Resource{}) {
		rExists = true
	}

	if rExists &&
		failIfExists {
		return Resource{}, rExists, pkgerrors.New("Resource already exists")
	}

	if err = db.DBconn.Insert(rc.db.storeName, key, nil, rc.db.tagMeta, res); err != nil {
		return Resource{}, rExists, err
	}

	if !reflect.DeepEqual(content, ResourceContent{}) {
		if err = db.DBconn.Insert(rc.db.storeName, key, nil, rc.db.tagContent, content); err != nil {
			return Resource{}, rExists, err
		}
	}

	return res, rExists, nil
}

// GetResource returns a resource
func (rc *ResourceClient) GetResource(resource, project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string) (Resource, error) {

	key := ResourceKey{
		Resource:              resource,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   compositeAppVersion,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      genericK8sIntent,
	}

	value, err := db.DBconn.Find(rc.db.storeName, key, rc.db.tagMeta)
	if err != nil {
		return Resource{}, err
	}

	if len(value) == 0 {
		return Resource{}, pkgerrors.New("Resource not found")
	}

	if value != nil {
		r := Resource{}
		if err = db.DBconn.Unmarshal(value[0], &r); err != nil {
			return Resource{}, err
		}
		return r, nil
	}

	return Resource{}, pkgerrors.New("Unknown Error")
}

// GetAllResources shall return all the resources for the intent
func (rc *ResourceClient) GetAllResources(project, compositeApp, compositeAppVersion, deploymentIntentGroup,
	genericK8sIntent string) ([]Resource, error) {

	key := ResourceKey{
		Resource:              "",
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   compositeAppVersion,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      genericK8sIntent,
	}

	values, err := db.DBconn.Find(rc.db.storeName, key, rc.db.tagMeta)
	if err != nil {
		return []Resource{}, err
	}

	var resources []Resource
	for _, value := range values {
		r := Resource{}
		if err = db.DBconn.Unmarshal(value, &r); err != nil {
			return []Resource{}, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

// GetResourceContent returns the content of the resource template
func (rc *ResourceClient) GetResourceContent(resource, project, compositeApp, compositeAppVersion,
	deploymentIntentGroup, genericK8sIntent string) (ResourceContent, error) {

	key := ResourceKey{
		Resource:              resource,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   compositeAppVersion,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      genericK8sIntent,
	}

	value, err := db.DBconn.Find(rc.db.storeName, key, rc.db.tagContent)
	if err != nil {
		return ResourceContent{}, err
	}

	if len(value) > 0 &&
		value[0] != nil {
		c := ResourceContent{}
		if err = db.DBconn.Unmarshal(value[0], &c); err != nil {
			return ResourceContent{}, err
		}
		return c, nil
	}

	return ResourceContent{}, nil
}

// DeleteResource deletes a given resource
func (rc *ResourceClient) DeleteResource(resource, project, compositeApp, compositeAppVersion,
	deploymentIntentGroup, genericK8sIntent string) error {

	key := ResourceKey{
		Resource:              resource,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   compositeAppVersion,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      genericK8sIntent,
	}

	return db.DBconn.Remove(rc.db.storeName, key)
}
