# VMware Carbon Black Cloud Container Operator
## Overview 

The Carbon Black Cloud Container Operator runs within a Kubernetes cluster. The Container Operator is a set of controllers which deploy and manage the VMware Carbon Black Cloud Container components. 
 
 Capabilities
 * Deploy and manage the Container Essentials product bundle (including the configuration and the image scanning for Kubernetes security)!
 * Automatically fetch and deploy the Carbon Black Cloud Container private image registry secret
 * Automatically register the Carbon Black Cloud Container cluster
 * Manage the Container Essentials validating webhook - dynamically manage the admission control webhook to avoid possible downtime
 * Monitor and report agent availability to the Carbon Black console

The Carbon Black Cloud Container Operator utilizes the operator-framework to create a GO operator, which is responsible for managing and monitoring the Cloud Container components deployment. 

## Compatibility Matrix

| Operator version | Kubernetes Sensor Component Version | Minimum Kubernetes Version |
|------------------|-------------------------------------|----------------------------|
| v6.0.x           | 2.10.0, 2.11.0, 2.12.0, 3.0.0       | 1.18                       |
| v5.6.x           | 2.10.0, 2.11.0, 2.12.0              | 1.16                       |
| v5.5.x           | 2.10.0, 2.11.0                      | 1.16                       |

## Operator Deployment

### Prerequisites
Kubernetes 1.18+ is supported.

### From script:
```
export OPERATOR_VERSION=v6.0.0
export OPERATOR_SCRIPT_URL=https://setup.containers.carbonblack.io/$OPERATOR_VERSION/operator-apply.sh
curl -s $OPERATOR_SCRIPT_URL | bash
```

{OPERATOR_VERSION} is of the format "v{VERSION}"

