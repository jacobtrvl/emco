package metrics

import (
	"context"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

func start() {
	go func() {
		client := cluster.NewClusterClient()
		fields := log.Fields{"service": "clm"}
		for {
			clps, err := client.GetClusterProviders(context.Background())
			if err != nil {
				log.Error(err.Error(), fields)
				continue
			}

			for _, clp := range clps {
				fields := fields
				fields["cluster_provider"] = clp.Metadata.Name
				CLPGauge.WithLabelValues(clp.Metadata.Name).Set(1)
				clusters, err := client.GetClusters(context.Background(), clp.Metadata.Name)
				if err != nil {
					log.Error(err.Error(), fields)
					continue
				}
				for _, cl := range clusters {
					ClusterGauge.WithLabelValues(cl.Metadata.Name, clp.Metadata.Name).Set(1)
				}
			}

			time.Sleep(time.Duration(15 * time.Second))
		}
	}()
}
