// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	pkgerrors "github.com/pkg/errors"
	"go.temporal.io/sdk/client"

	tmpl "gitlab.com/project-emco/core/emco-base/src/emcotemporalapi"
	"gitlab.com/project-emco/core/emco-base/src/workflowmgr/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/workflowmgr/pkg/infra/logutils"
	enums "go.temporal.io/api/enums/v1"
	history "go.temporal.io/api/history/v1"
)

// WorkflowIntent contains the parameters needed for managing workflows
type WorkflowIntent struct {
	Metadata Metadata           `json:"metadata"`
	Spec     WorkflowIntentSpec `json:"spec"`
}

// WorkflowIntentSpec is the specification of an EMCO workflow intemt,
type WorkflowIntentSpec struct {
	// Network endpoint at which the workflow client resides.
	WfClientSpec WfClientSpec `json:"workflowClient"`
	// See emcotemporalapi package.
	WfTemporalSpec tmpl.WfTemporalSpec `json:"temporal"`
}

// WfClientSpec is the network endpoint at which the workflow client resides.
type WfClientSpec struct {
	WfClientEndpointName string `json:"clientEndpointName"`
	WfClientEndpointPort int    `json:"clientEndpointPort"`
}

// WorkflowIntentKey is the key structure that is used in the database
type WorkflowIntentKey struct {
	WorkflowIntent      string `json:"workflowIntent"`
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeApp"`
	CompositeAppVersion string `json:"compositeAppVersion"`
	DigName             string `json:"deploymentIntentGroup"`
}

// Manager is an interface exposing the WorkflowIntent functionality
type WorkflowIntentManager interface {
	CreateWorkflowIntent(wfi WorkflowIntent, project, cApp, cAppVer, dig string, exists bool) (WorkflowIntent, error)
	GetWorkflowIntent(name, project, cApp, cAppVer, dig string) (WorkflowIntent, error)
	GetWorkflowIntents(project, cApp, cAppVer, dig string) ([]WorkflowIntent, error)
	DeleteWorkflowIntent(name, project, cApp, cAppVer, dig string) error
	StartWorkflowIntent(name, project, cApp, cAppVer, dig string) error
	GetStatusWorkflowIntent(name, project, cApp, cAppVer, dig string,
		query *tmpl.WfTemporalStatusQuery) (*tmpl.WfTemporalStatusResponse, error)
}

// WorkflowIntentClient implements the Manager
// It will also be used to maintain some localized state
type WorkflowIntentClient struct {
	db ClientDbInfo
}

// NewWorkflowIntentClient returns an instance of the WorkflowIntentClient
// which implements the Manager
func NewWorkflowIntentClient() *WorkflowIntentClient {
	return &WorkflowIntentClient{
		db: ClientDbInfo{
			storeName: "resources",
			tagMeta:   "data",
		},
	}
}