Versions list: [Releases](https://github.com/octarinesec/octarine-operator/releases)

### From Source Code
Clone the git project and deploy the operator from the source code

By default, the operator utilizes CustomResourceDefinitions v1, which requires Kubernetes 1.16+.
Deploying an operator with CustomResourceDefinitions v1beta1 (deprecated in Kubernetes 1.16, removed in Kubernetes 1.22) can be done - see the relevant section below.

#### Create the operator image
```
make docker-build docker-push IMG={IMAGE_NAME}
```

#### Deploy the operator resources
```
make deploy IMG={IMAGE_NAME}
```

* View [Developer Guide](docs/developers.md) to see how deploy the operator without using an image

## Data Plane Deployment

### 1. Apply the Carbon Black Container Api Token Secret

```
kubectl create secret generic cbcontainers-access-token \
--namespace cbcontainers-dataplane --from-literal=accessToken=\
{API_Secret_Key}/{API_ID}
```

### 2. Apply the Carbon Black Container Agent Custom Resource

The operator implements controllers for the Carbon Black Container custom resources definitions

[Full Custom Resources Definitions Documentation](docs/crds.md)

#### 2.1 Apply the Carbon Black Container Agent CR

<u>cbcontainersagents.operator.containers.carbonblack.io</u>

This is the CR you'll need to deploy in order to trigger the operator to deploy the data plane components.

```sh
apiVersion: operator.containers.carbonblack.io/v1
kind: CBContainersAgent
metadata:
  name: cbcontainers-agent
spec:
  account: {ORG_KEY}
  clusterName: {CLUSTER_GROUP}:{CLUSTER_NAME}
  version: {AGENT_VERSION}
  gateways:
    apiGateway:
      host: {API_HOST}
    coreEventsGateway:
      host: {CORE_EVENTS_HOST}
    hardeningEventsGateway:
      host: {HARDENING_EVENTS_HOST}
    runtimeEventsGateway:
      host: {RUNTIME_EVENTS_HOST}
```

* notice that without applying the api token secret, the operator will return the error:
`couldn't find access token secret k8s object`

### Uninstalling the Carbon Black Cloud Container Operator

```sh
make undeploy
```

* Notice that the above command will delete the Carbon Black Container custom resources definitions and instances.

## Reading Metrics With Prometheus

The operator metrics are protected by kube-auth-proxy.

You will need to grant permissions to your Prometheus server to allow it to scrape the protected metrics.

You can create a ClusterRole and bind it with ClusterRoleBinding to the service account that your Prometheus server uses.

If you don't have such cluster role & cluster role binding configured, you can use the following:

Cluster Role:
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
name: cbcontainers-metrics-reader
rules:
- nonResourceURLs:
    - /metrics
      verbs:
      - get
```

Cluster Role binding creation:
```sh
kubectl create clusterrolebinding metrics --clusterrole=cbcontainers-metrics-reader --serviceaccount=<prometheus-namespace>:<prometheus-service-account-name>
```

## Changing components resources:
```yaml
spec:
  components:
    basic:
      monitor:
        resources:
          limits:
            cpu: 200m
            memory: 256Mi
          requests:
            cpu: 30m
            memory: 64Mi
      enforcer:
        resources:
          #### DESIRED RESOURCES SPEC - for hardening enforcer container
      stateReporter:
        resources:
          #### DESIRED RESOURCES SPEC - for hardening state reporter container
    runtimeProtection:
      resolver:
        resources:
          #### DESIRED RESOURCES SPEC - for runtime resolver container
      sensor:
        resources:
          #### DESIRED RESOURCES SPEC - for node-agent runtime container
    clusterScanning:
      imageScanningReporter:
        resources:
          #### DESIRED RESOURCES SPEC - for image scanning reporter pod
      clusterScanner:
        resources:
          #### DESIRED RESOURCES SPEC - for node-agent cluster-scanner container
```
#### Cluster Scanner Component Memory
The `clusterScanning.clusterScanner` component, tries by default to scan images with size up to 1GB.
To do so, its recommended resources are:
```yaml
resources:
  requests:
    cpu: 100m
    memory: 1Gi
  limits:
    cpu: 2000m
    memory: 4Gi
```

If your images are larger than 1GB, and you want to scan them, you'll need to allocate higher memory resources in the 
component's `requests.memory` & `limits.memory`, and add an environment variable `MAX_COMPRESSED_IMAGE_SIZE_MB`, to override
the max images size in MB, the scanner tries to scan.

For example, setting the cluster scanner to be able to scan images up to 1.5 GB configuration will be:
```yaml
spec:
  components:
    clusterScanning:
      clusterScanner:
        env:
          MAX_COMPRESSED_IMAGE_SIZE_MB: "1536" // 1536 MB == 1.5 GB
        resources:
          requests:
            cpu: 100m
            memory: 2Gi
          limits:
            cpu: 2000m
            memory: 5Gi
```

If your nodes have low memory, and you want the cluster scanner to consume less memory, you need to reduce the 
component's `requests.memory` & `limits.memory` , and override the `MAX_COMPRESSED_IMAGE_SIZE_MB`, to be less than 1GB (1024MB).

For example, assigning lower memory resources, and set the cluster-scanner to try and scan images up to 250MB:
```yaml
spec:
  components:
    clusterScanning:
      clusterScanner:
        env:
          MAX_COMPRESSED_IMAGE_SIZE_MB: "250" // 250 MB
        resources:
          requests:
            cpu: 100m
            memory: 250Mi
          limits:
            cpu: 2000m
            memory: 1Gi
```

### When using Prometheus Operator

Use the following ServiceMonitor to start scraping metrics from the CBContainers operator:
* Make sure that your Prometheus custom resource service monitor selectors match it. 
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    control-plane: operator
  name: cbcontainers-operator-metrics-monitor
  namespace: cbcontainers-dataplane
spec:
  endpoints:
  - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    path: /metrics
    port: https
    scheme: https
    tlsConfig:
      insecureSkipVerify: true
  selector:
    matchLabels:
      control-plane: operator
```

## Using HTTP proxy

Configuring the Carbon Black Cloud Container services to use HTTP proxy can be done by enabling the centralized proxy settings or by setting HTTP_PROXY, HTTPS_PROXY and NO_PROXY environment variables manually.
The centralized proxy settings apply an HTTP proxy configuration for all components, while the manual setting of environment variables allows this to be done on a per component basis.
If both HTTP proxy environment variables and centralized proxy settings are provided, the environment variables would take precedence.
The operator does not make use of the centralized proxy settings, so you have to use the environment variables for it instead.

### Configure centralized proxy settings

In order to configure the proxy environment variables in the Operator, use the following command to patch the Operator deployment:
```sh
kubectl set env -n cbcontainers-dataplane deployment cbcontainers-operator HTTP_PROXY="<proxy-url>" HTTPS_PROXY="<proxy-url>" NO_PROXY="<kubernetes-api-server-ip>/<range>"
```

Update the `CBContainersAgent` CR with the centralized proxy settings (`kubectl edit cbcontainersagents.operator.containers.carbonblack.io cbcontainers-agent`):

```yaml
spec:
  components:
    settings:
      proxy:
        enabled: true
        httpProxy: "<proxy-url>"
        httpsProxy: "<proxy-url>"
        noProxy: "<exclusion1>,<exclusion2>"
```

You can disable the centralized proxy settings without having to delete them, by setting the `enabled` key above to `false`.

By default, the centralized proxy settings take care of determining the API server IP address(es) and the necessary proxy exclusions for the cbcontainers-dataplane namespace.
These determined values are automatically appended to the `noProxy` values from above or the specified `NO_PROXY` environment variable for a particular component.
However, if you wish to change those pre-determined values, you can specify the `noProxySuffix` key at the same level as the `noProxy` key.
It has the same format as the `noProxy` key and its values are treated in the same way as if they were pre-determined.
One can also force nothing to be appended to `noProxy` or `NO_PROXY` by setting `noProxySuffix` to an empty string.

### Configure HTTP proxy environment variables (per component proxy settings)

In order to configure those environment variables for the basic, Runtime and Image Scanning  components,
update the `CBContainersAgent` CR using the proxy environment variables (`kubectl edit cbcontainersagents.operator.containers.carbonblack.io cbcontainers-agent`):

```yaml
spec:
  components:
    basic:
      enforcer:
        env:
          HTTP_PROXY: "<proxy-url>"
          HTTPS_PROXY: "<proxy-url>"
          NO_PROXY: "<kubernetes-api-server-ip>/<range>"
      stateReporter:
        env:
          HTTP_PROXY: "<proxy-url>"
          HTTPS_PROXY: "<proxy-url>"
          NO_PROXY: "<kubernetes-api-server-ip>/<range>"
    runtimeProtection:
      resolver:
        env:
          HTTP_PROXY: "<proxy-url>"
          HTTPS_PROXY: "<proxy-url>"
          NO_PROXY: "<kubernetes-api-server-ip>/<range>"
      sensor:
        env:
          HTTP_PROXY: "<proxy-url>"
          HTTPS_PROXY: "<proxy-url>"
          NO_PROXY: "<kubernetes-api-server-ip>/<range>,cbcontainers-runtime-resolver.cbcontainers-dataplane.svc.cluster.local"
    clusterScanning:
      clusterScanner:
        env:
          HTTP_PROXY: "<proxy-url>"
          HTTPS_PROXY: "<proxy-url>"
          NO_PROXY: "<kubernetes-api-server-ip>/<range>,cbcontainers-image-scanning-reporter.cbcontainers-dataplane.svc.cluster.local"
      imageScanningReporter:
        env:
          HTTP_PROXY: "<proxy-url>"
          HTTPS_PROXY: "<proxy-url>"
          NO_PROXY: "<kubernetes-api-server-ip>/<range>"
```

It is very important to configure the NO_PROXY environment variable with the value of the Kubernetes API server IP.

Finding the API-server IP:
```sh
kubectl -n default get service kubernetes -o=jsonpath='{..clusterIP}'
```

### Other proxy considerations

When using non-transparent HTTPS proxy you will need to configure the agent to use the proxy certificate authority:
```yaml
spec:
  gateways:
    gatewayTLS:
      rootCAsBundle: <Base64 encoded proxy CA>
```
Another option will be to allow the agent communicate without verifying the certificate. this option is not recommended and exposes the agent to MITM attack.
```yaml
spec:
  gateways:
    gatewayTLS:
      insecureSkipVerify: true
```

## Changing the source of the images

By default, all the images for the operator and agent deployment are going to be pulled from Docker Hub.

We understand that some companies might not want to pull images from Docker Hub and would prefer to mirror them into their internal repositories.

For that reason, we allow specifying the image yourself.
To do that modify the `CBContainersAgent` resource you're applying to your cluster.

Modify the following properties to specify the image for each service:

- monitor - `spec.components.basic.monitor.image`
- enforcer - `spec.components.basic.enforcer.image`
- state-reporter - `spec.components.basic.stateReporter.image`
- runtime-resolver - `spec.components.runtimeProtection.resolver.image`
- runtime-sensor - `spec.components.runtimeProtection.sensor.image`
- image-scanning-reporter - `spec.components.clusterScanning.imageScanningReporter.image`
- cluster-scanner - `spec.components.clusterScanning.clusterScanner.image`

The `image` object consists of 4 properties:

- `repository` - the repository of the image, e.g. `docker.io/my-org/monitor`
- `tag` - the version tag of the image, e.g. `1.0.0`, `latest`, etc.
- `pullPolicy` - the pull policy for that image, e.g. `IfNotPresent`, `Always`, or `Never`.
  See [docs](https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy).
- `pullSecrets` - the image pull secrets that are going to be used to pull the container images.
  The secrets must already exist in the cluster.
  See [docs](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).

A sample configuration can look like this:

```yaml
spec:
  monitor:
    image:
      repository: docker.io/my-org/monitor
      tag: 1.0.0
      pullPolicy: Always
      pullSecrets:
        - my-pull-secret
```

This means that the operator will try to run the monitor service from the `docker.io/my-org/monitor:1.0.0` container image and the kubelet will be instruted to **always** pull the image, using the `my-pull-secret` secret.

### Using a shared secret for all images

If you want to use just one pull secret to pull all the custom images, you don't need to add it every single image configuration.
Instead you can specify it(them) under `spec.settings.imagePullSecrets`.

The secrets you put on that list will be added to the `imagePullSecrets` list of ALL agent workloads.

## Utilizing v1beta1 CustomResourceDefinition versions
The operator supports Kubernetes clusters from v1.13+. 
The CustomResourceDefinition APIs were in beta stage in those cluster and were later promoted to GA in v1.16. They are no longer served as of v1.22 of Kubernetes.

To maintain compatibility, this operator offers 2 sets of CustomResourceDefinitions - one under the `apiextensions/v1beta1` API and one under `apiextensons/v1`.

By default, all operations in the repository like `deploy` or `install` work with the v1 version of the `apiextensions` API. Utilizing `v1beta1` is supported by passing the `CRD_VERSION=v1beta1` option when running make.
Note that both `apiextensions/v1` and `apiextensions/v1beta1` versions of the CRDs are generated and maintained by `make` - only commands that use the final output work with 1 version at a time.

For example, this command will deploy the operator resources on the current cluster but utilizing the `apiextensions/v1beta1` API version for them.

```
make deploy CRD_VERSION=v1beta1
```
## Deploying on Openshift

The operator and its agent require elevated permissions to operate properly. However, this violates the default SecurityContextConstraints on most Openshift clusters, hence the components fail to start.
This can be fixed by applying the following custom security constraint configurations on the cluster (cluster admin priveleges required).

```yaml
kind: SecurityContextConstraints
apiVersion: security.openshift.io/v1
metadata:
  name: scc-anyuid
runAsUser:
  type: MustRunAsNonRoot
allowHostPID: false
allowHostPorts: false
allowHostNetwork: false
allowHostDirVolumePlugin: false
allowHostIPC: false
allowPrivilegedContainer: false
readOnlyRootFilesystem: true
seLinuxContext:
  type: RunAsAny
fsGroup:
  type: RunAsAny
supplementalGroups:
  type: RunAsAny
users:
- system:serviceaccount:cbcontainers-dataplane:cbcontainers-operator
- system:serviceaccount:cbcontainers-dataplane:cbcontainers-enforcer
- system:serviceaccount:cbcontainers-dataplane:cbcontainers-state-reporter
- system:serviceaccount:cbcontainers-dataplane:cbcontainers-monitor
- system:serviceaccount:cbcontainers-dataplane:cbcontainers-runtime-resolver
---
kind: SecurityContextConstraints
apiVersion: security.openshift.io/v1
metadata:
  name: scc-image-scanning # This probably needs to be fixed in the actual deployment
runAsUser:
  type: RunAsAny
allowHostPID: false
allowHostPorts: false
allowHostNetwork: false
allowHostDirVolumePlugin: false
allowHostIPC: false
allowPrivilegedContainer: false
readOnlyRootFilesystem: false
seLinuxContext:
  type: RunAsAny
fsGroup:
  type: RunAsAny
supplementalGroups:
  type: RunAsAny
allowedCapabilities:
- 'NET_BIND_SERVICE'
users:
- system:serviceaccount:cbcontainers-dataplane:cbcontainers-image-scanning
---
kind: SecurityContextConstraints
apiVersion: security.openshift.io/v1
metadata:
  name: scc-node-agent
runAsUser:
  type: RunAsAny
allowHostPID: true
allowHostPorts: false
allowHostNetwork: true
allowHostDirVolumePlugin: true
allowHostIPC: false
allowPrivilegedContainer: true
readOnlyRootFilesystem: false
seLinuxContext:
  type: RunAsAny
fsGroup:
  type: RunAsAny
supplementalGroups:
  type: RunAsAny
volumes:
- configMap
- downwardAPI
- emptyDir
- hostPath
- persistentVolumeClaim
- projected
- secret
users:
- system:serviceaccount:cbcontainers-dataplane:cbcontainers-agent-node
```
