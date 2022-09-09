package metricscontroller

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/metrics"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

func Start(m *metrics.Metrics) {
	ComAppGauge := m.Gauges["ComAppGauge"]
	ProjectGauge := m.Gauges["ProjectGauge"]
	ControllerGauge := m.Gauges["ControllerGauge"]
	DIGGauge := m.Gauges["DIGGauge"]
	GenericPlacementIntentGauge := m.Gauges["GenericPlacementIntentGauge"]
	CompositeProfileGauge := m.Gauges["CompositeProfileGauge"]
	AppProfileGauge := m.Gauges["AppProfileGauge"]
	GenericAppPlacementIntentGauge := m.Gauges["GenericAppPlacementIntentGauge"]
	GroupIntentGauge := m.Gauges["GroupIntentGauge"]

	go func() {
		client := module.NewClient()
		for {
			if err := handleControllers(ControllerGauge, client); err != nil {
				fmt.Println(err)
			}

			projects, err := module.NewProjectClient().GetAllProjects()
			if err != nil {
				fmt.Println(err)
				continue
			}

			for _, proj := range projects {
				ProjectGauge.WithLabelValues(proj.MetaData.Name).Set(1)
				apps, err := client.CompositeApp.GetAllCompositeApps(proj.MetaData.Name)
				if err != nil {
					fmt.Println(err)
					continue
				}
				for _, app := range apps {
					ComAppGauge.WithLabelValues(app.Spec.Version, app.Metadata.Name, proj.MetaData.Name).Set(1)
					digs, err := client.DeploymentIntentGroup.GetAllDeploymentIntentGroups(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version)
					if err != nil {
						fmt.Println(err)
						continue
					}
					for _, dig := range digs {
						DIGGauge.WithLabelValues(dig.MetaData.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version).Set(1)
						gpIntents, err := client.GenericPlacementIntent.GetAllGenericPlacementIntents(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name)
						if err != nil {
							fmt.Println(err)
							continue
						}
						for _, gpi := range gpIntents {
							GenericPlacementIntentGauge.WithLabelValues(gpi.MetaData.Name, dig.MetaData.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version).Set(1)
							appIntents, err := client.AppIntent.GetAllAppIntents(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, gpi.MetaData.Name, dig.MetaData.Name)
							if err != nil {
								fmt.Println(err)
								continue
							}
							for _, appIntent := range appIntents {
								GenericAppPlacementIntentGauge.WithLabelValues(appIntent.MetaData.Name, gpi.MetaData.Name, dig.MetaData.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version).Set(1)
							}
						}

						groupIntents, err := client.Intent.GetAllIntents(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name)
						if err != nil {
							fmt.Println(err)
							continue
						}
						for _, groupIntent := range groupIntents.ListOfIntents {
							if name, ok := groupIntent["genericPlacementIntent"]; ok {
								GroupIntentGauge.WithLabelValues(name, dig.MetaData.Name, app.Spec.Version, app.Metadata.Name, proj.MetaData.Name).Set(1)
							}
						}
					}

					comProfiles, err := client.CompositeProfile.GetCompositeProfiles(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version)
					if err != nil {
						fmt.Println(err)
						continue
					}
					for _, comProfile := range comProfiles {
						CompositeProfileGauge.WithLabelValues(comProfile.Metadata.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version).Set(1)
						appProfiles, err := client.AppProfile.GetAppProfiles(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, comProfile.Metadata.Name)
						if err != nil {
							fmt.Println(err)
							continue
						}
						for _, appProfile := range appProfiles {
							AppProfileGauge.WithLabelValues(appProfile.Metadata.Name, comProfile.Metadata.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version).Set(1)
						}
					}
				}
			}
			time.Sleep(time.Duration(15 * time.Second))
		}
	}()
}

func handleControllers(p *prometheus.GaugeVec, client *module.Client) error {
	controllers, err := client.Controller.GetControllers()
	if err != nil {
		return err
	}
	p.Reset()
	for _, c := range controllers {
		p.WithLabelValues(c.Metadata.Name).Set(1)
	}
	return nil
}
