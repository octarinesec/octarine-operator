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

| Operator version | Kubernetes Sensor Component Version  | Minimum Kubernetes Version |
|------------------|--------------------------------------|----------------------------|
| v6.1.x           | 2.10.0, 2.11.0, 2.12.0, 3.0.X, 3.1.X | 1.18                       |
| v6.0.x           | 2.10.0, 2.11.0, 2.12.0, 3.0.X, 3.1.X | 1.18                       |
| v5.6.x           | 2.10.0, 2.11.0, 2.12.0               | 1.16                       |
| v5.5.x           | 2.10.0, 2.11.0                       | 1.16                       |

## Operator Deployment

### Prerequisites
Kubernetes 1.18+ is supported.

### From script:
```
export OPERATOR_VERSION=v6.1.0
export OPERATOR_SCRIPT_URL=https://setup.containers.carbonblack.io/$OPERATOR_VERSION/operator-apply.sh
curl -s $OPERATOR_SCRIPT_URL | bash
```

{OPERATOR_VERSION} is of the format "v{VERSION}"

Versions list: [Releases](https://github.com/octarinesec/octarine-operator/releases)

### OpenShift Deployment:
For OpenShift clusters, follow the OpenShift Deployment instructions:

[OpenShift Deployment and Uninstall](docs/OpenshiftDeployment.md)


* For deploying from the source code, follow the instructions in the [Operator Deployment](docs/OperatorDeployment.md) documentation
* View [Developer Guide](docs/developers.md) to see how deploy the operator without using an image

## Data Plane Deployment

### 1. Apply the Carbon Black Container Api Token Secret and Company Code Secret

```
kubectl create secret generic cbcontainers-access-token --namespace cbcontainers-dataplane --from-literal=accessToken={API_Secret_Key}/{API_ID}
kubectl create secret generic cbcontainers-company-code --namespace cbcontainers-dataplane --from-literal=companyCode=RXXXXXXXXXXG\!XXXX
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
export OPERATOR_VERSION=v6.1.0
export OPERATOR_SCRIPT_URL=https://setup.containers.carbonblack.io/$OPERATOR_VERSION/operator-apply.sh
curl -s $OPERATOR_SCRIPT_URL | bash -s -- -u 
```

* Notice that the above command will delete the Carbon Black Container custom resources definitions and instances.

## Helm Charts Documentation
[VMware Carbon Black Cloud Container Helm Charts Documentation](charts/README.md)

## Full Documentation
[VMware Carbon Black Cloud Container Operator Documentation](docs/Main.md)
