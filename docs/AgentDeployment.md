## Agent Deployment

### 1. Apply the Carbon Black Container Api Token Secret

```
kubectl create secret generic cbcontainers-access-token \
--namespace cbcontainers-dataplane --from-literal=accessToken=\
{API_Secret_Key}/{API_ID}
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
