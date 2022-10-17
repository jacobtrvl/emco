// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	infra "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/metrics"
)

type Metrics struct {
	Gauges  map[string]*prometheus.GaugeVec
	Handler http.Handler
}

func Initialize(startNow bool) *Metrics {
	metrics := &Metrics{
		Gauges: make(map[string]*prometheus.GaugeVec),
	}
	metrics.Gauges["ComAppGauge"] = ComAppGauge
	metrics.Gauges["ProjectGauge"] = ProjectGauge
	metrics.Gauges["ControllerGauge"] = ControllerGauge
	metrics.Gauges["DIGGauge"] = DIGGauge
	metrics.Gauges["GenericPlacementIntentGauge"] = GenericPlacementIntentGauge
	metrics.Gauges["CompositeProfileGauge"] = CompositeProfileGauge
	metrics.Gauges["AppProfileGauge"] = AppProfileGauge
	metrics.Gauges["GenericAppPlacementIntentGauge"] = GenericAppPlacementIntentGauge
	metrics.Gauges["GroupIntentGauge"] = GroupIntentGauge
	metrics.Gauges["AppGauge"] = AppGauge
	metrics.Gauges["DependencyGauge"] = DependencyGauge

	prometheus.MustRegister(ComAppGauge)
	prometheus.MustRegister(ProjectGauge)
	prometheus.MustRegister(ControllerGauge)
	prometheus.MustRegister(DIGGauge)
	prometheus.MustRegister(GenericPlacementIntentGauge)
	prometheus.MustRegister(CompositeProfileGauge)
	prometheus.MustRegister(AppProfileGauge)
	prometheus.MustRegister(GenericAppPlacementIntentGauge)
	prometheus.MustRegister(GroupIntentGauge)
	prometheus.MustRegister(AppGauge)
	prometheus.MustRegister(DependencyGauge)

	prometheus.MustRegister(infra.NewBuildInfoCollector("orchestrator"))

	if startNow {
		start()
	}
	metrics.Handler = promhttp.Handler()
	return metrics
}
