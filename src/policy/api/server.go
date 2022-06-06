// HTTP Server related functions
//TODO api package may not be right place for this. Need to refactor
package api

import (
	"context"
	"emcopolicy/internal/sacontroller"
	"github.com/gorilla/handlers"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

func StartHTTPServer(ctrl *sacontroller.Controller, wg *sync.WaitGroup) {
	defer wg.Done()
	httpServer := &http.Server{
		Handler: handlers.LoggingHandler(os.Stdout, NewRouter(ctrl)),
		Addr:    ":" + config.GetConfiguration().ServicePort,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Warn("http server exit status", log.Fields{"Error": err})
		}
	}()

	// Graceful shutdown of Mux Server
	// https://github.com/gorilla/mux#graceful-shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	if err := httpServer.Shutdown(context.Background()); err != nil {
		log.Warn("Shutting down httpServer failed.", log.Fields{"err:": err})
	}
}
