package policy

import (
	event "emcopolicy/internal/events"
	events "emcopolicy/pkg/grpc"
	"encoding/json"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// Client holds
// Do not declare any fields in Client with exported variables
type Client struct {
	db        db.Store
	tag       string
	storeName string
}

type Policy struct {
	Metadata Metadata   `json:"metadata"`
	Spec     PolicySpec `json:"spec"`
}

type PolicyRequest struct {
	PolicyId string
	Policy   *Policy
}

type PolicyKey struct {
	Id string
}

type PolicySpec struct {
	EngineUrl  string `json:"engineUrl"`
	Namespace  string `json:"namespace"`
	PolicyName string `json:"policyName"`
}

type IntentSpec struct {
	PolicyIntentID        string           `json:"policyIntentID"`
	Project               string           `json:"project"`
	CompositeApp          string           `json:"compositeApp"`
	CompositeAppVersion   string           `json:"compositeAppVersion"`
	DeploymentIntentGroup string           `json:"deploymentIntentGroup"`
	Policy                PolicySpec       `json:"policy"`
	Actor                 string           `json:"actor"`
	ActorArg              *json.RawMessage `json:"actorArg,omitempty"`
	Event                 event.Event      `json:"event"`
	SupportingEvents      []event.Event    `json:"supportingEvent,omitempty"`
}

type Intent struct {
	Metadata Metadata   `json:"metadata"`
	Spec     IntentSpec `json:"spec"`
}

type Metadata struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"-"`
	UserData1   string `json:"userData1" yaml:"-"`
	UserData2   string `json:"userData2" yaml:"-"`
}

type IntentData struct {
	PolicyId   string
	Actor      Actor
	ActorParam any
	Events     []events.Event
}

type IntentRequest struct {
	Project               string
	CompositeApp          string
	CompositeAppVersion   string
	DeploymentIntentGroup string
	PolicyIntentId        string
	IntentData            *Intent
}

type IntentKey struct {
	PolicyIntent        string `json:"policyIntent"`
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeApp"`
	CompositeAppVersion string `json:"compositeAppVersion"`
	DigName             string `json:"deploymentIntentGroup"`
}