// CreateWorkflowIntent - create a new WorkflowIntent
func (v *WorkflowIntentClient) CreateWorkflowIntent(wfi WorkflowIntent,
	project, cApp, cAppVer, dig string, exists bool) (WorkflowIntent, error) {

	log.Warn("CreateWFI", log.Fields{"WfIntent": wfi, "project": project,
		"cApp": cApp})
	//Construct key and tag to select the entry
	key := WorkflowIntentKey{
		WorkflowIntent:      wfi.Metadata.Name,
		Project:             project,
		CompositeApp:        cApp,
		CompositeAppVersion: cAppVer,
		DigName:             dig,
	}

	//Check if this WorkflowIntent already exists
	_, err := v.GetWorkflowIntent(wfi.Metadata.Name, project, cApp, cAppVer, dig)
	if err == nil && !exists {
		return WorkflowIntent{}, pkgerrors.New("WorkflowIntent already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, wfi)
	if err != nil {
		return WorkflowIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return wfi, nil
}

// GetWorkflowIntent returns the named Workflow intent.
func (v *WorkflowIntentClient) GetWorkflowIntent(name,
	project, cApp, cAppVer, dig string) (WorkflowIntent, error) {

	//Construct key and tag to select the entry
	key := WorkflowIntentKey{
		WorkflowIntent:      name,
		Project:             project,
		CompositeApp:        cApp,
		CompositeAppVersion: cAppVer,
		DigName:             dig,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return WorkflowIntent{}, err
	} else if len(value) == 0 {
		return WorkflowIntent{}, pkgerrors.New("Workflow Intent not found")
	}

	//value is a byte array
	if value == nil {
		return WorkflowIntent{}, pkgerrors.New("Unknown Error")
	}

	wfi := WorkflowIntent{}
	if err = db.DBconn.Unmarshal(value[0], &wfi); err != nil {
		return WorkflowIntent{}, err
	}

	log.Warn("GetWFI", log.Fields{"WfIntent": wfi, "project": project,
		"cApp": cApp})
	return wfi, nil
}

// GetWorkflowIntents returns all WorkflowIntents for a DIG.
func (v *WorkflowIntentClient) GetWorkflowIntents(
	project, cApp, cAppVer, dig string) ([]WorkflowIntent, error) {

	//Construct key and tag to select the entry
	key := WorkflowIntentKey{
		WorkflowIntent:      "",
		Project:             project,
		CompositeApp:        cApp,
		CompositeAppVersion: cAppVer,
		DigName:             dig,
	}

	var resp []WorkflowIntent
	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []WorkflowIntent{}, err
	}

	for _, value := range values {
		wfi := WorkflowIntent{}
		err = db.DBconn.Unmarshal(value, &wfi)
		if err != nil {
			return []WorkflowIntent{}, err
		}
		resp = append(resp, wfi)
	}

	return resp, nil
}

// Delete the  WorkflowIntent from database
func (v *WorkflowIntentClient) DeleteWorkflowIntent(name,
	project, cApp, cAppVer, dig string) error {

	//Construct key and tag to select the entry
	key := WorkflowIntentKey{
		WorkflowIntent:      name,
		Project:             project,
		CompositeApp:        cApp,
		CompositeAppVersion: cAppVer,
		DigName:             dig,
	}

	err := db.DBconn.Remove(v.db.storeName, key)
	return err
}

// Start the workflow
func (v *WorkflowIntentClient) StartWorkflowIntent(name,
	project, cApp, cAppVer, dig string) error {

	wfi, err := v.GetWorkflowIntent(name, project, cApp, cAppVer, dig)
	if err != nil {
		log.Error("StartWorkflowIntent failed to get workflow intent",
			log.Fields{"error": err.Error()})
		return err
	}

	url := "http://" + wfi.Spec.WfClientSpec.WfClientEndpointName + ":" +
		strconv.Itoa(wfi.Spec.WfClientSpec.WfClientEndpointPort) + "/invoke/" +
		wfi.Spec.WfTemporalSpec.WfClientName

	jsonBytes, err := json.Marshal(wfi.Spec.WfTemporalSpec)
	if err != nil {
		log.Error("StartWorkflowIntent marshaling error",
			log.Fields{"error": err.Error()})
		return err
	}
	log.Info("StartWorkflowIntent",
		log.Fields{"url": url, "wfi": string(jsonBytes)})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Error("StartWorkflowIntent could not POST",
			log.Fields{"error": err.Error()})
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		postErr := fmt.Errorf("HTTP POST returned status code %s for URL %s.\n",
			resp.Status, url)
		log.Error("StartWorkflowIntent POST returned error",
			log.Fields{"status code": resp.Status, "urL": url})
		return postErr

	}

	return nil
}

// GetStatusWorkflowIntent performs different types of Temporal workflow
// status queries depending on the flags specified in the status API call.
func (v *WorkflowIntentClient) GetStatusWorkflowIntent(name, project, cApp, cAppVer,
	dig string, query *tmpl.WfTemporalStatusQuery) (*tmpl.WfTemporalStatusResponse, error) {
	log.Info("Entered GetStatusWorkflowIntent", log.Fields{"project": project,
		"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
		"intent name": name, "query": query, "queryType": query.QueryType})

	resp := tmpl.WfTemporalStatusResponse{
		WfID:  query.WfID,
		RunID: query.RunID,
	}

	clientOptions := client.Options{HostPort: query.TemporalServer}
	c, err := client.NewClient(clientOptions)
	ctx := context.Background() // TODO include query options later
	if err != nil {
		wrapErr := fmt.Errorf("Failed to connect to Temporal server (%s). Error: %s",
			query.TemporalServer, err.Error())
		log.Error(wrapErr.Error(), log.Fields{"project": project,
			"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
			"intent name": name, "query": query})
		return &resp, wrapErr
	}

	if query.RunDescribeWfExec {
		log.Info("Running DescribeWorkflowExecution", log.Fields{"project": project,
			"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
			"intent name": name, "query": query})
		result, err := c.DescribeWorkflowExecution(ctx, query.WfID, query.RunID)
		if err != nil {
			log.Error("DescribeWorkflowExecution error", log.Fields{"project": project,
				"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
				"intent name": name, "query": query, "error": err.Error()})
			return &resp, err
		}
		log.Info("DescribeWorkflowExecution success", log.Fields{"project": project,
			"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
			"intent name": name, "query": query})
		resp.WfExecDesc = *result
	}

	if query.GetWfHistory {
		msg := "Getting workflow history"
		if query.WaitForResult {
			msg += ": will now block till workflow completes"
		}
		log.Info(msg, log.Fields{"name": name, "project": project, "cApp": cApp,
			"cAppVer": cAppVer, "dig": dig, "intent name": name,
			"workflow ID": query.WfID, "workflow Run ID": query.RunID,
			"Temporal Server": query.TemporalServer,
		})

		resp.WfHistory = []history.HistoryEvent{}
		iter := c.GetWorkflowHistory(ctx, query.WfID, query.RunID,
			query.WaitForResult, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
		for iter.HasNext() {
			event, err := iter.Next()
			if err != nil {
				log.Error("c.GetWorkflowHistory error", log.Fields{"project": project,
					"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
					"intent name": name, "query": query, "error": err.Error()})
				return &resp, err
			}
			resp.WfHistory = append(resp.WfHistory, *event)
		}
		log.Info("Got workflow history", log.Fields{"name": name, "project": project,
			"cApp": cApp, "cAppVer": cAppVer, "dig": dig, "intent name": name,
			"workflow ID": query.WfID, "workflow Run ID": query.RunID,
			"Temporal Server": query.TemporalServer,
		})
	}

	if query.QueryType != "" {
		log.Info("Querying workflow", log.Fields{"name": name, "project": project,
			"cApp": cApp, "cAppVer": cAppVer, "dig": dig, "intent name": name,
			"workflow ID": query.WfID, "workflow Run ID": query.RunID,
			"queryType": query.QueryType, "queryARgs": query.QueryParams,
			"Temporal Server": query.TemporalServer,
		})
		queryWithOptions := &client.QueryWorkflowWithOptionsRequest{
			WorkflowID:           query.WfID,
			RunID:                query.RunID,
			QueryType:            query.QueryType,
			Args:                 query.QueryParams,
			QueryRejectCondition: enums.QUERY_REJECT_CONDITION_NONE,
		}
		response, err := c.QueryWorkflowWithOptions(ctx, queryWithOptions)
		if err != nil {
			wrapErr := fmt.Errorf("Query failed (%s). Error: %s",
				query.QueryType, err.Error())
			log.Error(wrapErr.Error(), log.Fields{"project": project,
				"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
				"intent name": name, "query": query})
			return &resp, wrapErr
		}
		// type of response: *QueryWorkflowWithOptionsResponse
		// TODO handle response.QueryRejected (reason why query was rejected, if any
		if !response.QueryResult.HasValue() {
			wrapErr := fmt.Errorf("Got no result for query (%s).", query.QueryType)
			log.Error(wrapErr.Error(), log.Fields{"project": project,
				"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
				"intent name": name, "query": query})
			return &resp, wrapErr
		}

		err = response.QueryResult.Get(&resp.WfQueryResult)
		if err != nil {
			wrapErr := fmt.Errorf("Failed to get result for query (%s). Error: %s",
				query.QueryType, err.Error())
			log.Error(wrapErr.Error(), log.Fields{"project": project,
				"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
				"intent name": name, "query": query})
			return &resp, wrapErr
		}
		log.Info("Query got result", log.Fields{"project": project,
			"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
			"intent name": name, "query": query, "result": resp.WfQueryResult})
	}

	if query.WaitForResult {
		// TODO Status query can take options for timeouts and retries.
		log.Info("GetStatusWorkflowIntent will now block till workflow completes",
			log.Fields{"name": name, "project": project, "cApp": cApp,
				"cAppVer": cAppVer, "dig": dig, "intent name": name,
				"workflow ID": query.WfID, "workflow Run ID": query.RunID,
				"Temporal Server": query.TemporalServer,
			})
		workflowRun := c.GetWorkflow(context.Background(), query.WfID, query.RunID)

		var result interface{}
		err = workflowRun.Get(ctx, &result)
		log.Info("Workflow got result", log.Fields{"project": project,
			"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
			"intent name": name, "query": query, "result": result})
	}

	return &resp, nil
}
