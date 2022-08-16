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

package controller

import (
	clientSet "edgemetricscollector/pkg/generated/clientset/versioned"
	listers "edgemetricscollector/pkg/generated/listers/edgemetricscollector/v1alpha1"
	"encoding/json"
	events "gitlab.com/project-emco/core/emco-base/src/policy/pkg/grpc"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

// Controller is the k8s custom controller implementation for MetricsController resources
type Controller struct {
	kubeClientSet      kubernetes.Interface
	collectorClientSet clientSet.Interface
	collectorLister    listers.MetricsCollectorLister
	collectorSynced    cache.InformerSynced
	podLister          corelisters.PodLister
	podsSynced         cache.InformerSynced
	workQueue          workqueue.DelayingInterface
	recorder           record.EventRecorder
	metricsStream      chan<- Measurements
}

// StreamServer holds the data required for gprc server
// metricsStream  is the channel that base controller (metrics fetch part) and
// grpc server (metrics send to EMCO) communicate.
type StreamServer struct {
	events.EventsServer
	metricsStream <-chan Measurements
	agentId       string
}

// Measurements is base structure to send to EMCO policy controller
type Measurements struct {
	Pod       string          `json:"pod"`
	Namespace string          `json:"namespace"`
	ContextId string          `json:"contextId"`
	AppName   string          `json:"appName"`
	Metric    string          `json:"metric"`
	Value     json.RawMessage `json:"value"`
}
