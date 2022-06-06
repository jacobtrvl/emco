package main

import (
	"emcopolicy/api"
	"emcopolicy/internal/sacontroller"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"sync"
)

var unitTest bool

func main() {
	log.Info("Starting Policy SA controller", log.Fields{})
	wg := new(sync.WaitGroup)
	//Create controller context
	// TODO Remove unitTest part after initial dev
	controller := sacontroller.InitTestController()
	if !unitTest {
		controller = sacontroller.InitController()
	}
	//HTTP Server Initialization
	wg.Add(1)
	go api.StartHTTPServer(controller, wg)
	wg.Wait()
}
