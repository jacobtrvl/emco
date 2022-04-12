package module

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"encoding/json"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// Resource consists of metadata and Spec
type ResType struct {
	Metadata Metadata    `json:"metadata"`
	Spec     ResTypeSpec `json:"spec"`
}

type Metadata struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"-"`
	UserData1   string `json:"userData1" yaml:"-"`
	UserData2   string `json:"userData2" yaml:"-"`
}

// ResTypeSpec consists of AppName, NewObject, ExistingResource
type ResTypeSpec struct {
	AppName string      `json:"app"`
	Type    string      `json:"type"`
	CRType  ResourceGVK `json:"crType,omitempty"`
}

// ResourceGVK consists of ApiVersion, Kind, Name
type ResourceGVK struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

// ResTypeFileContent contains the content of resourceTemplate
type ResTypeFileContent struct {
	TemplateContent string `json:"filecontent"`
}

// ResTypeKey consists of resourceName, ProjectName, CompAppName, CompAppVersion, DeploymentIntentgroupName, GenericK8sIntentName
type ResTypeKey struct {
	ResType             string `json:"restype"`
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeApp"`
	CompositeAppVersion string `json:"compositeAppVersion"`
}

// ResourceManager is an interface that exposes resource related functionalities
type ResTypeManager interface {
	CreateResType(b ResType, t ResTypeFileContent, p, ca, cv string, exists bool) (ResType, error)
	GetResType(name, p, ca, cv string) (ResType, error)
	GetResTypeContent(brName, p, ca, cv string) (ResTypeFileContent, error)
	GetAllResTypes(p, ca, cv string) ([]ResType, error)
	DeleteResType(brName, p, ca, cv string) error
}

type clientDbInfo struct {
	storeName  string // name of the mongodb collection to use for client documents
	tagMeta    string // attribute key name for the json data of a client document
	tagContent string // attribute key name for the file data of a client document
}

// ResourceClient implements the resourceManager
type ResTypeClient struct {
	db clientDbInfo
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (rk ResTypeKey) String() string {
	out, err := json.Marshal(rk)
	if err != nil {
		return ""
	}
	return string(out)
}

// NewResourceClient returns an instance of the resourceClient
// which implements the Manager
func NewResTypeClient() *ResTypeClient {
	return &ResTypeClient{
		db: clientDbInfo{
			storeName:  "resources",
			tagMeta:    "data",
			tagContent: "resourcecontent",
		},
	}
}

// CreateResource creates a resource
func (rc *ResTypeClient) CreateResType(r ResType, t ResTypeFileContent, p, ca, cv string, exists bool) (ResType, error) {

	key := ResTypeKey{
		ResType:             r.Metadata.Name,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
	}

	_, err := rc.GetResType(r.Metadata.Name, p, ca, cv)
	if err == nil && !exists {
		return ResType{}, pkgerrors.New("ResType already exists")
	}
	err = db.DBconn.Insert(rc.db.storeName, key, nil, rc.db.tagMeta, r)
	if err != nil {
		return ResType{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	err = db.DBconn.Insert(rc.db.storeName, key, nil, rc.db.tagContent, t)
	if err != nil {
		return ResType{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}
	return r, nil
}

// GetResType returns a resource
func (rc *ResTypeClient) GetResType(brName, p, ca, cv string) (ResType, error) {

	key := ResTypeKey{
		ResType:             brName,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
	}

	value, err := db.DBconn.Find(rc.db.storeName, key, rc.db.tagMeta)
	if err != nil {
		return ResType{}, err
	}

	if len(value) == 0 {
		return ResType{}, pkgerrors.New("ResType not found")
	}

	//value is a byte array
	if value != nil {
		br := ResType{}
		err = db.DBconn.Unmarshal(value[0], &br)
		if err != nil {
			return ResType{}, err
		}
		return br, nil
	}

	return ResType{}, pkgerrors.New("Unknown Error")
}

// GetAllResTypes shall return all the resources for the intent
func (rc *ResTypeClient) GetAllResTypes(p, ca, cv string) ([]ResType, error) {

	key := ResTypeKey{
		ResType:             "",
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
	}

	var brs []ResType
	values, err := db.DBconn.Find(rc.db.storeName, key, rc.db.tagMeta)
	if err != nil {
		return []ResType{}, err
	}

	for _, value := range values {
		br := ResType{}
		err = db.DBconn.Unmarshal(value, &br)
		if err != nil {
			return []ResType{}, err
		}
		brs = append(brs, br)
	}

	return brs, nil
}

// GetResTypeContent returns the content of the resourceTemplate
func (rc *ResTypeClient) GetResTypeContent(rName, p, ca, cv string) (ResTypeFileContent, error) {
	key := ResTypeKey{
		ResType:             rName,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
	}

	value, err := db.DBconn.Find(rc.db.storeName, key, rc.db.tagContent)
	if err != nil {
		return ResTypeFileContent{}, err
	}

	if len(value) == 0 {
		return ResTypeFileContent{}, pkgerrors.New("ResType File Content not found")
	}

	if value != nil {
		rfc := ResTypeFileContent{}
		err = db.DBconn.Unmarshal(value[0], &rfc)
		if err != nil {
			return ResTypeFileContent{}, err
		}
		return rfc, nil
	}

	return ResTypeFileContent{}, pkgerrors.New("Unknown Error")
}

// DeleteResType deletes a given resource
func (rc *ResTypeClient) DeleteResType(rName, p, ca, cv string) error {
	key := ResTypeKey{
		ResType:             rName,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
	}

	err := db.DBconn.Remove(rc.db.storeName, key)
	return err
}
