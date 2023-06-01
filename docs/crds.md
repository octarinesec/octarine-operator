# VMware Carbon Black Cloud Container Operator

## Custom Resources Definitions

The operator implements controllers for the Carbon Black Container custom resources definitions:

### 1. Carbon Black Container Agent CR
<u>cbcontainersagents.operator.containers.carbonblack.io</u>

This is the CR you'll need to deploy in order to trigger the operator to deploy the data plane components.

### Required parameters

| Parameter                                   | Description                                                                                        |
|---------------------------------------------|----------------------------------------------------------------------------------------------------|
| `spec.account`                              | Carbon Black Container org key                                                                     |
| `spec.clusterName`                          | Carbon Black Container cluster name  (<cluster_group:cluster_name>)                                |
| `spec.version`                              | Carbon Black Container agent version                                                               |
| `spec.gateways.apiGateway.host`             | Carbon Black Container api host                                                                    |
| `spec.gateways.coreEventsGateway.host`      | Carbon Black Container core events host (e.g Health Checks)                                        |
| `spec.gateways.hardeningEventsGateway.host` | Carbon Black Container hardening events host (e.g applied, deleted, validated & blocked resources) |
| `spec.gateways.runtimeEventsGateway.host`   | Carbon Black Container runtime events host (e.g traffic events)                                    |

### Optional parameters

| Parameter                                   | Description                                         | Default                     |
|---------------------------------------------|-----------------------------------------------------|-----------------------------|
| `spec.apiGateway.port`                      | Carbon Black Container api port                     | 443                         |
| `spec.accessTokenSecretName`                | Carbon Black Container api access token secret name | `cbcontainers-access-token` |
| `spec.gateways.coreEventsGateway.port`      | Carbon Black Container core events port             | 443                         |
| `spec.gateways.hardeningEventsGateway.port` | Carbon Black Container hardening events port        | 443                         |
| `spec.gateways.runtimeEventsGateway.port`   | Carbon Black Container runtime events port          | 443                         |

### Basic Components Optional parameters

| Parameter                                              | Description                                                                | Default                                                                            |
|--------------------------------------------------------|----------------------------------------------------------------------------|------------------------------------------------------------------------------------|
| `spec.components.basic.enforcer.replicasCount`         | Carbon Black Container Hardening Enforcer number of replicas               | 1                                                                                  |
| `spec.components.basic.monitor.image.repository`       | Carbon Black Container Monitor image repository                            | `cbartifactory/monitor`                                                            |
| `spec.components.basic.enforcer.image.repository`      | Carbon Black Container Hardening Enforcer image repository                 | `cbartifactory/guardrails-enforcer`                                                |
| `spec.components.basic.stateReporter.image.repository` | Carbon Black Container Hardening State Reporter image repository           | `cbartifactory/guardrails-state-reporter`                                          |
| `spec.components.basic.monitor.resources`              | Carbon Black Container Monitor resources                                   | `{requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "256Mi", cpu: "200m"}}` |
| `spec.components.basic.enforcer.resources`             | Carbon Black Container Hardening Enforcer resources                        | `{requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "256Mi", cpu: "200m"}}` |
| `spec.components.basic.stateReporter.resources`        | Carbon Black Container Hardening State Reporter resources                  | `{requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "256Mi", cpu: "200m"}}` |

### Runtime Components Optional parameters

| Parameter                                                     | Description                                                                  | Default                                                                              |
|---------------------------------------------------------------|------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| `spec.components.runtimeProtection.enabled`                   | Carbon Black Container flag to control Runtime components deployment         | true                                                                                 |
| `spec.components.runtimeProtection.resolver.image.repository` | Carbon Black Container Runtime Resolver image repository                     | `cbartifactory/runtime-kubernetes-resolver`                                          |
| `spec.components.runtimeProtection.sensor.image.repository`   | Carbon Black Container Runtime Sensor image repository                       | `cbartifactory/runtime-kubernetes-sensor`                                            |
| `spec.components.runtimeProtection.internalGrpcPort`          | Carbon Black Container Runtime gRPC port the resolver exposes for the sensor | 443                                                                                  |
| `spec.components.runtimeProtection.resolver.logLevel`         | Carbon Black Container Runtime Resolver log level                            | "panic", "fatal", "error", "warn", "info", "debug", "trace"  (default info)          |
| `spec.components.runtimeProtection.resolver.resources`        | Carbon Black Container Runtime Resolver resources                            | `{requests: {memory: "64Mi", cpu: "200m"}, limits: {memory: "1024Mi", cpu: "900m"}}` |
| `spec.components.runtimeProtection.sensor.logLevel`           | Carbon Black Container Runtime Sensor log level                              | "panic", "fatal", "error", "warn", "info", "debug", "trace"  (default info)          |
| `spec.components.runtimeProtection.sensor.resources`          | Carbon Black Container Runtime Sensor resources                              | `{requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "1024Mi", cpu: "500m"}}`  |

