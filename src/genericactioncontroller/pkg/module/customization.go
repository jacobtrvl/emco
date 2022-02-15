package module

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// Customization consists of metadata and Spec
type Customization struct {
	Metadata Metadata          `json:"metadata"`
	Spec     CustomizationSpec `json:"spec"`
}

// CustomizationSpec consists of ClusterSpecific and ClusterInfo
type CustomizationSpec struct {
	ClusterSpecific  string                   `json:"clusterSpecific"`
	ClusterInfo      ClusterInfo              `json:"clusterInfo"`
	PatchType        string                   `json:"patchType,omitempty"`
	PatchJSON        []map[string]interface{} `json:"patchJson,omitempty"`
	ConfigMapOptions ConfigMapOptions         `json:"configMapOptions,omitempty"`
	SecretOptions    SecretOptions            `json:"secretOptions,omitempty"`
}

// ClusterInfo consists of scope, Clusterprovider, ClusterName, ClusterLabel and Mode
type ClusterInfo struct {
	Scope           string `json:"scope"`
	ClusterProvider string `json:"clusterProvider"`
	ClusterName     string `json:"cluster"`
	ClusterLabel    string `json:"clusterLabel"`
	Mode            string `json:"mode"`
}

// ConfigMapOptions holds properties for customizing configmap
type ConfigMapOptions struct {
	DataKeyOptions []KeyOptions `json:"dataKeyOptions,omitempty"`
}

// SecretOptions holds properties for customizing secret
type SecretOptions struct {
	DataKeyOptions []KeyOptions `json:"dataKeyOptions,omitempty"`
}

// KeyOptions holds properties for customizing configmap/secret data keys
type KeyOptions struct {
	FileName string `json:"fileName"`
	KeyName  string `json:"keyName"`
}

// Content represent either a merge content or a JSON patch
// and its targets. The content of the patch can either be from a file
// or from an inline string.
type Content struct {
	FileName string
	Content  string
	KeyName  string
}

// CustomizationContent is a list of patches, where each one can be either a
// merge content or a JSON patch.
type CustomizationContent struct {
	Content []Content
}

// CustomizationKey consists of CustomizationName, project, CompApp, CompAppVersion,
// DeploymentIntentGroupName, GenericIntentName, ResourceName
type CustomizationKey struct {
	Customization         string `json:"customization"`
	Project               string `json:"project"`
	CompositeApp          string `json:"compositeApp"`
	CompositeAppVersion   string `json:"compositeAppVersion"`
	DeploymentIntentGroup string `json:"deploymentIntentGroup"`
	GenericK8sIntent      string `json:"genericK8sIntent"`
	Resource              string `json:"genericResource"`
}

// CustomizationManager exposes all the functionalities of customization
type CustomizationManager interface {
	CreateCustomization(customization Customization, content CustomizationContent,
		project, compositeApp, version, deploymentIntentGroup, intent, resource string,
		failIfExists bool) (Customization, bool, error)
	DeleteCustomization(customization, project, compositeApp, version, deploymentIntentGroup, intent, resource string) error
	GetAllCustomization(project, compositeApp, version, deploymentIntentGroup, intent, resource string) ([]Customization, error)
	GetCustomization(customization, project, compositeApp, version, deploymentIntentGroup, intent, resource string) (Customization, error)
	GetCustomizationContent(customization, project, compositeApp, version, deploymentIntentGroup, intent, resource string) (CustomizationContent, error)
}

// CustomizationClientDbInfo consists of tableName and columns
type CustomizationClientDbInfo struct {
	storeName  string // name of the mongodb collection to use for client documents
	tagMeta    string // attribute key name for the json data of a client document
	tagContent string // attribute key name for the file data of a client document
}

