package metrics

import (
	"context"
	"fmt"
	"strconv"
	"time"

	orchModule "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/tac/pkg/module"
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
						tacs, err := client.WorkflowIntentClient.GetWorkflowHookIntents(proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name)
						if err != nil {
							fmt.Println(err)
							continue
						}
						for _, tac := range tacs {
							TACIntentGauge.WithLabelValues(
								tac.Metadata.Name,
								proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name,
								tac.Spec.HookType,
								tac.Spec.WfClientSpec.WfClientEndpointName,
								strconv.Itoa(tac.Spec.WfClientSpec.WfClientEndpointPort),
								tac.Spec.WfTemporalSpec.WfClientName,
							).Set(1)
						}

					}
				}
			}

			time.Sleep(time.Duration(15 * time.Second))
		}
	}()
}
