package controller

import (
	"context"
	event "emcopolicy/internal/events"
	"emcopolicy/internal/intent"
	"fmt"
	"github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"go.mongodb.org/mongo-driver/bson"
)

// scheduler listens to the different channels for updates/events and invoke
// corresponding  goroutines for handling them.
// Handle functions like handleIntentUpdate runs in scheduler context, not in
// the http request context, since those request will be completed once channel is read.
// There's possible corner case in current implementation, that if handling is failed,
// the updates will not get reflected in in-memory structure till next application boot.
func (c *Controller) scheduler(ctx context.Context) error {
	var (
		e event.Event
		p intent.StreamData
	)
	log.Info("Starting Scheduler. Listening to events", log.Fields{})
	for {
		select {
		case e = <-c.eventStream:
			go c.handleEvent(ctx, e)
		case p = <-c.updateStream:
			go c.handleIntentUpdate(ctx, p)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Controller) handleEvent(_ context.Context, _ event.Event) {
	// do real action.. call opa.. call actor
	fmt.Println("Lets call it OPA")
}

func (c *Controller) handleIntentUpdate(ctx context.Context, data intent.StreamData) {
	switch data.Operation {
	case "DELETE":
		c.deleteIntent(ctx, data.Intent)
	case "APPEND":
		c.addIntent(ctx, data.Intent)
	}
}

func (c *Controller) deleteIntent(_ context.Context, p intent.Intent) {
	err := c.eventList.RemovePolicyIntent(p.Spec.Event, p)
	if err != nil {
		log.Error("Delete intent failed", log.Fields{"error": err})
	}
}

func (c *Controller) addIntent(_ context.Context, p intent.Intent) {
	c.eventList.AddPolicyIntent(p.Spec.Event, p)
}

// AddPolicyIntent adds an intent from event's intentList in eventMap
func (l *EventList) AddPolicyIntent(e event.Event, p intent.Intent) {
	var index int
	l.mutex.Lock()
	defer l.mutex.Unlock()
	// Remove Intent from list if this is an update.
	// Deletion can be handled by  RemovePolicyIntent, but it's safer to do while lock is holding
	// This lock could starve other appends and reads, but it is unavoidable in current design.
	// This delay won't be user visible since this is a background activity
	// Also appends will be less often compared to reads
	for index < len(l.eventMap[e]) && !isEqual(l.eventMap[e][index], p) {
		index++
	}
	if index < len(l.eventMap[e]) {
		l.eventMap[e] = append(l.eventMap[e][:index], l.eventMap[e][index+1:]...)
	}
	if _, ok := l.eventMap[e]; !ok {
		l.eventMap[e] = []intent.Intent{}
	}
	l.eventMap[e] = append(l.eventMap[e], p)
	log.Info("Added Intent to in-memory event list", log.Fields{"Event": e.Id, "Intent": p.Spec.PolicyIntentID})
}

// RemovePolicyIntent deletes an intent from event's intentList in eventMap
func (l *EventList) RemovePolicyIntent(e event.Event, p intent.Intent) error {
	var index int
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for index < len(l.eventMap[e]) && !isEqual(l.eventMap[e][index], p) {
		index++
	}
	if index == len(l.eventMap[e]) {
		i := p.Spec.Project + "/" + p.Spec.CompositeApp + "/" + p.Spec.CompositeAppVersion + "/" +
			p.Spec.DeploymentIntentGroup + "/" + p.Spec.PolicyIntentID
		return errors.Errorf("RemoveIntent failed. No Policy Intent "+
			"(%s) found in the IntentList of event %s", i, e.Id)
	}
	l.eventMap[e] = append(l.eventMap[e][:index], l.eventMap[e][index+1:]...)
	log.Info("Removed Intent from in-memory event list", log.Fields{"Event": e.Id, "Intent": p.Spec.PolicyIntentID})
	return nil
}

// GetPolicyIntentList returns intentList of an event from eventMap
func (l *EventList) GetPolicyIntentList(e event.Event) (error, []intent.Intent) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	if _, ok := l.eventMap[e]; !ok {
		// TODO error
		return errors.Errorf("GetPolicyIntentList failed. Event %s not found", e.Id), nil
	}
	intentCopy := make([]intent.Intent, len(l.eventMap[e]))
	// slice can get modified from other threads
	// Hence returning a snapshot
	copy(intentCopy, l.eventMap[e])
	return nil, intentCopy
}

func isEqual(p, q intent.Intent) bool {
	return p.Spec.Project == q.Spec.Project &&
		p.Spec.CompositeApp == q.Spec.CompositeApp &&
		p.Spec.CompositeAppVersion == q.Spec.CompositeAppVersion &&
		p.Spec.DeploymentIntentGroup == q.Spec.DeploymentIntentGroup &&
		p.Spec.PolicyIntentID == q.Spec.PolicyIntentID
}

// BuildEventListFromDB builds a reverse map of events to policy.
// Should be called only during boot up, when no other controller threads are running
// This method modifies eventList without locking for a better boot up performance
func (c *Controller) BuildEventListFromDB(ctx context.Context) error {

	policyIntents, err := c.getAllPolicyIntents(ctx)
	if err != nil {
		return errors.Wrap(err, "Building In-memory EventList failed::")
	}

	for _, policyIntent := range policyIntents {
		if c.eventList.eventMap[policyIntent.Spec.Event] == nil {
			c.eventList.eventMap[policyIntent.Spec.Event] = []intent.Intent{}
		}
		c.eventList.eventMap[policyIntent.Spec.Event] = append(c.eventList.eventMap[policyIntent.Spec.Event], policyIntent)
	}

	return nil
}

// getAllPolicyIntents is helper function for getting all the policy intents from DB
// If there is no policy intent, it will return  ([]intent.Intent{}, nil)
func (c *Controller) getAllPolicyIntents(_ context.Context) ([]intent.Intent, error) {
	var intents []intent.Intent

	key := struct {
		PolicyIntent bson.M `json:"policyIntent"`
	}{bson.M{"$exists": true}}
	value, err := c.db.Find(c.storeName, key, c.tag)
	if err != nil {
		return intents, err
	}

	for _, v := range value {
		item := new(intent.Intent)
		if err := c.db.Unmarshal(v, item); err != nil {
			return nil, err
		}
		intents = append(intents, *item)
	}

	return intents, nil
}
