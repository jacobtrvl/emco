package controller

import (
	"context"
	event "emcopolicy/internal/events"
	"emcopolicy/internal/intent"
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"time"
)

// BuildReverseMap builds a reverse map of events to policy.
// Should be called only during boot up, when no other controller threads are running
func (c *Controller) BuildReverseMap(ctx context.Context) error {
	policyIntents, err := c.getAllPolicyIntents(ctx)
	if err != nil {
		return errors.Wrap(err, "Building In-memory ReverseMap failed")
	}
	c.reverseMap.mutex.Lock()
	defer c.reverseMap.mutex.Unlock()
	for _, policyIntent := range policyIntents {
		if c.reverseMap.eventMap[policyIntent.Spec.Event] == nil {
			c.reverseMap.eventMap[policyIntent.Spec.Event] = []intent.Intent{}
		}
		c.reverseMap.eventMap[policyIntent.Spec.Event] = append(c.reverseMap.eventMap[policyIntent.Spec.Event], policyIntent)
	}
	return nil
}

// BuildAgentMap builds a reverse map of events to policy.
// Should be called only during boot up, when no other controller threads are running
func (c *Controller) BuildAgentMap(ctx context.Context) error {
	agents, err := c.getAllAgents(ctx)
	if err != nil {
		return errors.Wrap(err, "BuildAgentMap failed")
	}
	if agents == nil {
		log.Warn("No Agents found in DB", log.Fields{})
		return nil
	}
	c.agentMap.mutex.Lock()
	defer c.agentMap.mutex.Unlock()
	for _, agent := range agents {
		runtime := &AgentRuntime{
			spec: agent,
		}
		c.agentMap.runtime[AgentID(agent.Id)] = runtime
	}
	return nil
}

// AddPolicyIntent adds an intent from event's intentList in eventMap
func (m *ReverseMap) AddPolicyIntent(e event.Event, p intent.Intent) {
	var index int
	m.mutex.Lock()
	defer m.mutex.Unlock()
	// Remove Intent from list if this is an update.
	// Deletion can be handled by  RemovePolicyIntent, but it's safer to do while lock is holding
	// This lock could starve other appends and reads, but it is unavoidable in current design.
	// This delay won't be user visible since this is a background activity
	// Also appends will be less often compared to reads
	for index < len(m.eventMap[e]) && !isEqual(m.eventMap[e][index], p) {
		index++
	}
	if index < len(m.eventMap[e]) {
		m.eventMap[e] = append(m.eventMap[e][:index], m.eventMap[e][index+1:]...)
	}
	if _, ok := m.eventMap[e]; !ok {
		m.eventMap[e] = []intent.Intent{}
	}
	m.eventMap[e] = append(m.eventMap[e], p)
	log.Info("Added Intent to in-memory event list", log.Fields{"Event": e.Id, "Intent": p.Spec.PolicyIntentID})
}

// RemovePolicyIntent deletes an intent from event's intentList in eventMap
func (m *ReverseMap) RemovePolicyIntent(e event.Event, p intent.Intent) error {
	var index int
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for index < len(m.eventMap[e]) && !isEqual(m.eventMap[e][index], p) {
		index++
	}
	if index == len(m.eventMap[e]) {
		i := p.Spec.Project + "/" + p.Spec.CompositeApp + "/" + p.Spec.CompositeAppVersion + "/" +
			p.Spec.DeploymentIntentGroup + "/" + p.Spec.PolicyIntentID
		return errors.Errorf("RemoveIntent failed. No Policy Intent "+
			"(%s) found in the IntentList of event %s", i, e.Id)
	}
	m.eventMap[e] = append(m.eventMap[e][:index], m.eventMap[e][index+1:]...)
	log.Info("Removed Intent from in-memory event list", log.Fields{"Event": e.Id, "Intent": p.Spec.PolicyIntentID})
	return nil
}

// GetPolicyIntentList returns intentList of an event from eventMap
func (m *ReverseMap) GetPolicyIntentList(e event.Event) (error, []intent.Intent) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if _, ok := m.eventMap[e]; !ok {
		// TODO error
		return errors.Errorf("GetPolicyIntentList failed. Event %s not found", e.Id), nil
	}
	intentCopy := make([]intent.Intent, len(m.eventMap[e]))
	// slice can get modified from other threads
	// Hence returning a snapshot
	copy(intentCopy, m.eventMap[e])
	return nil, intentCopy
}

func (m *ReverseMap) GetIntents(event event.Event) ([][]byte, error) {
	var (
		intents [][]byte
	)
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	for _, i := range m.eventMap[event] {
		data, err := json.Marshal(i.Spec)
		intents = append(intents, data)
		if err != nil {
			return [][]byte{}, errors.Wrap(err, "GetIntents failed")
		}
	}
	return intents, nil
}
func (m *AgentMap) Cancel(id AgentID) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if _, ok := m.runtime[id]; ok {
		m.runtime[id].cancel()
	}
}

func (m *AgentMap) IsAgentExists(id AgentID) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	_, ok := m.runtime[id]
	return ok
}

func (m *AgentMap) DeleteAgent(id AgentID) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, ok := m.runtime[id]; ok {
		delete(m.runtime, id)
	}
}

func (m *AgentMap) VerifySpec(id AgentID, spec event.AgentSpec) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	val, ok := m.runtime[id]
	return ok && cmp.Equal(spec, val.spec)
}

func (m *AgentMap) UpdateSpec(id AgentID, spec event.AgentSpec) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.runtime[id] = &AgentRuntime{
		spec: spec,
	}
}

func (m *AgentMap) MarkForRecovery(id AgentID) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if runtime, ok := m.runtime[id]; ok {
		//A small delay to avoid busy looping in case of agent errors
		time.Sleep(time.Second)
		runtime.isRunning = false
	}
}

func isEqual(p, q intent.Intent) bool {
	return p.Spec.Project == q.Spec.Project &&
		p.Spec.CompositeApp == q.Spec.CompositeApp &&
		p.Spec.CompositeAppVersion == q.Spec.CompositeAppVersion &&
		p.Spec.DeploymentIntentGroup == q.Spec.DeploymentIntentGroup &&
		p.Spec.PolicyIntentID == q.Spec.PolicyIntentID
}
