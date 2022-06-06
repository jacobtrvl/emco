// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

// GenericK8sIntent holds the intent data
type GenericK8sIntent struct {
	Metadata types.Metadata `json:"metadata"`
}

// GenericK8sIntentKey represents the resources associated with a GenericK8sIntent
type GenericK8sIntentKey struct {
	GenericK8sIntent      string `json:"genericK8sIntent"`
	Project               string `json:"project"`
	CompositeApp          string `json:"compositeApp"`
	CompositeAppVersion   string `json:"compositeAppVersion"`
	DeploymentIntentGroup string `json:"deploymentIntentGroup"`
}

// GenericK8sIntentClient holds the client properties
type GenericK8sIntentClient struct {
	db ClientDbInfo
}

// Convert the key to string to preserve the underlying structure
func (k GenericK8sIntentKey) String() string {
	out, err := json.Marshal(k)
	if err != nil {
		return ""
	}
	return string(out)
}

// NewGenericK8sIntentClient returns an instance of the GenericK8sIntentClient which implements the Manager
func NewGenericK8sIntentClient() *GenericK8sIntentClient {
	return &GenericK8sIntentClient{
		db: ClientDbInfo{
			storeName: "resources",
			tagMeta:   "data",
		},
	}
}

// GenericK8sIntentManager exposes all the functionalities related to GenericK8sIntent
type GenericK8sIntentManager interface {
	CreateGenericK8sIntent(gki GenericK8sIntent,
		project, compositeApp, compositeAppVersion, deploymentIntentGroup string,
		failIfExists bool) (GenericK8sIntent, bool, error)
	DeleteGenericK8sIntent(intent, project, compositeApp, compositeAppVersion, deploymentIntentGroup string) error
	GetAllGenericK8sIntents(project, compositeApp, compositeAppVersion, deploymentIntentGroup string) ([]GenericK8sIntent, error)
	GetGenericK8sIntent(intent, project, compositeApp, compositeAppVersion, deploymentIntentGroup string) (GenericK8sIntent, error)
}

// CreateGenericK8sIntent creates a GenericK8sIntent
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
		return GenericK8sIntent{}, gkiExists, errors.New("GenericK8sIntent already exists")
	}

	if err = db.DBconn.Insert(g.db.storeName, key, nil, g.db.tagMeta, gki); err != nil {
		return GenericK8sIntent{}, gkiExists, err
	}

	return gki, gkiExists, nil
}

// GetGenericK8sIntent returns a GenericK8sIntent
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
		return GenericK8sIntent{}, errors.New("GenericK8sIntent not found")
	}

	if value != nil {
		gki := GenericK8sIntent{}
		if err = db.DBconn.Unmarshal(value[0], &gki); err != nil {
			return GenericK8sIntent{}, err
		}
		return gki, nil
	}

	return GenericK8sIntent{}, errors.New("Unknown Error")
}

// GetAllGenericK8sIntents returns all the GenericK8sIntents
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

// DeleteGenericK8sIntent deletes a given GenericK8sIntent
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
