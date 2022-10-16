package metrics

import (
	"context"
	"time"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	orchModule "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/module"
)

func start() {
	go func() {
		orchClient := orchModule.NewClient()
		client := module.NewClient()
		fields := log.Fields{"service": "sfc"}
		for {
			projects, err := orchModule.NewProjectClient().GetAllProjects(context.Background())
			if err != nil {
				log.Error(err.Error(), fields)
				continue
			}

			for _, proj := range projects {
				fields := fields
				fields["project"] = proj.MetaData.Name

				apps, err := orchClient.CompositeApp.GetAllCompositeApps(context.Background(), proj.MetaData.Name)
				if err != nil {
					log.Error(err.Error(), fields)
					continue
				}
				for _, app := range apps {
					fields := fields
					fields["composite_app"] = app.Metadata.Name
					digs, err := orchClient.DeploymentIntentGroup.GetAllDeploymentIntentGroups(context.Background(), proj.MetaData.Name, app.Metadata.Name, app.Spec.Version)
					if err != nil {
						log.Error(err.Error(), fields)
						continue
					}
					for _, dig := range digs {
						fields := fields
						fields["dig"] = dig.MetaData.Name

						sfcs, err := client.SfcIntent.GetAllSfcIntents(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name)
						if err != nil {
							log.Error(err.Error(), fields)
							continue
						}
						for _, sfc := range sfcs {
							fields := fields
							fields["sfc"] = sfc.Metadata.Name
							SFCIntentGauge.WithLabelValues(
								sfc.Metadata.Name,
								proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name,
								sfc.Spec.ChainType,
								sfc.Spec.Namespace,
							).Set(1)
							links, err := client.SfcLinkIntent.GetAllSfcLinkIntents(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name, sfc.Metadata.Name)
							if err != nil {
								log.Error(err.Error(), fields)
								continue
							}
							for _, link := range links {
								SFCIntentLinkGauge.WithLabelValues(
									link.Metadata.Name,
									proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name, sfc.Metadata.Name,
									link.Spec.LeftNet,
									link.Spec.RightNet,
									link.Spec.LinkLabel,
									link.Spec.AppName,
									link.Spec.WorkloadResource,
									link.Spec.ResourceType,
								).Set(1)
							}
						}

					}
				}
			}

			time.Sleep(time.Duration(15 * time.Second))
		}
	}()
}
