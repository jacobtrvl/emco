package controller

import (
	"context"
	event "emcopolicy/internal/events"
	"emcopolicy/internal/intent"
	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"sync"
)

func Init() *Controller {
	err := db.InitializeDatabaseConnection("emco")
	//item := make(map[string]map[string][]byte)
	//items := []map[string]map[string][]byte{item}
	// dbTest := db.MockDB{
	//	Items:      items,
	//	Err:        nil,
	//	MarshalErr: nil,
	// }
	if err != nil {
		log.Fatal("Unable to initialize mongo database connection", log.Fields{"Error": err})
	}
	// DB connection is a package level variable (db.DBconn) in orchestrator db package.
	// Scoping this to the client context for better readability
	c := &Controller{
		db: db.DBconn,
		//db:        &dbTest,
		tag:       "data",
		storeName: "resources",
		eventList: &EventList{
			eventMap: make(map[event.Event][]intent.Intent),
			mutex:    sync.RWMutex{},
		},
		updateStream: make(chan intent.StreamData),
		eventStream:  make(chan event.Event),
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
		EventStream: c.eventStream,
	})
	return c
}

func (c *Controller) StartScheduler() error {
	c.eventList = new(EventList)
	c.eventList.eventMap = make(map[event.Event][]intent.Intent)
	c.storeName = "resources"
	c.tag = "data"
	err := c.BuildEventListFromDB(context.Background())
	if err != nil {
		return errors.Wrap(err, "Starting scheduler failed::")
	}
	go func() {
		err := c.scheduler(context.Background())
		if err != nil {
			log.Warn("Scheduler exited", log.Fields{"Reason": err})
		}
	}()

	return nil
}

func (c Controller) PolicyClient() *intent.Client {
	return c.policyClient
}

func (c Controller) EventClient() *event.Client {
	return c.eventClient
}
