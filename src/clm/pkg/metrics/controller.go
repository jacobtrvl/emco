package metrics

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
)

func start() {
	go func() {
		client := cluster.NewClusterClient()
		for {
			clps, err := client.GetClusterProviders(context.Background())
			if err != nil {
				fmt.Println(err)
				continue
			}

			for _, clp := range clps {
				CLPGauge.WithLabelValues(clp.Metadata.Name).Set(1)
				clusters, err := client.GetClusters(context.Background(), clp.Metadata.Name)
				if err != nil {
					fmt.Println(err)
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
