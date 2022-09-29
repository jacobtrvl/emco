package metrics

import (
	"fmt"
	"strconv"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/dtc/pkg/module"
	orchModule "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

func start() {
	go func() {
		orchClient := orchModule.NewClient()
		client := module.NewClient()
		for {
			projects, err := orchModule.NewProjectClient().GetAllProjects()
			if err != nil {
				fmt.Println(err)
				continue
			}

			for _, proj := range projects {
				apps, err := orchClient.CompositeApp.GetAllCompositeApps(proj.MetaData.Name)
				if err != nil {
					fmt.Println(err)
					continue
				}
				for _, app := range apps {

					digs, err := orchClient.DeploymentIntentGroup.GetAllDeploymentIntentGroups(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version)
					if err != nil {
						fmt.Println(err)
						continue
					}
					for _, dig := range digs {
						tgis, err := client.TrafficGroupIntent.GetTrafficGroupIntents(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name)
						if err != nil {
							fmt.Println(err)
							continue
						}
						for _, tgi := range tgis {
							TrafficGroupIntentGauge.WithLabelValues(tgi.Metadata.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name).Set(1)
							iis, err := client.ServerInboundIntent.GetServerInboundIntents(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name, tgi.Metadata.Name)
							if err != nil {
								fmt.Println(err)
								continue
							}
							for _, ii := range iis {
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
									fmt.Println(err)
									continue
								}
								for _, iic := range iics {
									InboundIntentClientGauge.WithLabelValues(
										iic.Metadata.Name,
										proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name, tgi.Metadata.Name, ii.Metadata.Name,
										iic.Spec.AppName,
										iic.Spec.AppLabel,
										iic.Spec.ServiceName,
									).Set(1)
									aps, err := client.ClientsAccessInboundIntent.GetClientsAccessInboundIntents(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name, tgi.Metadata.Name, ii.Metadata.Name, iic.Metadata.Name)
									if err != nil {
										fmt.Println(err)
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
