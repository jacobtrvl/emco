package controller

import (
	event "emcopolicy/internal/events"
	"emcopolicy/internal/intent"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"sync"
)

type Controller struct {
	policyClient *intent.Client
	eventClient  *event.Client
	db           db.Store
	tag          string
	storeName    string
	eventList    *EventList
	eventStream  chan event.Event
	updateStream chan intent.StreamData
}

// EventList contains Events to Policy Intent mapping
// This is the core in-memory data structure of policy controller
// This mapping helps to easily loop through all policy intents that is registered
// for a particular event
//
// Appends to both eventMap and Intent list is not thread-safe
// But updates to these data structures will be rare compared to the reads.
// Explicit Locking could be efficient than sync.Map
type EventList struct {
	eventMap map[event.Event][]intent.Intent
	mutex    sync.RWMutex
}
