package controller

import (
	"context"
	event "emcopolicy/internal/events"
	"emcopolicy/internal/intent"
	events "emcopolicy/pkg/grpc"
	"encoding/json"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"time"
)

// OperationalScheduler listens to the different channels for management updates  and invoke
// corresponding  goroutines for handling them.
// Handle functions like handleIntentUpdate runs in OperationalScheduler context, not in
// the http request context, since those request will be completed once channel is read.
// There's possible corner case in current implementation, that if handling is failed,
// the updates will not get reflected in in-memory structure till next application boot.
func (c *Controller) OperationalScheduler(ctx context.Context) {
	var (
		p intent.StreamData
		a event.StreamAgentData
	)
	log.Info("Starting Scheduler. Listening to events", log.Fields{})
	for {
		select {
		case p = <-c.updateStream:
			go c.handleIntentUpdate(ctx, p)
		case a = <-c.agentStream:
			go c.handleAgentUpdate(ctx, a)
		case <-ctx.Done():
			log.Warn("OperationalScheduler exiting", log.Fields{"Content Err": ctx.Err()})
			return
		}
	}
}

// AgentManager blah blah
func (c *Controller) AgentManager(ctx context.Context) {
	log.Info("Starting Agent Manager", log.Fields{})
	for {
		var notRunning []AgentID
		checkAgents := func() {
			c.agentMap.mutex.RLock()
			defer c.agentMap.mutex.RUnlock()
			for id, runtime := range c.agentMap.runtime {
				if !runtime.isRunning {
					notRunning = append(notRunning, id)
				}
			}
		}
		select {
		case <-c.requireRecovery:
			log.Info("Require agent recovery. Checking agents", log.Fields{})
			checkAgents()
		case <-time.After(time.Minute * 5):
			log.Info("Scheduled Health Check of Agents", log.Fields{})
			checkAgents()
		}
		// Why this step of restart of agent not done along with checkAgents?
		// Number of agents to be started will a small number compared to total number of agents
		// In checkAgents, we traverse all the agents to identify which agents require attention.
		// This process require only read-lock, while update of agent require a write-lock.
		// Hence, to avoid holding write-lock for longer period, updating agents as a separate step
		// Updates that can happen in between these two critical session can be ignored, since next run
		// will take care of them.
		func() {
			c.agentMap.mutex.Lock()
			defer c.agentMap.mutex.Unlock()
			for _, id := range notRunning {
				var newCtx context.Context
				newCtx, c.agentMap.runtime[id].cancel = context.WithCancel(ctx)
				c.agentMap.runtime[id].isRunning = true
				go event.ListenOne(newCtx, c.agentMap.runtime[id].spec.EndPoint, func() {
					c.markForRecovery(id)
				}, c.eventsQueue)
			}
		}()
	}
}

func (c *Controller) EventsManager(_ context.Context) {
	log.Info("Starting Events manager", log.Fields{})
	for e := range c.eventsQueue {
		go c.processEvent(e)
	}
}

func (c *Controller) markForRecovery(id AgentID) {
	c.agentMap.MarkForRecovery(id)
	c.requireRecovery <- struct{}{}
}

func (c *Controller) handleAgentUpdate(_ context.Context, data event.StreamAgentData) {
	if c.agentMap == nil || c.agentMap.runtime == nil {
		log.Fatal("AgentMap or runtime is nil", log.Fields{})
	}
	switch data.Operation {
	case "DELETE":
		c.deleteAgent(data.Spec)
	case "APPEND":
		c.appendAgent(data.Spec)
	}

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
	err := c.reverseMap.RemovePolicyIntent(p.Spec.Event, p)
	if err != nil {
		log.Error("Delete intent failed", log.Fields{"error": err})
	}
}

func (c *Controller) addIntent(_ context.Context, p intent.Intent) {
	c.reverseMap.AddPolicyIntent(p.Spec.Event, p)
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

func (c *Controller) getAllAgents(ctx context.Context) ([]event.AgentSpec, error) {
	return c.eventClient.GetAllAgents(ctx)
}

func (c *Controller) deleteAgent(agent event.AgentSpec) {
	var (
		agentId = AgentID(agent.Id)
	)
	if !c.agentMap.IsAgentExists(agentId) {
		log.Warn("Delete failed: AgentID not found", log.Fields{"Id": agent.Id})
		return
	}
	c.agentMap.Cancel(agentId)
	c.agentMap.DeleteAgent(agentId)
}

func (c *Controller) appendAgent(agent event.AgentSpec) {
	var (
		agentId = AgentID(agent.Id)
	)
	if c.agentMap.VerifySpec(agentId, agent) {
		log.Warn("appendAgent Specs are same. Not updated", log.Fields{"AgentId": agentId})
		return
	}
	c.agentMap.Cancel(agentId)
	c.agentMap.UpdateSpec(agentId, agent)
	c.requireRecovery <- struct{}{}
}

func (c *Controller) processEvent(e *events.Event) {
	agentId := e.AgentId
	eventId := e.EventId
	log.Debug("Processing event", log.Fields{"id": eventId, "agent": agentId})
	// Convert agent spec from proto format to Json
	agentSpec, err := anypb.UnmarshalNew(e.Spec, proto.UnmarshalOptions{})
	if err != nil {
		log.Error("processEvent failed", log.Fields{"err": err, "id": eventId, "agent": agentId})
		return
	}
	agentSpecJson, err := json.Marshal(agentSpec)
	if err != nil {
		log.Error("processEvent failed", log.Fields{"err": err, "id": eventId, "agent": agentId})
		return
	}
	// Convert event message from proto format to Json
	eventMessage, err := anypb.UnmarshalNew(e.Message, proto.UnmarshalOptions{})
	if err != nil {
		log.Error("processEvent failed", log.Fields{"err": err, "id": eventId, "agent": agentId})
		return
	}
	eventMessageJson, err := json.Marshal(eventMessage)
	if err != nil {
		log.Error("processEvent failed", log.Fields{"err": err, "id": eventId, "agent": agentId})
		return
	}
	// Process events with Agent ID mentioned in policy intent
	eventSpec := event.Event{
		Id:      eventId,
		AgentID: agentId,
	}
	// Deep copying the intent. This allows remaining tasks,
	// which can be time-consuming to do without taking lock
	intentJsons, err := c.reverseMap.GetIntents(eventSpec)
	if err != nil {
		log.Error("processEvent failed", log.Fields{"err": err, "id": eventId, "agent": agentId})
		return
	}
	// Process events without Agent ID mentioned in policy intent
	eventSpec = event.Event{
		Id:      eventId,
		AgentID: "",
	}
	temp, err := c.reverseMap.GetIntents(eventSpec)
	if err != nil {
		log.Error("processEvent failed", log.Fields{"err": err, "id": eventId, "agent": agentId})
		return
	}
	intentJsons = append(intentJsons, temp...)
	// Execute events for each attached intent
	for _, intentJson := range intentJsons {
		go c.ExecuteEvent(intentJson, agentSpecJson, eventMessageJson)
	}
}
