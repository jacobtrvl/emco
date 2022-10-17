package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

func Start() {
	go func() {
		client := module.NewClient()
		fields := log.Fields{"service": "orchestrator"}
		for {
			if err := handleControllers(ControllerGauge, client); err != nil {
				log.Error(err.Error(), fields)
			}

			projects, err := module.NewProjectClient().GetAllProjects(context.Background())
			if err != nil {
				log.Error(err.Error(), fields)
				continue
			}

			for _, proj := range projects {
				fields := fields
				fields["project"] = proj.MetaData.Name

				ProjectGauge.WithLabelValues(proj.MetaData.Name).Set(1)
				apps, err := client.CompositeApp.GetAllCompositeApps(context.Background(), proj.MetaData.Name)
				if err != nil {

					log.Error(err.Error(), fields)
					continue
				}
				for _, app := range apps {
					fields := fields
					fields["composite_app"] = app.Metadata.Name

					ComAppGauge.WithLabelValues(app.Spec.Version, app.Metadata.Name, proj.MetaData.Name).Set(1)

					applications, err := client.App.GetApps(context.Background(), proj.MetaData.Name, app.Metadata.Name, app.Spec.Version)
					if err != nil {
						log.Error(err.Error(), fields)
						continue
					}
					for _, application := range applications {
						fields := fields
						fields["app"] = application.Metadata.Name

						AppGauge.WithLabelValues(application.Metadata.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version).Set(1)

						dependencies, err := client.AppDependency.GetAllAppDependency(context.Background(), proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, application.Metadata.Name)
						if err != nil {
							log.Error(err.Error(), fields)
							continue
						}

						for _, dependency := range dependencies {
							DependencyGauge.WithLabelValues(dependency.MetaData.Name, proj.MetaData.Name, application.Metadata.Name, app.Metadata.Name, app.Spec.Version).Set(1)
						}
					}

					digs, err := client.DeploymentIntentGroup.GetAllDeploymentIntentGroups(context.Background(), proj.MetaData.Name, app.Metadata.Name, app.Spec.Version)
					if err != nil {
						log.Error(err.Error(), fields)
						continue
					}
					for _, dig := range digs {
						fields := fields
						fields["dig"] = dig.MetaData.Name

						DIGGauge.WithLabelValues(dig.MetaData.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version).Set(1)
						gpIntents, err := client.GenericPlacementIntent.GetAllGenericPlacementIntents(context.Background(), proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name)
						if err != nil {
							log.Error(err.Error(), fields)
							continue
						}
						for _, gpi := range gpIntents {
							fields := fields
							fields["gpi"] = dig.MetaData.Name

							GenericPlacementIntentGauge.WithLabelValues(gpi.MetaData.Name, dig.MetaData.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version).Set(1)
							appIntents, err := client.AppIntent.GetAllAppIntents(context.Background(), proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, gpi.MetaData.Name, dig.MetaData.Name)
							if err != nil {
								log.Error(err.Error(), fields)
								continue
							}
							for _, appIntent := range appIntents {
								GenericAppPlacementIntentGauge.WithLabelValues(appIntent.MetaData.Name, gpi.MetaData.Name, dig.MetaData.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version).Set(1)
							}
						}

						groupIntents, err := client.Intent.GetAllIntents(context.Background(), proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name)
						if err != nil {
							log.Error(err.Error(), fields)
							continue
						}
						for _, groupIntent := range groupIntents.ListOfIntents {
							if name, ok := groupIntent["genericPlacementIntent"]; ok {
								GroupIntentGauge.WithLabelValues(name, dig.MetaData.Name, app.Spec.Version, app.Metadata.Name, proj.MetaData.Name).Set(1)
							}
						}
					}

					comProfiles, err := client.CompositeProfile.GetCompositeProfiles(context.Background(), proj.MetaData.Name, app.Metadata.Name, app.Spec.Version)
					if err != nil {
						log.Error(err.Error(), fields)
						continue
					}
					for _, comProfile := range comProfiles {
						fields := fields
						fields["composite_profile"] = comProfile.Metadata.Name

						CompositeProfileGauge.WithLabelValues(comProfile.Metadata.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version).Set(1)
						appProfiles, err := client.AppProfile.GetAppProfiles(context.Background(), proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, comProfile.Metadata.Name)
						if err != nil {
							log.Error(err.Error(), fields)
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
	controllers, err := client.Controller.GetControllers(context.Background())
	if err != nil {
		return err
	}
	p.Reset()
	for _, c := range controllers {
		p.WithLabelValues(c.Metadata.Name).Set(1)
	}
	return nil
}
