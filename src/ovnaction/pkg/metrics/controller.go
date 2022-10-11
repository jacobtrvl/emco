package metrics

import (
	"context"
	"fmt"
	"time"

	orchModule "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/ovnaction/pkg/module"
)

func start() {
	go func() {
		orchClient := orchModule.NewClient()
		client := module.NewClient()
		for {
			projects, err := orchModule.NewProjectClient().GetAllProjects(context.Background())
			if err != nil {
				fmt.Println(err)
				continue
			}

			for _, proj := range projects {
				apps, err := orchClient.CompositeApp.GetAllCompositeApps(context.Background(), proj.MetaData.Name)
				if err != nil {
					fmt.Println(err)
					continue
				}
				for _, app := range apps {
					digs, err := orchClient.DeploymentIntentGroup.GetAllDeploymentIntentGroups(context.Background(), proj.MetaData.Name, app.Metadata.Name, app.Spec.Version)
					if err != nil {
						fmt.Println(err)
						continue
					}
					for _, dig := range digs {
						ncis, err := client.NetControlIntent.GetNetControlIntents(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name)
						if err != nil {
							fmt.Println(err)
							continue
						}
						for _, nci := range ncis {
							NetworkControllerIntentGauge.WithLabelValues(nci.Metadata.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name).Set(1)
							wis, err := client.WorkloadIntent.GetWorkloadIntents(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name, nci.Metadata.Name)
							if err != nil {
								fmt.Println(err)
								continue
							}
							for _, wi := range wis {
								WorkloadIntentGauge.WithLabelValues(wi.Metadata.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name, nci.Metadata.Name, wi.Spec.AppName, wi.Spec.WorkloadResource, wi.Spec.Type).Set(1)
								wiifs, err := client.WorkloadIfIntent.GetWorkloadIfIntents(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name, nci.Metadata.Name, wi.Metadata.Name)
								if err != nil {
									fmt.Println(err)
									continue
								}
								for _, wiif := range wiifs {
									WorkloadInterfaceIntentGauge.WithLabelValues(
										wiif.Metadata.Name,
										proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name, nci.Metadata.Name, wi.Metadata.Name,
										wiif.Spec.IfName,
										wiif.Spec.NetworkName,
										wiif.Spec.DefaultGateway,
										wiif.Spec.IpAddr,
										wiif.Spec.MacAddr,
									).Set(1)
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
