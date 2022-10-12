package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	nameLabel                       = []string{"name"}
	comAppGaugeLabels               = []string{"project", "composite_app", "composite_app_version"}
	groupIntentGaugeLabels          = []string{"name", "deployment_intent_group", "composite_app_version", "composite_app_name", "project"}
	nameWithComAppLabels            = []string{"name", "project", "composite_app", "composite_app_version"}
	dependencyLabels                = []string{"name", "project", "app", "composite_app", "composite_app_version"}
	appProfileLabels                = []string{"name", "composite_profile", "project", "composite_app", "composite_app_version"}
	genericPlacementIntentLabels    = []string{"name", "deployment_intent_group", "project", "composite_app", "composite_app_version"}
	genericAppPlacementIntentLabels = []string{"name", "generic_placement_intent", "deployment_intent_group", "project", "composite_app", "composite_app_version"}
)

var ControllerGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_controller_resource",
	Help: "Count of Controllers",
}, nameLabel)

var ProjectGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_project_resource",
	Help: "Count of Projects",
}, nameLabel)

var ComAppGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_composite_app_resource",
	Help: "Count of Composite Apps",
}, comAppGaugeLabels)

var AppGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_app_resource",
	Help: "Count of Apps",
}, nameWithComAppLabels)

var DependencyGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_dependency_resource",
	Help: "Count of Dependencies",
}, dependencyLabels)

var DIGGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_deployment_intent_group_resource",
	Help: "Count of Deployment Intent Groups",
}, nameWithComAppLabels)

var GenericPlacementIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_generic_placement_intent_resource",
	Help: "Count of Generic Placement Intents",
}, genericPlacementIntentLabels)

var GenericAppPlacementIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_generic_app_placement_intent_resource",
	Help: "Count of Generic App Placement Intents",
}, genericAppPlacementIntentLabels)

var CompositeProfileGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_composite_profile_resource",
	Help: "Count of Composite Profiles",
}, nameWithComAppLabels)

var AppProfileGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_app_profile_resource",
	Help: "Count of App Profiles",
}, appProfileLabels)

var GroupIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_group_intent_resource",
	Help: "Count of Group Intents",
}, groupIntentGaugeLabels)
