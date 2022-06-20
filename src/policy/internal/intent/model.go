package intent

import (
	event "emcopolicy/internal/events"
	events "emcopolicy/pkg/grpc"
	"encoding/json"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

type Client struct {
	db           db.Store
	tag          string
	storeName    string
	updateStream chan StreamData
}

type Config struct {
	Db           db.Store
	Tag          string
	StoreName    string
	UpdateStream chan StreamData
}

type PolicySpec struct {
	EngineUrl  string `json:"engineUrl"`
	Namespace  string `json:"namespace"`
	PolicyName string `json:"policyName"`
}

type Spec struct {
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
	Metadata Metadata `json:"metadata"`
	Spec     Spec     `json:"spec"`
}

type Metadata struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"-"`
	UserData1   string `json:"userData1" yaml:"-"`
	UserData2   string `json:"userData2" yaml:"-"`
}

type Data struct {
	PolicyId   string
	Actor      Actor
	ActorParam any
	Events     []events.Event
}

type Request struct {
	Project               string
	CompositeApp          string
	CompositeAppVersion   string
	DeploymentIntentGroup string
	PolicyIntentId        string
	IntentData            *Intent
}

type Key struct {
	PolicyIntent        string `json:"policyIntent"`
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeApp"`
	CompositeAppVersion string `json:"compositeAppVersion"`
	DigName             string `json:"deploymentIntentGroup"`
}

type StreamData struct {
	Operation string
	Intent    Intent
}
