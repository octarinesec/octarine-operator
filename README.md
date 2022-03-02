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

## Operator Deployment

### Prerequisites
Kubernetes 1.13+ is supported.

### From script:
```
export OPERATOR_VERSION=v4.0.0
export OPERATOR_SCRIPT_URL=https://setup.containers.carbonblack.io/operator-$OPERATOR_VERSION-apply.sh
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

Configuring the Carbon Black Cloud Container services to use HTTP proxy can be done by setting HTTP_PROXY, HTTPS_PROXY and NO_PROXY environment variables.

In order to configure those environment variables in the Operator, use the following command to patch the Operator deployment:
```sh
kubectl set env -n cbcontainers-dataplane deployment cbcontainers-operator HTTP_PROXY="<proxy-url>" HTTPS_PROXY="<proxy-url>" NO_PROXY="<kubernetes-api-server-ip>/<range>"
```

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
          NO_PROXY: "<kubernetes-api-server-ip>/<range>"
    clusterScanning:
      clusterScanner:
        env:
          HTTP_PROXY: "<proxy-url>"
          HTTPS_PROXY: "<proxy-url>"
          NO_PROXY: "<kubernetes-api-server-ip>/<range>"
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

When using non transparent HTTPS proxy you will need to configure the agent to use the proxy certificate authority:
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

## Utilizing v1beta1 CustomResourceDefinition versions
The operator supports Kubernetes clusters from v1.13+. 
The CustomResourceDefinition APIs were in beta stage in those cluster and were later promoted to GA in v1.16. They are no longer served as of v1.22 of Kubernetes.

To maintain compatibility, this operator offers 2 sets of CustomResoruceDefinitions - one under the `apiextensions/v1beta1` API and one under `apiextensons/v1`.

By default, all operations in the repository like `deploy` or `install` work with the v1 version of the `apiextensions` API. Utilizing `v1beta1` is supported by passing the `CRD_VERSION=v1beta1` option when running make.
Note that both `apiextensions/v1` and `apiextensions/v1beta1` versions of the CRDs are generated and maintained by `make` - only commands that use the final output work with 1 version at a time.

For example, this command will deploy the operator resources on the current cluster but utilizing the `apiextensions/v1beta1` API version for them.

```
make deploy CRD_VERSION=v1beta1
```
