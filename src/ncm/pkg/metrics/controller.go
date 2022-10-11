package metrics

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	netintents "gitlab.com/project-emco/core/emco-base/src/ncm/pkg/networkintents"
)

func start() {
	go func() {
		clusterClient := cluster.NewClusterClient()
		netClient := netintents.NewNetworkClient()
		providerNetClient := netintents.NewProviderNetClient()
		for {
			clps, err := clusterClient.GetClusterProviders(context.Background())
			if err != nil {
				fmt.Println(err)
				continue
			}

			for _, clp := range clps {
				clusters, err := clusterClient.GetClusters(context.Background(), clp.Metadata.Name)
				if err != nil {
					fmt.Println(err)
					continue
				}
				for _, cl := range clusters {
					networks, err := netClient.GetNetworks(clp.Metadata.Name, cl.Metadata.Name)
					if err != nil {
						fmt.Println(err)
						continue
					}
					for _, network := range networks {
						NetworkGauge.WithLabelValues(clp.Metadata.Name, cl.Metadata.Name, network.Metadata.Name, network.Spec.CniType).Set(1)
					}

					providerNets, err := providerNetClient.GetProviderNets(clp.Metadata.Name, cl.Metadata.Name)
					if err != nil {
						fmt.Println(err)
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
