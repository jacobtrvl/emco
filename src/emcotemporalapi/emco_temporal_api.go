// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package emcotemporalapi

// This package represents the API exported by EMCO for 3rd party workflows.
// A 3rd party workflow is expected to import this package + Temporal SDK.
// See docs/user/Temporal_Workflows_In_EMCO.md .

// TODO What about Temporal workflows in languages other than Go?
// TODO Version this API?

import (
	"fmt"

	history "go.temporal.io/api/history/v1"
	wfsvc "go.temporal.io/api/workflowservice/v1"
	cl "go.temporal.io/sdk/client"
	wf "go.temporal.io/sdk/workflow"
)

// WfTemporalSpec is the specification needed to start a workflow.
// It is part of the EMCO workflow intent (see WorkflowIntentSpec in
// workflowmgr).
type WfTemporalSpec struct {
	// Name of the workflow client to invoke. Required.
	WfClientName string `json:"workflowClientName"`
	// Options needed by wf client to start a workflow. Workflow ID is required.
	WfStartOpts cl.StartWorkflowOptions `json:"workflowStartOptions"`
	// Parameters that the wf client needs to pass to the workflow. Optional.
	WfParams WorkflowParams `json:"workflowParams,omitempty"`
}

// WorkflowParams are the per-activity data that the wf client passes to a workflow.
type WorkflowParams struct {
	// map of Temporal activity options indexed by activity name
	ActivityOpts map[string]wf.ActivityOptions `json:"activityOptions,omitempty"`
	// map of wf-specific key-value pairs indexed by activity name
	ActivityParams map[string]map[string]string `json:"activityParams,omitempty"`
}

// WfTemporalStatusQuery encapsulates the data needed to check status of a
// Temporal workflow from EMCO. It includes various flags to indicate the
// types of status queries to be run.
type WfTemporalStatusQuery struct {
	// The Temporal server's endpoint. E.g. "temporal.foo.com:7233"
	TemporalServer string `json:"temporalServer"`
	// Temporal workflow ID. TODO get this from workflow intent if not provided.
	WfID string `json:"workflowID"`
	// Temporal Run ID. If it is "", the open or latest closed wf run is used.
	RunID string `json:"runID,omitempty"`
	// WaitForResult=true: block till workflow completes.
	WaitForResult bool `json:"waitForResult,omitempty"`
	// If true, run the DescribeWorkflowExecution API.
	RunDescribeWfExec bool `json:"runDescribeWfExec,omitempty"`
	// If true, run the GetWorkflowHistory API.
	// If WaitForResult = true, this returns all history events, incl.
	// those yet to happen (using  a long poll). If false, it returns only
	// current events.
	// TODO There is an option to return just the last event; if
	// WaitForResult = true, this would be the last event which contains
	// the workflow execution end result. For now, we always return all
	// events, either till now or till the end.
	GetWfHistory bool `json:"getWfHistory,omitempty"`
	// See docs.temporal.io/docs/go/how-to-send-a-query-to-a-workflow-execution-in-go
	QueryType   string        `json:"queryType,omitempty"`
	QueryParams []interface{} `json:"queryParams,omitempty"`
}

// WfTemporalStatusResponse is the aggregation of responses from various
// Temporal status APIs.
type WfTemporalStatusResponse struct {
	WfID  string `json:"workflowID"`
	RunID string `json:"runID,omitempty"`

	// TODO This is a dump from temporal. Needs polishing.
	WfExecDesc wfsvc.DescribeWorkflowExecutionResponse `json:"workflowExecutionDescription,omitempty"`

	WfHistory []history.HistoryEvent `json:"workflowHistory,omitempty"`
	// For WfResult to be logged, it must implement the Stringer interface.
	WfResult interface{} `json:"workflowResult,omitempty"`
	// For WfQueryResult to be logged, it must implement the Stringer interface.
	WfQueryResult interface{} `json:"workflowQueryResult,omitempty"`
}

// WfTemporalCancelRequest is the set of parameters needed to invoke the
// CancelWorkflow/TerminateWorkflow APIs.
// Most fields, except the TemporalServer, are optional.
type WfTemporalCancelRequest struct {
	// The Temporal server's endpoint. E.g. "temporal.foo.com:7233". Required.
	TemporalServer string `json:"temporalServer"`
	// If WfID is specified, that overrides the one in the workflow intent.
	WfID  string `json:"workflowID,omitempty"`
	RunID string `json:"runID,omitempty"`
	// If Terminate == true, TerminateWorkflow() is called, else CancelWorkflow().
	Terminate bool          `json:"terminate,omitempty"`
	Reason    string        `json:"reason,omitempty"`
	Details   []interface{} `json:"details,omitempty"`
}

// Implement Stringer interface for query/response structs, so they can be logged.
func (q WfTemporalStatusQuery) String() string {
	return fmt.Sprintf("%#v", q)
}

func (r WfTemporalStatusResponse) String() string {
	return fmt.Sprintf("%#v", r)
}

func (r WfTemporalCancelRequest) String() string {
	return fmt.Sprintf("%#v", r)
}
