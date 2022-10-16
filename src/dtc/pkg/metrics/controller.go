package metrics

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/dtc/pkg/module"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	orchModule "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

func start() {
	go func() {
		orchClient := orchModule.NewClient()
		client := module.NewClient()
		fields := log.Fields{"service": "dtc"}
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

						tgis, err := client.TrafficGroupIntent.GetTrafficGroupIntents(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name)
						if err != nil {
							log.Error(err.Error(), fields)
							continue
						}
						for _, tgi := range tgis {
							fields := fields
							fields["traffic_group_intent"] = tgi.Metadata.Name
							TrafficGroupIntentGauge.WithLabelValues(tgi.Metadata.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name).Set(1)
							iis, err := client.ServerInboundIntent.GetServerInboundIntents(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name, tgi.Metadata.Name)
							if err != nil {
								log.Error(err.Error(), fields)
								continue
							}
							for _, ii := range iis {
								fields := fields
								fields["server_inbound_intent"] = ii.Metadata.Name
								InboundIntentGauge.WithLabelValues(
									ii.Metadata.Name,
									proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name, tgi.Metadata.Name,
									ii.Spec.AppName,
									ii.Spec.AppLabel,
									ii.Spec.ServiceName,
									ii.Spec.ExternalName,
									strconv.Itoa(ii.Spec.Port),
									ii.Spec.Protocol,
									fmt.Sprintf("%t", ii.Spec.ExternalSupport),
									ii.Spec.ServiceMesh,
									ii.Spec.Management.SidecarProxy,
									ii.Spec.Management.TlsType,
								).Set(1)
								iics, err := client.ClientsInboundIntent.GetClientsInboundIntents(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name, tgi.Metadata.Name, ii.Metadata.Name)
								if err != nil {
									log.Error(err.Error(), fields)
									continue
								}
								for _, iic := range iics {
									fields := fields
									fields["inbound_client_intent"] = iic.Metadata.Name
									InboundIntentClientGauge.WithLabelValues(
										iic.Metadata.Name,
										proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name, tgi.Metadata.Name, ii.Metadata.Name,
										iic.Spec.AppName,
										iic.Spec.AppLabel,
										iic.Spec.ServiceName,
									).Set(1)
									aps, err := client.ClientsAccessInboundIntent.GetClientsAccessInboundIntents(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name, tgi.Metadata.Name, ii.Metadata.Name, iic.Metadata.Name)
									if err != nil {
										log.Error(err.Error(), fields)
										continue
									}
									for _, ap := range aps {
										InboundIntentClientAPGauge.WithLabelValues(
											ap.Metadata.Name,
											proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name, tgi.Metadata.Name, ii.Metadata.Name, iic.Metadata.Name,
											ap.Spec.Action,
										).Set(1)
									}
								}
							}
						}
					}
				}
			}

			time.Sleep(time.Duration(15 * time.Second))
		}
	}()
}
