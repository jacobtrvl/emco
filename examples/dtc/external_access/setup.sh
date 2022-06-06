#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2021 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

HOST_IP=${HOST_IP:-"oops"}
KUBE_PATH1=${KUBE_PATH1:-"oops"}
KUBE_PATH2=${KUBE_PATH2:-"oops"}
# tar files
function create {
    # make the GMS helm charts and profiles
    mkdir -p output
    tar -czf output/httpbin.tgz -C ../../helm_charts/httpbin-server/helm httpbin
    tar -czf output/sleep.tgz -C ../../helm_charts/httpbin-client/helm sleep
    tar -czf output/httpbin-server-profile.tar.gz -C ../../helm_charts/httpbin-server/profile .
    tar -czf output/sleep-client-profile.tar.gz -C ../../helm_charts/httpbin-client/profile .


        cat << NET > values.yaml
    ClusterProvider: httpbinprovider1
    Cluster1: httpbincluster1
    Cluster2: sleepcluster1
    KubeConfig1: $KUBE_PATH2
    KubeConfig2: $KUBE_PATH1
    ProjectName: httpbinproj1
    LogicalCloud1RefName: httpbinlc1
    LogicalCloud2RefName: sleeplc1
    Cluster1Label: edge-cluster
    Cluster2Label: edge-cluster1
    Cluster1IstioIngressGatewayKvName: httpbinistioingresskvpairs1
    Cluster2IstioIngressGatewayKvName: httpbinistioingresskvpairs2
    AdminCloud: default
    LogicalCloud: std-lc1
    LogicalCloudNamespace: httpbin-ns
    IstioNamespace: istio-system
    LogicalCloudPermission: standard-permission
    IstioPermission: istio-permission
    CompositeApp: httpbin-collection-composite-app
    CompositeAppVersion: v1
    Applist:
      - httpbin
      - sleep
    AppsInCluster1:
      - httpbin
    AppsInCluster2:
      - sleep
    CompositeProfile: httpbin-collection-composite-profile
    DeploymentIntentGroup: httpbin-collection-deployment-intent-group
    DeploymentIntent: httpbin-collection-deployment-intent
    GenericPlacementIntent: httpbin-collection-placement-intent
    GenericPlacementIntent2: sleep-collection-placement-intent
    DtcIntent: httpbintestdtc
    DtcHttpbinServerIntent: httpbinserver
    DtcHttpbinServerIntentHTTP: httpbinserverhttp
    Intent: httpbinintent
    DtcClientsIntent: sleepclient1
    DtcClientsAccessIntent: allowstatus
    ServiceCertificate: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMzakNDQWNZQ0FRQXdEUVlKS29aSWh2Y05BUUVMQlFBd0xURVZNQk1HQTFVRUNnd01aWGhoYlhCc1pTQkoKYm1NdU1SUXdFZ1lEVlFRRERBdGxlR0Z0Y0d4bExtTnZiVEFlRncweU1qQXlNakl4TmpNMk5ERmFGdzB5TXpBeQpNakl4TmpNMk5ERmFNRDB4SERBYUJnTlZCQU1NRTJoMGRIQmlhVzR1WlhoaGJYQnNaUzVqYjIweEhUQWJCZ05WCkJBb01GR2gwZEhCaWFXNGdiM0puWVc1cGVtRjBhVzl1TUlJQklqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FROEEKTUlJQkNnS0NBUUVBdVBtTDBuek1udXQ3ZjhVZUNXclFCa2NvVnJVYWNtMS9GZ01PMGlPeUFISU0zWWROZ014TApPWktKaW9iK0VqbEZiMlVxcC9MSnhxTlJtakV6WUpmczZrZHZLMStMODQvY2lrTzBNY0d1S0tReDVReTF6ZkJhCktWaGRFNEJqd3JkR0J0bVNkV1E2THd6Q1hJYXlnc0c1dWx2NUx3Y1NtOGI3ditHUmEwRXJranllY3k0a1BxYTQKZ0FFY1h1RldWWXFJNlU2MUh0ZVVyeG1pNHMwdnlIMlpaeUN0Wk1zVHpvOXYxeGlhdHZRN3ZPYThXT2FQdFpmOApWdjJnVnJhWlFIdFJ2ZzhDZllLTE5YM3Y5aldHdDRYUzE3d1NRdWYybkExSUN5VzdUczM0ZmMrMS82ZUtsU1hLCnpsVUV4Q2QxYkJqYUZXWUwxYlZvY1FRS21oSHozMnJmNlFJREFRQUJNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUIKQVFBUm53bFRndHVCbzB0c2xibXpkZmtLelRVMnpLVi90eTUrN2JQQk02M2VPNUxyMTNEblpXeTA0dkh0OTN4dwpiRjZ5VW5QWDhId3dwbUxMOHMwM2ExYXlPQzh6SWwvSS96bVQyVmVpbjJnTGFrenFHNlBxakRYb3Q3T3FNYkVDCkwyNUdER3FsVTJoWmZQV004VTRpSTZ0ZUVnQUVPWnF0NXNYZGpKM1JNUFgxVll4UHlIaFczcXFtb0tZZmxZNWkKOFlrejd1a0U1THdoY1VuazNRTDVXS3NhNGpPUWswalhZS242SDdYdnNNV09qb1BuZ1ZFWERGZHB2N0JCU0tJagpZcTVqaDErMFZGc1FQREtDYXkzSFdva3gvYTBwZEdpN2tta3hBcGFGLy9BaE0zYU8zdDNhaXdwV1VnL2JkOXJWCkNUM1lKeHBEdlFRRklCMmlKcmhMdklxcAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
    ServicePrivateKey: "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2UUlCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktjd2dnU2pBZ0VBQW9JQkFRQzQrWXZTZk15ZTYzdC8KeFI0SmF0QUdSeWhXdFJweWJYOFdBdzdTSTdJQWNnemRoMDJBekVzNWtvbUtodjRTT1VWdlpTcW44c25HbzFHYQpNVE5nbCt6cVIyOHJYNHZ6ajl5S1E3UXh3YTRvcERIbERMWE44Rm9wV0YwVGdHUEN0MFlHMlpKMVpEb3ZETUpjCmhyS0N3Ym02Vy9rdkJ4S2J4dnUvNFpGclFTdVNQSjV6TGlRK3ByaUFBUnhlNFZaVmlvanBUclVlMTVTdkdhTGkKelMvSWZabG5JSzFreXhQT2oyL1hHSnEyOUR1ODVyeFk1bysxbC94Vy9hQld0cGxBZTFHK0R3Sjlnb3MxZmUvMgpOWWEzaGRMWHZCSkM1L2FjRFVnTEpidE96Zmg5ejdYL3A0cVZKY3JPVlFURUozVnNHTm9WWmd2VnRXaHhCQXFhCkVmUGZhdC9wQWdNQkFBRUNnZ0VBUmtTR0dTL1BpNDlwR3VDR3lJMEsrVmVPdTJHUTZtY3VIKzZKY3NxY2xBNi8KVkdoUnlOdlN0OHd5ODZ6VVY1ZnFDS2NselNjdC80ZUxPRWY0ZkhrNlJzVmNOZDNXREhCYUZ5d2hCOFhMb3lTOAp6NFpFaWpjRUNURElLdUJiQlYwWi9RQXA0dTV3Sys5czVqbEZGdWNBNXdxSlhwUVJQWndaaG9ycDh4U091TDRwCkhaeG9yQjJmMUR5TWw2WE42cE03ZjBUcFFYejVBcXB3YkVvTmxubGROVTlVa2pkNjdHYTNnTmYvKzliY0VyY00KeXB4K2JhRWFmS05pSTBmdFhVck80NkFIdlF0MC9WNEtmQURPejBzd3duMEVybWVnazhxbHdMcGF6dFRNMTMvQQpOZlFFM2FyMUhaVWoxa0pERldGcjdnQ0dqS2xlNmRKU1I4em5yeDdwQVFLQmdRRGF3eFJ0ZEttTDc3ZytqdkF2CmpGZTBodkl3bHZkZUpYMGh6ZTNPM2M1MVo5UUtqT1FQY0F2emk0VTNWUVIwcGVMcUFZZitTa3NQYnJWdVkzYjYKMFhYbHV3UnQ1Y2ExQkhFMWgwdFpFclY1VkVRaktFTGtjOGVnWm9HT3RDVTB0K3RUU1ZNVjZwNEVMc3pacFdzSgp4KzNDOTFTeUdxZ2xTaFlQRzAzZlo5YnU1d0tCZ1FEWWRpR0w3TTJyRUZKbWZTbGNsTjR6bmo1OGtWcXMvN1JVCkdPREIreENxYldiYndMNjg3UG1TNi8zOTFONXIyRXlMNkdKc2RpbjJlUlhjL0svN2dEWlRMMDg4RzlCaG5oa20Kd0djYkRTOGRmR3F5eFV0cVk3YnAvbmgwNW9KUGZnZXQxMXNYQ00zL3lqUHExcHlTT2RiWEg1Vys4UzVna2JGNgpQVXNvNkxQd3J3S0JnUUNwSi9kL3U1bnVydXFVMVFvOGVoVEhieUdQR08wbVMyNjYyUFZ0NUcxa3MyaHUwQXI1Ck5QYkkxN1dtMTRLZWdEYzZJdno1VUpGQjJhVkpPbmdoOGgxc0NuU2VWZktVdmw2YnVZWTExaFdsUDllQUovMngKa1NWbmpsdlg4TXhrTzJNbi82YlRaNXZRT0RBR2k3WjgxYSt2OW5melVGRjhwQkR5bFhaZHJYbXhPUUtCZ0hITwp3bHFJT2FZOElhYkRIYkVRa0RkQmR3Y1ZnVUE5L1BqT1Q3V05wRGlHNXJLWmgyOWJoT2lMYlhJOHJtaXpRNk8zCm5hLzYvSnNiRkxTb29ub3Y4ZUFRbXE4MnpIdlduTkMwRGtHNXo5REg3bTFwci9vU3pVUC95Q05tWXBNYTV6eXUKWXJVY3F2cFd2SzgzQVFFY3FlbFhNT3RBY1NyU3p5WSsvYnBYaHV2L0FvR0FRUnJuQ2JaMEZQTXdjMkptMk90cgpmeWxyb0RZRlFmNzhvRTBzWHFGeFVmOXFQSzJUd0RjamRQSnVScFAyTk14YnRlZzY0enkxWEc4YURaVXUrSGpaCnEyMWI2SkVmd2I5RTlBeHhtanRsd1BLV05uWnRMUUJ5RUErQXEvQ3JqOVpBT3ViQ0VMMXBHOTgrakgxZXhrZkEKeittYkdEQnNEczB4ZTNHUEEwbnYwZUE9Ci0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0K"
    CaCertificate: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURPekNDQWlPZ0F3SUJBZ0lVQlFPMTF4TUxnakdIdkxmYXRJamYxcjd2L3Nzd0RRWUpLb1pJaHZjTkFRRUwKQlFBd0xURVZNQk1HQTFVRUNnd01aWGhoYlhCc1pTQkpibU11TVJRd0VnWURWUVFEREF0bGVHRnRjR3hsTG1OdgpiVEFlRncweU1qQXlNakl4TmpNMk5ERmFGdzB5TXpBeU1qSXhOak0yTkRGYU1DMHhGVEFUQmdOVkJBb01ER1Y0CllXMXdiR1VnU1c1akxqRVVNQklHQTFVRUF3d0xaWGhoYlhCc1pTNWpiMjB3Z2dFaU1BMEdDU3FHU0liM0RRRUIKQVFVQUE0SUJEd0F3Z2dFS0FvSUJBUUN6S3BSckdDY0JBaE9WamxaVlpQSlRYMUVhTUs3UUdxbEpxT04vL0tHdQplRVFYRjlpc1ordTVwQ0xVaTVERmFMUDVZZ1JUNURJaXZkNTdhSEp5NndOUk8wck0rMXNBRWQ3ajAvN0VQcFQ3CkJ4ZFg0UmM0WWtERk1jWXBobmZuNEpoL0M4clNybE5zUzNscnZHOFRSQWkxTkNGYTloQkQyUUcvN2RTTlFUWWkKbFFMUU41UW5lUU82Y3l3RnMvK1lLUWtZdGpGcXJ0S3pzVGFvZGt1ZzIxWXNwWHljaDl4YnFEcHVlZm9EVkdzSQpFZndSeWI1V1dPZHJEVkJ1dXR1SGFJcnN2OGJSeUJWOXdwczhPdEFNSTZXM3hhbzgrbFBwTWt5YjVhaHcxOVN2ClIwRlRjUEpFYkFkMFRvTEd0TVRQMTI4M3NpaUZpK3FGMklsZU5xbE5SL04xQWdNQkFBR2pVekJSTUIwR0ExVWQKRGdRV0JCUlNxREtIZHVzTXZyWVZ1MXQwU1kxSjdlRjdNVEFmQmdOVkhTTUVHREFXZ0JSU3FES0hkdXNNdnJZVgp1MXQwU1kxSjdlRjdNVEFQQmdOVkhSTUJBZjhFQlRBREFRSC9NQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUJBUUE1ClZqaDh3d3gwdVZWTm5FVE1hVUtUdnE2WWhNbEg3ejJ6YjhCaGwxT3pkS2NhYURPVXRCRnRGWDd1UDZUVGdVd2sKWjQzSFNoYzcxU0JwcjJQdHZuTXZTQ1BtdG9weVBJdFFoMlRqdllUOW1ySTQ0UTJuRGZ4SWwwejNNdmppM3RiVgplVzJVclRPTmNXRDJxSnZhbzlZaGxkRGdjcmlOM0JBcG0rZjB3MElDdXlNNFYzemVEdTFoRVMrM1I3K1ZycDc4CmNmbW80RC9rMkpLdXM2K1E5WjYxMUs3aEFXQWphQ3F3ckhFUUN0L0s5T0JoSkJ3SGxhM0V5YWhLOXFMYmxhdUUKMnI3d21yZTZzcGRoSXAxL2FhcDRZNmxXRkxKdEs3d3R1ZEtYSGRDZHNZT3kyTDM4V2lhdmFDNU80eWo2UnRlZApNZHpVTlpwdFoxSUxQNERSNVVVWAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
    RsyncPort: 30431
    DtcPort: 30448
    ItsPort: 30440
    HostIP: $HOST_IP
NET
cat << NET > emco-cfg.yaml
  orchestrator:
    host: $HOST_IP
    port: 30415
  clm:
    host: $HOST_IP
    port: 30461
  ncm:
    host: $HOST_IP
    port: 30481
  ovnaction:
    host: $HOST_IP
    port: 30451
  dcm:
    host: $HOST_IP
    port: 30477
  gac:
    host: $HOST_IP
    port: 30420
  dtc:
   host: $HOST_IP
   port: 30418
  rsync:
   host: $HOST_IP
   port: 30431
NET

}

function usage {
    echo "Usage: $0  create|cleanup"
}

function cleanup {
    rm -f *.tar.gz
    rm -f values.yaml
    rm -f emco-cfg.yaml
    rm -rf output
}

if [ "$#" -lt 1 ] ; then
    usage
    exit
fi

case "$1" in
    "create" )
        if [ "${HOST_IP}" == "oops" ] || [ "${KUBE_PATH1}" == "oops" ] || [ "${KUBE_PATH2}" == "oops" ]; then
            echo -e "ERROR - HOST_IP, KUBE_PATH1 & KUBE_PATH2 environment variable needs to be set"
        else
            create
        fi
        ;;
    "cleanup" )
        cleanup
    ;;
    *)
        usage ;;
esac
