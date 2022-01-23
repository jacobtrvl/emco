package module

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"encoding/json"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// Customization consists of metadata and Spec
type Customization struct {
	Metadata Metadata      `json:"metadata"`
	Spec     CustomizeSpec `json:"spec"`
}

// CustomizeSpec consists of ClusterSpecific and ClusterInfo
type CustomizeSpec struct {
	ClusterSpecific string                   `json:"clusterSpecific"`
	ClusterInfo     ClusterInfo              `json:"clusterInfo"`
	PatchType       string                   `json:"patchType,omitempty"`
	PatchJSON       []map[string]interface{} `json:"patchJson,omitempty"`
	SpecFile        []SpecFile               `json:"specFile,omitempty"`
}

// ClusterInfo consists of scope, Clusterprovider, ClusterName, ClusterLabel and Mode
type ClusterInfo struct {
	Scope           string `json:"scope"`
	ClusterProvider string `json:"clusterProvider"`
	ClusterName     string `json:"cluster"`
	ClusterLabel    string `json:"clusterLabel"`
	Mode            string `json:"mode"`
}

type SpecFile struct {
	ConfigFile string `json:"configFile"`
	ConfigKey  string `json:"configKey"`
}

// SpecFileContent contains the array of file contents
type SpecFileContent struct {
	FileContents []string
	FileNames    []string
}

// CustomizationKey consists of CustomizationName, project, CompApp, CompAppVersion, DeploymentIntentGroupName, GenericIntentName, ResourceName
type CustomizationKey struct {
	Customization       string `json:"customization"`
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeApp"`
	CompositeAppVersion string `json:"compositeAppVersion"`
	DigName             string `json:"deploymentIntentGroup"`
	GenericK8sIntent    string `json:"genericK8sIntent"`
	Resource            string `json:"genericResource"`
}

// CustomizationManager exposes all the functionalities of customization
type CustomizationManager interface {
	CreateCustomization(c Customization, t SpecFileContent, p, ca, cv, dig, gi, rs string, exists bool) (Customization, error)
	GetCustomization(c, p, ca, cv, dig, gi, rs string) (Customization, error)
	GetCustomizationContent(c, p, ca, cv, dig, gi, rs string) (SpecFileContent, error)
	GetAllCustomization(p, ca, cv, dig, gi, rs string) ([]Customization, error)
	DeleteCustomization(c, p, ca, cv, dig, gi, rs string) error
}

// CustomizationClientDbInfo consists of tableName and columns
type CustomizationClientDbInfo struct {
	storeName  string // name of the mongodb collection to use for client documents
	tagMeta    string // attribute key name for the json data of a client document
	tagContent string // attribute key name for the file data of a client document
	//tagFileName string // attribute key name for storing all the file names
}

// CustomizationClient consists of CustomizationClientDbInfo
type CustomizationClient struct {
	db CustomizationClientDbInfo
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (ck CustomizationKey) String() string {
	out, err := json.Marshal(ck)
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
			//tagFileName: "customizationfiles",
		},
	}
}

// CreateCustomization creates a new Customization
func (cc *CustomizationClient) CreateCustomization(c Customization, t SpecFileContent, p, ca, cv, dig, gi, rs string, exists bool) (Customization, error) {

	key := CustomizationKey{
		Customization:       c.Metadata.Name,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
		GenericK8sIntent:    gi,
		Resource:            rs,
	}

	_, err := cc.GetCustomization(c.Metadata.Name, p, ca, cv, dig, gi, rs)
	if err == nil && !exists {
		return Customization{}, pkgerrors.New("Customization already exists")
	}

	err = db.DBconn.Insert(cc.db.storeName, key, nil, cc.db.tagMeta, c)
	if err != nil {
		return Customization{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	err = db.DBconn.Insert(cc.db.storeName, key, nil, cc.db.tagContent, t)
	if err != nil {
		return Customization{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return c, nil

}

// GetCustomization returns Customization
func (cc *CustomizationClient) GetCustomization(c, p, ca, cv, dig, gi, rs string) (Customization, error) {

	key := CustomizationKey{
		Customization:       c,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
		GenericK8sIntent:    gi,
		Resource:            rs,
	}

	value, err := db.DBconn.Find(cc.db.storeName, key, cc.db.tagMeta)
	if err != nil {
		return Customization{}, err
	}

	if len(value) == 0 {
		return Customization{}, pkgerrors.New("Customization not found")
	}

	//value is a byte array
	if value != nil {
		c := Customization{}
		err = db.DBconn.Unmarshal(value[0], &c)
		if err != nil {
			return Customization{}, err
		}
		return c, nil
	}

	return Customization{}, pkgerrors.New("Unknown Error")

}

// GetAllCustomization returns all the customization objects
func (cc *CustomizationClient) GetAllCustomization(p, ca, cv, dig, gi, rs string) ([]Customization, error) {

	key := CustomizationKey{
		Customization:       "",
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
		GenericK8sIntent:    gi,
		Resource:            rs,
	}

	var czs []Customization
	values, err := db.DBconn.Find(cc.db.storeName, key, cc.db.tagMeta)
	if err != nil {
		return []Customization{}, err
	}

	for _, value := range values {
		cz := Customization{}
		err = db.DBconn.Unmarshal(value, &cz)
		if err != nil {
			return []Customization{}, err
		}
		czs = append(czs, cz)
	}

	return czs, nil
}

// GetCustomizationContent returns the customizationContent
func (cc *CustomizationClient) GetCustomizationContent(c, p, ca, cv, dig, gi, rs string) (SpecFileContent, error) {
	key := CustomizationKey{
		Customization:       c,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
		GenericK8sIntent:    gi,
		Resource:            rs,
	}

	value, err := db.DBconn.Find(cc.db.storeName, key, cc.db.tagContent)
	if err != nil {
		return SpecFileContent{}, err
	}

	if len(value) == 0 {
		return SpecFileContent{}, pkgerrors.New("Customization Spec File Content not found")
	}

	if value != nil {
		sFileContent := SpecFileContent{}

		err = db.DBconn.Unmarshal(value[0], &sFileContent)
		if err != nil {
			return SpecFileContent{}, err
		}
		return sFileContent, nil
	}

	return SpecFileContent{}, pkgerrors.New("Unknown Error")
}

// DeleteCustomization deletes Customization
func (cc *CustomizationClient) DeleteCustomization(c, p, ca, cv, dig, gi, rs string) error {

	key := CustomizationKey{
		Customization:       c,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
		GenericK8sIntent:    gi,
		Resource:            rs,
	}

	err := db.DBconn.Remove(cc.db.storeName, key)
	return err
}
