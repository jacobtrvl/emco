package main

import (
	"context"
	"emcopolicy/internal/controller"
	"emcopolicy/pkg/http"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"sync"
)

func main() {
	log.Info("Starting Policy Controller", log.Fields{})
	wg := new(sync.WaitGroup)
	// Create Controller context and start scheduler.
	// Scheduler should start before the api & event server
	c, err := controller.Init()
	if err != nil {
		log.Fatal("Policy controller init failed", log.Fields{"Err": err})
	}
	if c == nil {
		log.Fatal("Policy controller init failed. Controller is nil", log.Fields{})
	}
	if err = c.StartScheduler(context.Background()); err != nil {
		log.Fatal("Scheduler failed to start", log.Fields{"err": err})
	}
	// HTTP Server Initialization
	wg.Add(1)
	go http.StartHTTPServer(c, wg)
	wg.Wait()
}
