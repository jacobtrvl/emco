package module

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"encoding/json"
	"reflect"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// GenericK8sIntent consists of metadata
type GenericK8sIntent struct {
	Metadata Metadata `json:"metadata"`
}

// GenericK8sIntentKey consists generick8sintentName, project, compositeApp,
// compositeAppVersion, deploymentIntentGroupName
type GenericK8sIntentKey struct {
	GenericK8sIntent      string `json:"genericK8sIntent"`
	Project               string `json:"project"`
	CompositeApp          string `json:"compositeApp"`
	CompositeAppVersion   string `json:"compositeAppVersion"`
	DeploymentIntentGroup string `json:"deploymentIntentGroup"`
}

// GenericK8sIntentManager is an interface exposing the GenericK8sIntent functionality
type GenericK8sIntentManager interface {
	CreateGenericK8sIntent(gki GenericK8sIntent,
		project, compositeApp, compositeAppVersion, deploymentIntentGroup string,
		failIfExists bool) (GenericK8sIntent, bool, error)
	DeleteGenericK8sIntent(intent, project, compositeApp, compositeAppVersion, deploymentIntentGroup string) error
	GetAllGenericK8sIntents(project, compositeApp, compositeAppVersion, deploymentIntentGroup string) ([]GenericK8sIntent, error)
	GetGenericK8sIntent(intent, project, compositeApp, compositeAppVersion, deploymentIntentGroup string) (GenericK8sIntent, error)
}

// GenericK8sIntentClient consists of the clientInfo
type GenericK8sIntentClient struct {
	db ClientDbInfo
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (k GenericK8sIntentKey) String() string {
	out, err := json.Marshal(k)
	if err != nil {
		return ""
	}
	return string(out)
}

// NewGenericK8sIntentClient returns an instance of the GenericK8sIntentClient
func NewGenericK8sIntentClient() *GenericK8sIntentClient {
	return &GenericK8sIntentClient{
		db: ClientDbInfo{
			storeName: "resources",
			tagMeta:   "data",
		},
	}
}

// CreateGenericK8sIntent creates a new GenericK8sIntent
func (g *GenericK8sIntentClient) CreateGenericK8sIntent(gki GenericK8sIntent,
	project, compositeApp, compositeAppVersion, deploymentIntentGroup string,
	failIfExists bool) (GenericK8sIntent, bool, error) {

	gkiExists := false
	key := GenericK8sIntentKey{
		GenericK8sIntent:      gki.Metadata.Name,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   compositeAppVersion,
		DeploymentIntentGroup: deploymentIntentGroup,
	}

	i, err := g.GetGenericK8sIntent(gki.Metadata.Name, project, compositeApp, compositeAppVersion, deploymentIntentGroup)
	if err == nil &&
		!reflect.DeepEqual(i, GenericK8sIntent{}) {
		gkiExists = true
	}

	if gkiExists &&
		failIfExists {
		return GenericK8sIntent{}, gkiExists, pkgerrors.New("GenericK8sIntent already exists")
	}

	if err = db.DBconn.Insert(g.db.storeName, key, nil, g.db.tagMeta, gki); err != nil {
		return GenericK8sIntent{}, gkiExists, err
	}

	return gki, gkiExists, nil
}

// GetGenericK8sIntent returns GenericK8sIntent with the corresponding name
func (g *GenericK8sIntentClient) GetGenericK8sIntent(intent, project, compositeApp, compositeAppVersion,
	deploymentIntentGroup string) (GenericK8sIntent, error) {

	key := GenericK8sIntentKey{
		GenericK8sIntent:      intent,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   compositeAppVersion,
		DeploymentIntentGroup: deploymentIntentGroup,
	}

	value, err := db.DBconn.Find(g.db.storeName, key, g.db.tagMeta)
	if err != nil {
		return GenericK8sIntent{}, err
	}

	if len(value) == 0 {
		return GenericK8sIntent{}, pkgerrors.New("GenericK8sIntent not found")
	}

	if value != nil {
		gki := GenericK8sIntent{}
		if err = db.DBconn.Unmarshal(value[0], &gki); err != nil {
			return GenericK8sIntent{}, err
		}
		return gki, nil
	}

	return GenericK8sIntent{}, pkgerrors.New("Unknown Error")
}

// GetAllGenericK8sIntents returns all of the GenericK8sIntent for corresponding name
func (g *GenericK8sIntentClient) GetAllGenericK8sIntents(project, compositeApp, compositeAppVersion,
	deploymentIntentGroup string) ([]GenericK8sIntent, error) {

	key := GenericK8sIntentKey{
		GenericK8sIntent:      "",
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   compositeAppVersion,
		DeploymentIntentGroup: deploymentIntentGroup,
	}

	values, err := db.DBconn.Find(g.db.storeName, key, g.db.tagMeta)
	if err != nil {
		return []GenericK8sIntent{}, err
	}

	var intents []GenericK8sIntent
	for _, value := range values {
		gki := GenericK8sIntent{}
		if err = db.DBconn.Unmarshal(value, &gki); err != nil {
			return []GenericK8sIntent{}, err
		}
		intents = append(intents, gki)
	}

	return intents, nil
}

// DeleteGenericK8sIntent delete the GenericK8sIntent entry from the database
func (g *GenericK8sIntentClient) DeleteGenericK8sIntent(intent, project, compositeApp, compositeAppVersion,
	deploymentIntentGroup string) error {

	key := GenericK8sIntentKey{
		GenericK8sIntent:      intent,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   compositeAppVersion,
		DeploymentIntentGroup: deploymentIntentGroup,
	}

	return db.DBconn.Remove(g.db.storeName, key)
}
