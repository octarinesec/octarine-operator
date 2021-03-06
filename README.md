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
Kubernetes 1.13+ 

### Create the operator image
```
make docker-build docker-push IMG={IMAGE_NAME}
```
* or use the latest official image: ```cbartifactory/octarine-operator:3.0.1```

### Deploy the operator resources
```
make deploy IMG={IMAGE_NAME}
```

* View [Developer Guide](docs/developers.md#deploying-the-operator-without-using-an-image) to see how deploy the operator without using an image

## Data Plane Deployment

### 1. Apply the Carbon Black Container Api Token Secret

```
kubectl create secret generic cbcontainers-access-token \
--namespace cbcontainers-dataplane --from-literal=accessToken=\
{API_Secret_Key}/{API_ID}
```

### 2. Apply the Carbon Black Container Custom Resources

The operator implements controllers for the Carbon Black Container custom resources definitions

[Full Custom Resources Definitions Documentation](docs/crds.md)

#### 2.1 Apply the Carbon Black Container Cluster CR
<u>cbcontainersclusters.operator.containers.carbonblack.io</u>

This is the first CR you'll need to deploy in order to initialize the data plane components.
This will create the data plane Configmap, Registry Secret and PriorityClass.

```sh
apiVersion: operator.containers.carbonblack.io/v1
kind: CBContainersCluster
metadata:
  name: cbcontainerscluster-sample
spec:
  account: {ORG_KEY}
  apiGatewaySpec:
    host: {API_HOST}
  clusterName: {CLUSTER_GROUP}:{CLUSTER_NAME}
  eventsGatewaySpec:
    host: {EVENTS_HOST}
```

* notice that without applying the api token secret, the operator will return the error:
`couldn't find access token secret k8s object`

#### 2.2 Apply the Carbon Black Container Hardening CR
<u>cbcontainershardenings.operator.containers.carbonblack.io</u>

This is the CR you'll need to deploy in order to install the Carbon Black Container Hardening feature components.
This will install the Hardening Enforcer components that are responsible for enforcing the configured policies and
the State Reporter components that are responsible for reporting the cluster state.

* Notice that without the first CR, the Hardening components won't be able to work. 

```yaml
apiVersion: operator.containers.carbonblack.io/v1
kind: CBContainersHardening
metadata:
  name: cbcontainershardening-sample
spec:
  version: {HARDENING_VERSION}
  eventsGatewaySpec:
    host: {EVENTS_HOST}
```

#### 2.3 Apply the Carbon Black Container Runtime CR
<u>cbcontainersruntimes.operator.containers.carbonblack.io</u>

TODO

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

In order to configure those environment variables for the Hardening Enforcer and the Hardening State Reporter components,
update the Hardening CR using the proxy environment variables:

```yaml
spec:
  enforcerSpec:
    env:
      HTTP_PROXY: "<proxy-url>"
      HTTPS_PROXY: "<proxy-url>"
      NO_PROXY: "<kubernetes-api-server-ip>/<range>"
  stateReporterSpec:
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
