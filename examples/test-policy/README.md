```text
  Copyright (c) 2022 Aarna Networks
```
## Policy Controller Use case
This is a step-by-step guide for building a use-case for the policy controller. 
This use case explains how to set up a policy intent based on container memory usage and instantiate migrate-workflow based on the policy.
### Pre-requisite setup
1. Setup Temporal Server
   ```shell
   $ git clone https://github.com/temporalio/helm-charts.git  
   $ cd helm-charts  
   $ helm install --set server.replicaCount=1  --set cassandra.config.cluster_size=1 --set prometheus.enabled=false --set grafana.enabled=false --set elasticsearch.enabled=false temporaltest . --timeout 15m
   ``` 
2. Migrate Workflow.
   The steps for this are expained in https://gitlab.com/project-emco/ecosystem/temporal-migrate-workflow
   - In https://gitlab.com/project-emco/ecosystem/temporal-migrate-workflow/-/tree/main/samples/intents
     Run first 5 Steps only:
  ```text
   $ git clone https://gitlab.com/project-emco/ecosystem/temporal-migrate-workflow
   $ cd samples
   $ emcoctl --config intents/emco-cfg-remote.yaml apply -v intents/values*.yaml -f intents/00.define-clusters-proj.yaml
   $ emcoctl --config intents/emco-cfg-remote.yaml apply -v intents/values*.yaml -f intents/01.instantiate-lc.yaml
   $ emcoctl --config intents/emco-cfg-remote.yaml apply -v intents/values*.yaml -f intents/02.define-app-dig.yaml
   $ emcoctl --config intents/emco-cfg-remote.yaml apply -v intents/values*.yaml -f intents/03.instantiate-dig.yaml
   $ emcoctl --config intents/emco-cfg-remote.yaml apply -v intents/values*.yaml -f intents/04.define-workflow-1.yaml`
   ```
   Note: You might need to update IPs in samples/intents/emco-cfg-remote.yaml according to your system.  

4. Setup OPA Policy
   We can create our policy on OPA play ground and provide those bundles for OPA
   You can use this policy:
   https://play.openpolicyagent.org/p/KUPp8IuXVq  
   
   Paste following policy in playground
   ```text 
    package emco

   import future.keywords.in

   default actionRequired := false

   actionRequired {
      input.agentSpec.agentId == "EMCOTEST"
      input.event.metric == "memory_usage_bytes"
      some item in input.event.value.items
      to_number(item.value) > 999999999
   }
   ```
   - Change to_number(item.value) > 999999999 to the desired memory_usage_bytes threshold.
   This depends on the system.  
   
   - Press publish on right corner in OPA play  
   This will  generate a bundle. Replace bundle  in below command 
   ```shell
    $ docker run -p 8181:8181 openpolicyagent/opa run --server --log-format text --set decision_logs.console=true --set bundles.play.polling.long_polling_timeout_seconds=45 --set services.play.url=https://play.openpolicyagent.org --set bundles.play.resource=bundles/2UHfqTDOTl
    ```
   
5. Setup Prometheus Adapter on the edge clusters. Prometheus adapter will expose metrics on custom.metrics.io.

   _**Assuming Prometheus and cadvisor are already available on edge cluster. Steps for them are not provided in this document**_ 
   
   ```shell
   Note: These steps are for edge clusters. [See below for modifying promethues url]
   $ helm repo add prometheus-community https://prometheus-community.github.io/helm-charts  
   $ helm repo update  
   $ helm install my-release prometheus-community/prometheus-adapter
   ```
   You may need to change prometheus url configuration if its not default
   - https://github.com/prometheus-community/helm-charts/blob/main/charts/prometheus-adapter/values.yaml
   - See the session prometheus for details
6. Setup EdgeMetricsCollector on the edge clusters
   Helm charts for will be build as part of emco build. Copy metricscollector chart from  [EMCO_DIR]/bin/helm/ 
   ```shell
   Note: These steps are for edge clusters
   $ helm install metricscollector-helm-ubuntu-latest.tgz
   ```
7. Create a custom resource for enabling metrics collection
   Inorder to metricscollector to start watching the metrics, we need to create a resource of kind MetricsCollector.  

   metricscollector will watch the metrics listed in this CR. 
   ```shell
   Note: Run on edge clusters
   $ kubectl apply -f metricscollector-cr.yaml
   ```
8. Register metricscollector to policy controller
   Edit agentremote.json to update your metricscollector endpoint
   ```shell
   $ curl <policy controller-ip>:<port>/v2/policy/agents/<agent_name> @agentremote.json
   ```
   Example:
   curl -X POST 10.107.144.35:9060/v2/policy/agents/id1 -d @agentremote.json
9. Setup policy intent
   In the emco cluster, go to temporal-migrate-workflow clone we did in step (2)
   Edit samples/intents/emco-cfg-remote.yaml to add policy controller endpoint
   Add: 
   ```text
   policy:
    host: <host IP>
    port: 30460
   ```
   Copy file [06.define_policy_intent.yaml](06.define_policy_intent.yaml) to samples/intents. Update IP addresses (OPA, Worflow Manager) int policy intent to reflect your setup.
   ```text
    $ emcoctl --config intents/emco-cfg-remote.yaml apply -v intents/values*.yaml -f 06.define_policy_intent.yaml
   ```
10. You can simulate closed loop by changing the policy to a lower value than current measurement of memory_usage_bytes
    or simulating the memory load by running 'stress' command in the container shell.
   ```shell
    $ stress -c 2 -i 1 -m 1 --vm-bytes 128M -t 10s
   ```