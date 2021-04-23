# VMware Carbon Black Cloud Container Operator
## Overview 

The Carbon Black Cloud Container Operator runs within a Kubernetes cluster. The Container Operator is a set of controllers which deploy and manage the VMware Carbon Black Cloud Container components. 
 
 Capabilities
 * Deploy and manage the Container Essentials product bundle (including the configuration and the image scanning for Kubernetes security)! 
 * Deploy and manage the Container Advanced product bundle (including the runtime for Kubernetes security) 
 * Automatically fetch and deploy the Carbon Black Cloud Container private image registry secret
 * Automatically register the Carbon Black Cloud Container cluster
 * Manage the Container Essentials validationng webhook - dynamically manage the admission control webhook to avoid possible downtime
 * Monitor and report agent availability to the Carbon Black console

The Carbon Black Cloud Container Operator utilizes the operator-framework to create a GO operator, which is responsible for managing and monitoring the Cloud Container components deployment. 

## Operator Deployment

### Prerequisites
Kubernetes 1.13+ 

### Create the operator image
```
make docker-build docker-push IMG={IMAGE_NAME}
```

### Deploy the operator resources
```
make deploy IMG=={IMAGE_NAME}
```

### Uninstalling the Carbon Black Cloud Container Operator
```sh
make undeploy
```
* Notice that the above command will delete the Carbon Black Container custom resources definitions and instances.

## Data Plane Deployment
The operator implements controllers for the Carbon Black Container custom resources definitions

[Full Custom Resources Definitions Documentation](docs/crds.md)

### 1. Carbon Black Container Cluster CR
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

### 2. Carbon Black Container Hardening CR
<u>cbcontainershardenings.operator.containers.carbonblack.io</u>

This is the CR you'll need to deploy in order to install the Carbon Black Container Hardening feature components.
This will install the Hardening Enforcer components that are responsible for enforcing the configured policies and
the State Reporter components that are responsible for reporting the cluster state.

* Notice that without the first CR, the Hardening components won't be able to work. 

```sh
apiVersion: operator.containers.carbonblack.io/v1
kind: CBContainersHardening
metadata:
  name: cbcontainershardening-sample
spec:
  version: {HARDENING_VERSION}
  eventsGatewaySpec:
    host: {EVENTS_HOST}
```


## Using HTTP proxy

Configuring the Carbon Black Cloud Container services to use HTTP proxy can be done by setting HTTP_PROXY, HTTPS_PROXY and NO_PROXY environment variables.

In order to configure those environment variables in the Operator, use the following command to patch the Operator deployment:
```sh
kubectl set env -n cbcontainers-dataplane deployment cbcontainers-operator HTTP_PROXY="<proxy-url>" HTTPS_PROXY="<proxy-url>" NO_PROXY="<kubernetes-api-server-ip>/<range>"
```

In order to configure those environment variables for the Hardening Enforcer and the Hardening State Reporter components,
update the Hardening CR using the proxy environment variables:

```sh
spec:s
  enforcerSpec:
    env:
      HTTP_PROXY="<proxy-url>"
      HTTPS_PROXY="<proxy-url>"
      NO_PROXY="<kubernetes-api-server-ip>/<range>"
  stateReporterSpec:
    env:
      HTTP_PROXY="<proxy-url>"
      HTTPS_PROXY="<proxy-url>"
      NO_PROXY="<kubernetes-api-server-ip>/<range>"
```

It is very important to configure the NO_PROXY environment variable with the value of the Kubernetes API server IP.

Finding the API-server IP:
```sh
kubectl -n default get service kubernetes -o=jsonpath='{..clusterIP}'
```