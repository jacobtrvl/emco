// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package main

import (
	"math/rand"
	"os"
	"time"

	register "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	contextDb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"

	"gitlab.com/project-emco/core/emco-base/src/hpa-ac/pkg/grpc/contextupdateserver"

	"context"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	err := db.InitializeDatabaseConnection(context.Background(), "emco")
	if err != nil {
		log.Error("Unable to initialize mongo database connection", log.Fields{"Error": err})
		os.Exit(1)
	}
	err = contextDb.InitializeContextDatabase()
	if err != nil {
		log.Error("Unable to initialize etcd database connection", log.Fields{"Error": err})
		os.Exit(1)
	}

	err = register.StartGrpcServer("hpaaction", "HPAACTION_NAME", 9042,
		register.RegisterContextUpdateService, contextupdateserver.NewContextupdateServer())
	if err != nil {
		log.Error("GRPC server failed to start", log.Fields{"Error": err})
		os.Exit(1)
	}
}
