package main

import (
	"emcopolicy/internal/controller"
	"emcopolicy/pkg/http"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"sync"
)

func main() {
	log.Info("Starting Policy Controller", log.Fields{})
	wg := new(sync.WaitGroup)
	//Create Controller context and start scheduler.
	//Scheduler should start before the api & event server
	c := controller.Init()
	err := c.StartScheduler()
	if err != nil {
		log.Error("Scheduler failed to start", log.Fields{"err": err})
		return
	}

	//HTTP Server Initialization
	wg.Add(1)
	go http.StartHTTPServer(c, wg)
	wg.Wait()
}
