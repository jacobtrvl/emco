package metrics

import (
	"fmt"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/dcm/pkg/module"
	orchModule "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

func start() {
	go func() {
		client := module.NewClient()
		for {
			projects, err := orchModule.NewProjectClient().GetAllProjects()
			if err != nil {
				fmt.Println(err)
				continue
			}

			for _, proj := range projects {
				lcs, err := client.LogicalCloud.GetAll(proj.MetaData.Name)
				if err != nil {
					fmt.Println(err)
					continue
				}
				for _, lc := range lcs {
					st, err := client.LogicalCloud.GetState(proj.MetaData.Name, lc.MetaData.Name)
					if err != nil {
						fmt.Println("statusError", err)
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
