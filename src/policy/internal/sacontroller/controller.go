package sacontroller

import (
	event "emcopolicy/internal/events"
	"emcopolicy/internal/policy"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

type Controller struct {
	policyClient *policy.Client
	eventClient  *event.Client
}

func InitController() *Controller {
	err := db.InitializeDatabaseConnection("emco")
	if err != nil {
		log.Fatal("Unable to initialize mongo database connection", log.Fields{"Error": err})
	}
	// DB connection is a package level variable (db.DBconn) in orchestrator db package.
	// Scoping this to the client context for better readability
	return &Controller{
		policyClient: policy.NewClient(db.DBconn),
		eventClient:  event.NewClient(db.DBconn),
	}
}

// InitTestController can be used for Local Testing
func InitTestController() *Controller {
	//TODO Test with mock mongo
	return &Controller{
		policyClient: policy.NewClient(nil),
		eventClient:  event.NewClient(nil),
	}
}

func (c Controller) PolicyClient() *policy.Client {
	return c.policyClient
}

func (c Controller) EventClient() *event.Client {
	return c.eventClient
}
