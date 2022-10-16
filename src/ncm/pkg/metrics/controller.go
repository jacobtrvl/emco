package metrics

import (
	"context"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	netintents "gitlab.com/project-emco/core/emco-base/src/ncm/pkg/networkintents"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

func start() {
	go func() {
		clusterClient := cluster.NewClusterClient()
		netClient := netintents.NewNetworkClient()
		providerNetClient := netintents.NewProviderNetClient()
		fields := log.Fields{"service": "ncm"}
		for {
			clps, err := clusterClient.GetClusterProviders(context.Background())
			if err != nil {
				log.Error(err.Error(), fields)
				continue
			}

			for _, clp := range clps {
				fields := fields
				fields["cluster_provider"] = clp.Metadata.Name
				clusters, err := clusterClient.GetClusters(context.Background(), clp.Metadata.Name)
				if err != nil {
					log.Error(err.Error(), fields)
					continue
				}
				for _, cl := range clusters {
					fields := fields
					fields["cluster"] = cl.Metadata.Name
					networks, err := netClient.GetNetworks(clp.Metadata.Name, cl.Metadata.Name)
					if err != nil {
						log.Error(err.Error(), fields)
						continue
					}
					for _, network := range networks {
						NetworkGauge.WithLabelValues(clp.Metadata.Name, cl.Metadata.Name, network.Metadata.Name, network.Spec.CniType).Set(1)
					}

					providerNets, err := providerNetClient.GetProviderNets(clp.Metadata.Name, cl.Metadata.Name)
					if err != nil {
						log.Error(err.Error(), fields)
						continue
					}
					for _, network := range providerNets {
						ProviderNetworkGauge.WithLabelValues(
							clp.Metadata.Name,
							cl.Metadata.Name,
							network.Metadata.Name,
							network.Spec.CniType,
							network.Spec.ProviderNetType,
							network.Spec.Vlan.VlanId,
							network.Spec.Vlan.ProviderInterfaceName,
							network.Spec.Vlan.LogicalInterfaceName,
							network.Spec.Vlan.VlanNodeSelector,
						).Set(1)
					}
				}
			}

			time.Sleep(time.Duration(15 * time.Second))
		}
	}()
}