### Cluster Scanning Components Optional parameters

| Parameter                                                                            | Description                                                                                                                | Default                                                                              |
|--------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| `spec.components.clusterScanning.enabled`                                            | Carbon Black Container flag to control Cluster Scanning components deployment                                              | true                                                                                 |
| `spec.components.clusterScanning.imageScanningReporter.image.repository`             | Carbon Black Container Image Scanning Reporter image repository                                                            | `cbartifactory/image-scanning-reporter`                                              |
| `spec.components.clusterScanning.clusterScanner.image.repository`                    | Carbon Black Container Cluster Scanner Agent image repository                                                              | `cbartifactory/cluster-scanner`                                                      |
| `spec.components.clusterScanning.imageScanningReporter.resources`                    | Carbon Black Container Image Scanning Reporter resources                                                                   | `{requests: {memory: "64Mi", cpu: "200m"}, limits: {memory: "1024Mi", cpu: "900m"}}` |
| `spec.components.clusterScanning.clusterScanner.resources`                           | Carbon Black Container Cluster Scanner resources                                                                           | `{requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "1024Mi", cpu: "500m"}}`  |
| `spec.components.clusterScanning.clusterScanner.k8sContainerEngine.engineType`       | Carbon Black Container Cluster Scanner k8s container engine type. One of the options: `containerd`/`docker-daemon`/`cri-o` |                                                                                      |
| `spec.components.clusterScanning.clusterScanner.k8sContainerEngine.endpoint`         | Carbon Black Container Cluster Scanner k8s container engine endpoint path                                                  |                                                                                      |
| `spec.components.clusterScanning.clusterScanner.k8sContainerEngine.CRIO.storagePath` | Carbon Black Container Cluster Scanner override default image storage path (CRI-O only)                                    |                                                                                      |
| `spec.components.clusterScanning.clusterScanner.k8sContainerEngine.CRIO.configPath`  | Carbon Black Container Cluster Scanner override default image storage config path (CRI-O only)                             |                                                                                      |

### Components Common Optional parameters

| Parameter                                    | Description                                                                            | Default                                                       |
|----------------------------------------------|----------------------------------------------------------------------------------------|---------------------------------------------------------------|
| `labels`                                     | Carbon Black Container Component Deployment & Pod labels                               | Empty map                                                     |
| `deploymentAnnotations`                      | Carbon Black Container Component Deployment annotations                                | Empty map                                                     |
| `podTemplateAnnotations`                     | Carbon Black Container Component Pod annotations                                       | `{}`                                                          |
| `env`                                        | Carbon Black Container Component Pod environment vars                                  | Empty map                                                     |
| `image.tag`                                  | Carbon Black Container Component image tag                                             | The agent version                                             |
| `image.pullPolicy`                           | Carbon Black Container Component pull policy                                           | `IfNotPresent`                                                |
| `probes.port`                                | Carbon Black Container Component probes port                                           | 8181                                                          |
| `probes.scheme`                              | Carbon Black Container Component probes scheme                                         | `HTTP`                                                        |
| `probes.initialDelaySeconds`                 | Carbon Black Container Component probes initial delay seconds                          | 3                                                             |
| `probes.timeoutSeconds`                      | Carbon Black Container Component probes timeout seconds                                | 1                                                             |
| `probes.periodSeconds`                       | Carbon Black Container Component probes period seconds                                 | 30                                                            |
| `probes.successThreshold`                    | Carbon Black Container Component probes success threshold                              | 1                                                             |
| `probes.failureThreshold`                    | Carbon Black Container Component probes failure threshold                              | 3                                                             |
| `prometheus.enabled`                         | Carbon Black Container Component enable Prometheus scraping                            | false                                                         |
| `prometheus.port`                            | Carbon Black Container Component Prometheus server port                                | 7071                                                          |
| `nodeSelector`                               | Carbon Black Container Component node selector                                         | `{}`                                                          |
| `affinity`                                   | Carbon Black Container Component affinity                                              | `{}`                                                          |

### Other Components Optional parameters
| Parameter                                        | Description                                                                            | Default                                                       |
|--------------------------------------------------|----------------------------------------------------------------------------------------|---------------------------------------------------------------|
| `spec.components.settings.daemonSetsTolerations` | Carbon Black DaemonSet Component Tolerations                                           | Empty array                                                   |
