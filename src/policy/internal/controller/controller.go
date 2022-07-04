package controller

import (
	"context"
	event "emcopolicy/internal/events"
	"emcopolicy/internal/intent"
	events "emcopolicy/pkg/grpc"
	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"sync"
)

func Init() (*Controller, error) {
	err := db.InitializeDatabaseConnection("emco")
	if err != nil {
		return nil, errors.Errorf("Unable to initialize mongo database connection: %s", err)
	}
	/* TODO For unit testing. Remove once proper unit test code is in
	item := make(map[string]map[string][]byte)
	items := []map[string]map[string][]byte{item}
	dbTest := db.MockDB{
		Items:      items,
		Err:        nil,
		MarshalErr: nil,
	} */

	// DB connection is a package level variable (db.DBconn) in orchestrator db package.
	// Scoping this to the client context for better readability
	c := &Controller{
		db: db.DBconn,
		//db:        &dbTest,
		tag:       "data",
		storeName: "resources",
		reverseMap: &ReverseMap{
			eventMap: make(map[event.Event][]intent.Intent),
			mutex:    sync.RWMutex{},
		},
		agentMap: &AgentMap{
			runtime: make(map[AgentID]*AgentRuntime),
			mutex:   sync.RWMutex{},
		},
		updateStream:    make(chan intent.StreamData),
		eventStream:     make(chan event.Event),
		agentStream:     make(chan event.StreamAgentData),
		requireRecovery: make(chan any),
		eventsQueue:     make(chan *events.Event, 100),
	}
	c.policyClient = intent.NewClient(intent.Config{
		Db:           c.db,
		Tag:          c.tag,
		StoreName:    c.storeName,
		UpdateStream: c.updateStream,
	})
	c.eventClient = event.NewClient(event.Config{
		Db:          c.db,
		Tag:         c.tag,
		StoreName:   c.storeName,
		AgentStream: c.agentStream,
	})

	key := Module{"Agent"}
	err = c.db.Insert(c.storeName, key, nil, c.tag, key)
	if err != nil {
		return nil, errors.Errorf("Error while Initializing DB %s", err)
	}
	return c, nil
}

func (c *Controller) StartScheduler(ctx context.Context) error {
	err := c.BuildReverseMap(ctx)
	if err != nil {
		return errors.Wrap(err, "Starting OperationalScheduler failed")
	}
	err = c.BuildAgentMap(ctx)
	if err != nil {
		return errors.Wrap(err, "Starting OperationalScheduler failed")
	}
	go c.OperationalScheduler(ctx)
	go c.AgentManager(ctx)
	go c.EventsManager(ctx)
	return nil
}

func (c Controller) PolicyClient() *intent.Client {
	return c.policyClient
}

func (c Controller) EventClient() *event.Client {
	return c.eventClient
}
