# VMware Carbon Black Cloud Container Operator

## Custom Resources Definitions

The operator implements controllers for the Carbon Black Container custom resources definitions:

### 1. Carbon Black Container Cluster CR
<u>cbcontainersclusters.operator.containers.carbonblack.io</u>

This is the first CR you'll need to deploy in order to initialize the data plane components. This will create the data
plane Configmap, Registry Secret and PriorityClass.

### Required parameters

| Parameter                                    | Description                                                                     | Default |
| ---------------------------------------------| --------------------------------------------------------------------------------| ------- |
| `spec.account`                               | Carbon Black Container org key                                                  | ------- |
| `spec.clusterName`                           | Carbon Black Container cluster name  (<cluster_group:cluster_name>)             | ------- |
| `spec.apiGatewaySpec.host`                   | Carbon Black Container api host                                                 | ------- |
| `spec.eventsGatewaySpec.host`                | Carbon Black Container events host                                              | ------- |

### Optional parameters

| Parameter                                    | Description                                              | Default                     |
| ---------------------------------------------| ---------------------------------------------------------| ------------                |
| `spec.apiGatewaySpec.scheme`                 | Carbon Black Container api scheme                        | `https`                     |
| `spec.apiGatewaySpec.port`                   | Carbon Black Container api port                          | 443                         |
| `spec.apiGatewaySpec.adapter`                | Carbon Black Container api adapter                       | `containers`                |
| `spec.apiGatewaySpec.accessTokenSecretName`  | Carbon Black Container api access token secret name      | `cbcontainers-access-token` |
| `spec.eventsGatewaySpec.port`                | Carbon Black Container events port                       | 443                         |

### 2. Carbon Black Container Hardening CR
<u>cbcontainershardenings.operator.containers.carbonblack.io</u>

### Required parameters

| Parameter                                    | Description                                                      | Default |
| ---------------------------------------------| ---------------------------------------------------------------- | ------- |
| `spec.version`                               | Version of the Carbon Black Cloud Container Hardening components | ------- |
| `spec.eventsGatewaySpec.host`                | Carbon Black Container events host                               | ------- |

### Optional parameters

| Parameter                                    | Description                                                                | Default                                    |
| ---------------------------------------------| -------------------------------------------------------------------------- | ------------------------------------------ |                             
| `spec.enforcerSpec.replicasCount`            | Carbon Black Container Hardening Enforcer number of replicas               | 1                                          |
| `spec.accessTokenSecretName`                 | Carbon Black Container api access token secret name                        | `cbcontainers-access-token`                |
| `spec.enforcerSpec.image.repository`         | Carbon Black Container Hardening Enforcer image repository                 | `cbartifactory/guardrails-enforcer`        |
| `spec.stateReporterSpec.image.repository`    | Carbon Black Container Hardening State Reporter image repository           | `cbartifactory/guardrails-state-reporter`  |

### Common Optional parameters under both enforcerSpec & stateReporterSpec

| Parameter                                    | Description                                                                            | Default                                                       |                                                               
| ---------------------------------------------| ---------------------------------------------------------------------------------------| ------------------------------------------------------------- |                                                               
| `labels`                                     | Carbon Black Container Hardening Enforcer/State Reporter Deployment & Pod labels       | Empty map                                                     |
| `deploymentAnnotations`                      | Carbon Black Container Hardening Enforcer/State Reporter Deployment annotations        | Empty map                                                     |
| `podTemplateAnnotations`                     | Carbon Black Container Hardening Enforcer/State Reporter Pod annotations               | `{prometheus.io/scrape: "false", prometheus.io/port: "7071"}` |
| `env`                                        | Carbon Black Container Hardening Enforcer/State Reporter Pod environment vars          | Empty map                                                     |
| `image.tag`                                  | Carbon Black Container Hardening Enforcer/State Reporter image tag                     | The Hardening feature version                                 |
| `image.pullPolicy`                           | Carbon Black Container Hardening Enforcer/State Reporter pull policy                   | `Always`                                                      |
| `resources.requests`                         | Carbon Black Container Hardening Enforcer/State Reporter resources requests            | `{memory: "64Mi", cpu: "30m"}`                                |
| `resources.limits`                           | Carbon Black Container Hardening Enforcer/State Reporter resources limits              | `{memory: "256Mi", cpu: "200m"}`                              |
| `probes.livenessPath`                        | Carbon Black Container Hardening Enforcer/State Reporter probes liveness path          | `/alive`                                                      |
| `probes.port`                                | Carbon Black Container Hardening Enforcer/State Reporter probes port                   | 8181                                                          |
| `probes.scheme`                              | Carbon Black Container Hardening Enforcer/State Reporter probes scheme                 | `HTTP`                                                        |
| `probes.initialDelaySeconds`                 | Carbon Black Container Hardening Enforcer/State Reporter probes initial delay seconds  | 3                                                             |
| `probes.timeoutSeconds`                      | Carbon Black Container Hardening Enforcer/State Reporter probes timeout seconds        | 1                                                             |
| `probes.periodSeconds`                       | Carbon Black Container Hardening Enforcer/State Reporter probes period seconds         | 30                                                            |
| `probes.successThreshold`                    | Carbon Black Container Hardening Enforcer/State Reporter probes success threshold      | 1                                                             |
| `probes.failureThreshold`                    | Carbon Black Container Hardening Enforcer/State Reporter probes failure threshold      | 3                                                             |