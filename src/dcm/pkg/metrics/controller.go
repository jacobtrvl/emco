package metrics

import (
	"context"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/dcm/pkg/module"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	orchModule "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

func start() {
	go func() {
		client := module.NewClient()
		fields := log.Fields{"service": "dcm"}
		for {
			projects, err := orchModule.NewProjectClient().GetAllProjects(context.Background())
			if err != nil {
				log.Error(err.Error(), fields)
				continue
			}

			for _, proj := range projects {
				fields := fields
				fields["project"] = proj.MetaData.Name

				lcs, err := client.LogicalCloud.GetAll(context.Background(), proj.MetaData.Name)
				if err != nil {
					log.Error(err.Error(), fields)
					continue
				}
				for _, lc := range lcs {
					fields := fields
					fields["logical_cloud"] = lc.MetaData.Name
					st, err := client.LogicalCloud.GetState(context.Background(), proj.MetaData.Name, lc.MetaData.Name)
					if err != nil {
						log.Error(err.Error(), fields)
					}
					status := ""
					if len(st.Actions) > 0 {
						status = st.Actions[len(st.Actions)-1].State
					}
					LCGauge.WithLabelValues(proj.MetaData.Name, lc.MetaData.Name, lc.Specification.NameSpace, status).Set(1)
				}

			}

			time.Sleep(time.Duration(15 * time.Second))
		}
	}()
}
