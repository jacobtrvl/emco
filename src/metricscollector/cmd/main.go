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
	"edgemetricscollector/internal/controller"
	"flag"
	"math/rand"
	"strconv"
	"time"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	clientset "edgemetricscollector/pkg/generated/clientset/versioned"
	informers "edgemetricscollector/pkg/generated/informers/externalversions"
	"edgemetricscollector/pkg/signals"
)

var (
	masterURL   string
	kubeConfig  string
	workersFlag string
	agentId     string
	port        string
)

const (
	rsyncDuration = time.Second * 30
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	// AgentId, should be unique across the clusters that are managed by an instance of EMCO
	// This should be provided by the user, during the deployment of agent
	if len(agentId) == 0 {
		klog.Fatalf("AgentId is mandatory. Provide a unique name as agentId")
	}
	workers, err := strconv.Atoi(workersFlag)
	if err != nil || workers < 1 {
		workers = 1
	}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	metricsClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building example clientset: %s", err.Error())
	}

	metricsStream := make(chan controller.Measurements)
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, rsyncDuration)
	metricsInformerFactory := informers.NewSharedInformerFactory(metricsClient, rsyncDuration)
	podInformer := kubeInformerFactory.Core().V1().Pods()
	c := controller.NewController(kubeClient, metricsClient,
		podInformer,
		metricsInformerFactory.Edgemetricscollector().V1alpha1().MetricsCollectors(),
		metricsStream)

	kubeInformerFactory.Start(stopCh)
	metricsInformerFactory.Start(stopCh)
	go controller.MetricsServer(metricsStream, agentId, port)
	if err = c.Run(workers, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.StringVar(&kubeConfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&workersFlag, "workers", "1", "Number of workers to run. Default: 1")
	flag.StringVar(&agentId, "agentid", "", "AgentID: Unique ID for this agent. EMCO identifies metrics using this id.")
	flag.StringVar(&port, "port", "9091", "port: Port number this agent listens.Default: 9091")
}