// CustomizationClient consists of CustomizationClientDbInfo
type CustomizationClient struct {
	db CustomizationClientDbInfo
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (k CustomizationKey) String() string {
	out, err := json.Marshal(k)
	if err != nil {
		return ""
	}
	return string(out)
}

// NewCustomizationClient returns an instance of the CustomizationClient
func NewCustomizationClient() *CustomizationClient {
	return &CustomizationClient{
		db: CustomizationClientDbInfo{
			storeName:  "resources",
			tagMeta:    "data",
			tagContent: "customizationcontent",
		},
	}
}

// CreateCustomization creates a new Customization
func (cc *CustomizationClient) CreateCustomization(customization Customization, content CustomizationContent,
	project, compositeApp, version, deploymentIntentGroup, intent, resource string,
	failIfExists bool) (Customization, bool, error) {

	cExists := false
	key := CustomizationKey{
		Customization:         customization.Metadata.Name,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   version,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      intent,
		Resource:              resource,
	}

	c, err := cc.GetCustomization(
		customization.Metadata.Name, project, compositeApp, version, deploymentIntentGroup, intent, resource)
	if err == nil &&
		!reflect.DeepEqual(c, Customization{}) {
		cExists = true
	}

	if cExists &&
		failIfExists {
		return Customization{}, cExists, errors.New("Customization already exists")
	}

	if err = db.DBconn.Insert(cc.db.storeName, key, nil, cc.db.tagMeta, customization); err != nil {
		return Customization{}, cExists, err
	}

	if !reflect.DeepEqual(content, CustomizationContent{}) {
		if err = db.DBconn.Insert(cc.db.storeName, key, nil, cc.db.tagContent, content); err != nil {
			return Customization{}, cExists, err
		}
	}

	return customization, cExists, nil
}

// GetCustomization returns Customization
func (cc *CustomizationClient) GetCustomization(
	customization, project, compositeApp, version, deploymentIntentGroup, intent, resource string) (Customization, error) {

	key := CustomizationKey{
		Customization:         customization,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   version,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      intent,
		Resource:              resource,
	}

	value, err := db.DBconn.Find(cc.db.storeName, key, cc.db.tagMeta)
	if err != nil {
		return Customization{}, err
	}

	if len(value) == 0 {
		return Customization{}, errors.New("Customization not found")
	}

	if value != nil {
		c := Customization{}
		if err = db.DBconn.Unmarshal(value[0], &c); err != nil {
			return Customization{}, err
		}
		return c, nil
	}

	return Customization{}, errors.New("Unknown Error")
}

// GetAllCustomization returns all the customization objects
func (cc *CustomizationClient) GetAllCustomization(
	project, compositeApp, version, deploymentIntentGroup, intent, resource string) ([]Customization, error) {

	key := CustomizationKey{
		Customization:         "",
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   version,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      intent,
		Resource:              resource,
	}

	values, err := db.DBconn.Find(cc.db.storeName, key, cc.db.tagMeta)
	if err != nil {
		return []Customization{}, err
	}

	var customizations []Customization
	for _, value := range values {
		c := Customization{}
		if err = db.DBconn.Unmarshal(value, &c); err != nil {
			return []Customization{}, err
		}
		customizations = append(customizations, c)
	}

	return customizations, nil
}

// GetCustomizationContent returns the customizationContent
func (cc *CustomizationClient) GetCustomizationContent(
	customization, project, compositeApp, version, deploymentIntentGroup, intent, resource string) (CustomizationContent, error) {

	key := CustomizationKey{
		Customization:         customization,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   version,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      intent,
		Resource:              resource,
	}

	value, err := db.DBconn.Find(cc.db.storeName, key, cc.db.tagContent)
	if err != nil {
		return CustomizationContent{}, err
	}

	if len(value) > 0 &&
		value[0] != nil {
		c := CustomizationContent{}
		if err = db.DBconn.Unmarshal(value[0], &c); err != nil {
			return CustomizationContent{}, err
		}
		return c, nil
	}

	return CustomizationContent{}, nil
}

// DeleteCustomization deletes Customization
func (cc *CustomizationClient) DeleteCustomization(
	customization, project, compositeApp, version, deploymentIntentGroup, intent, resource string) error {

	key := CustomizationKey{
		Customization:         customization,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   version,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      intent,
		Resource:              resource,
	}

	return db.DBconn.Remove(cc.db.storeName, key)
}
