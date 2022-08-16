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
	"emcopolicy/internal/controller"
	"emcopolicy/pkg/http"
	"flag"
	"github.com/sirupsen/logrus"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"sync"
)

func main() {
	// Currently, we are reading workflow manager IP from command line
	// We are supporting only workflow manager now, and hence this parameter is must
	// But when we support more plugins, this need not be mandatory and should read from config file
	workflowManagerUrl := flag.String("workflowmgr", "", "EMCO workflow manager endpoint")
	flag.Parse()
	if len(*workflowManagerUrl) == 0 {
		log.Fatal("workflowmgr parameter is mandatory. See help for details", log.Fields{})
	}
	log.SetLoglevel(logrus.InfoLevel)
	log.Info("Starting Policy Controller", log.Fields{})
	wg := new(sync.WaitGroup)
	// Create Controller context and start scheduler.
	// Scheduler should start before the api & event server
	c, err := controller.Init(*workflowManagerUrl)
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
