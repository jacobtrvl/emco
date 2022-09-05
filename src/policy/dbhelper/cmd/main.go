//=======================================================================
// Copyright (c) 2022 Aarna Networks, Inc.
// All rights reserved.
// ======================================================================
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//           http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ========================================================================

package main

import (
	"context"
	"dbhelper/api"
	etcdhelper "dbhelper/etcd"
	"github.com/gorilla/handlers"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

// dbhelper will run as a separate service on same container.
// Hard-coding the port as a workaround.
const portToListen = "9090"

// This is a workaround for backward compatibility issue of etcd library.
// It is having conflict with generated grpc files
// This issue is resolved in etcd 3.5.
func main() {
	etcdCfg := etcdhelper.EtcdConfig{
		Endpoint: config.GetConfiguration().EtcdIP,
		CertFile: config.GetConfiguration().EtcdCert,
		KeyFile:  config.GetConfiguration().EtcdKey,
		CAFile:   config.GetConfiguration().EtcdCAFile,
	}
	etcdClient, err := etcdhelper.NewEtcdClient(nil, etcdCfg)

	if err != nil {
		log.Error("Context DB Error", log.Fields{"error": err})
		return
	}
	// HTTP Server Initialization
	wg := new(sync.WaitGroup)
	wg.Add(1)

	go StartHTTPServer(etcdClient, wg)
	wg.Wait()

}

func StartHTTPServer(ctrl etcdhelper.ContextDb, wg *sync.WaitGroup) {
	defer wg.Done()
	httpServer := &http.Server{
		Handler: handlers.LoggingHandler(os.Stdout, api.NewRouter(ctrl)),
		Addr:    portToListen,
	}
	go func() {
		log.Info("Starting HTTP Server", log.Fields{"port": httpServer.Addr})
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
