// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"gitlab.com/project-emco/core/emco-base/examples/sample-controller/pkg/model"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"

	"github.com/pkg/errors"
)

// SampleIntentClient implements the SampleIntentManager.
// It will also be used to maintain some localized state.
type SampleIntentClient struct {
	dbInfo DBInfo
}

func NewIntentClient() *SampleIntentClient {
	return &SampleIntentClient{
		dbInfo: DBInfo{
			collection: "resources", // should remain the same
			tag:        "data",      // should remain the same
		},
	}
}

// A manager is an interface for exposing the client's functionalities.
// You can have multiple managers based on the requirement and its implementation.
// In this example, SampleIntentManager exposes the SampleIntentClient functionalities.
type SampleIntentManager interface {
	CreateSampleIntent(intent model.SampleIntent, project, app, version, deploymentIntentGroup string, failIfExists bool) (model.SampleIntent, error)
	GetSampleIntents(name, project, app, version, deploymentIntentGroup string) ([]model.SampleIntent, error)
}

// CreateSampleIntent insert a new SampleIntent in the database
func (i *SampleIntentClient) CreateSampleIntent(intent model.SampleIntent, project, app, version, deploymentIntentGroup string, failIfExists bool) (model.SampleIntent, error) {
	// Construct key and tag to select the entry.
	key := model.SampleIntentKey{
		Project:               project,
		CompositeApp:          app,
		CompositeAppVersion:   version,
		DeploymentIntentGroup: deploymentIntentGroup,
		SampleIntent:          intent.Metadata.Name,
	}

	// Check if this SampleIntent already exists.
	intents, err := i.GetSampleIntents(intent.Metadata.Name, project, app, version, deploymentIntentGroup)
	if err == nil &&
		len(intents) > 0 &&
		intents[0].Metadata.Name == intent.Metadata.Name &&
		failIfExists {
		return model.SampleIntent{}, errors.New("SampleIntent already exists")
	}

	err = db.DBconn.Insert(i.dbInfo.collection, key, nil, i.dbInfo.tag, intent)
	if err != nil {
		return model.SampleIntent{}, err
	}

	return intent, nil
}

// GetSampleIntents returns the SampleIntent for the corresponding name
func (i *SampleIntentClient) GetSampleIntents(name, project, app, version, deploymentIntentGroup string) ([]model.SampleIntent, error) {
	// Construct key and tag to select the entry.
	key := model.SampleIntentKey{
		Project:               project,
		CompositeApp:          app,
		CompositeAppVersion:   version,
		DeploymentIntentGroup: deploymentIntentGroup,
		SampleIntent:          name,
	}

	values, err := db.DBconn.Find(i.dbInfo.collection, key, i.dbInfo.tag)
	if err != nil {
		return []model.SampleIntent{}, err
	}

	if len(values) == 0 {
		return []model.SampleIntent{}, errors.New("SampleIntent not found")
	}

	intents := []model.SampleIntent{}

	for _, v := range values {
		i := model.SampleIntent{}
		err = db.DBconn.Unmarshal(v, &i)
		if err != nil {
			return []model.SampleIntent{}, errors.Wrap(err, "Unmarshalling Value")
		}

		intents = append(intents, i)
	}

	return intents, nil
}
