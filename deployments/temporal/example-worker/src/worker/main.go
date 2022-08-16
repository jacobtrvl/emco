package main

// Worker process for the workflow.
// Registers app-specific workflow and activity code, then runs them.

import (
	"fmt"
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	app "example-worker/src/greeting"
)

const (
	temporal_env_var = "TEMPORAL_SERVER"
	temporal_port    = "7233"
)

func main() {
	// Get the Temporal Server's IP
	temporal_server := os.Getenv(temporal_env_var)
	if temporal_server == "" {
		fmt.Fprintf(os.Stderr, "Error: Need to define $TEMPORAL_SERVER\n")
		os.Exit(1)
	}
	hostPort := temporal_server + ":" + temporal_port
	fmt.Printf("Temporal server endpoint: (%s)\n", hostPort)

	// Create the client object just once per process
	options := client.Options{HostPort: hostPort}
	c, err := client.NewLazyClient(options)
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()
	// This worker hosts both Workflow and Activity functions
	w := worker.New(c, app.GreetingTaskQueue, worker.Options{})
	w.RegisterWorkflow(app.GreetingWorkflow)
	w.RegisterActivity(app.ComposeGreeting)
	// Start listening to the Task Queue
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("unable to start Worker", err)
	}

}
