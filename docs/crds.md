# VMware Carbon Black Cloud Container Operator

## Cloud Container Operator Custom Resources Definitions

The operator implements controllers for the Carbon Black Container custom resources definitions:

### 1. Carbon Black Container Cluster CR

<u>cbcontainersclusters.operator.containers.carbonblack.io</u>

This is the first CR you'll need to deploy in order to initialize the data plane components. This will create the data
plane Configmap, Registry Secret and PriorityClass.

### Required parameters

| Parameter                                    | Description                                                     | Default |
| ---------------------------------------------| --------------------------------------------------------------- | ------- |
| `spec.account`                               | Carbon Black Container org key                                  | ------- |
| `spec.clusterName`                           | Carbon Black Container cluster name  (<group:name>)             | ------- |
| `spec.apiGatewaySpec.host`                   | Carbon Black Container api host                                 | ------- |
| `spec.eventsGatewaySpec.host`                | Carbon Black Container events host                              | ------- |

### Optional parameters

| Parameter                                    | Description                                              | Default                     |
| ---------------------------------------------| ---------------------------------------------------------| ------------                |
| `spec.apiGatewaySpec.scheme`                 | Carbon Black Container api scheme                        | `https`                     |
| `spec.apiGatewaySpec.port`                   | Carbon Black Container api port                          | 443                         |
| `spec.apiGatewaySpec.adapter`                | Carbon Black Container api adapter                       | `containers`                |
| `spec.apiGatewaySpec.accessTokenSecretName`  | Carbon Black Container api access token secret name      | `cbcontainers-access-token` |
| `spec.eventsGatewaySpec.port`                | Carbon Black Container events port                       | 443                         |

### 2. Carbon Black Container Hardening CR

### Required parameters

| Parameter                                    | Description                                                     | Default |
| ---------------------------------------------| --------------------------------------------------------------- | ------- |
| `global.octarine.version`                    | Version of the Carbon Black Cloud components                    | ------- |
| `spec.account`                               | Carbon Black Container org key                                  | ------- |
| `spec.clusterName`                           | Carbon Black Container cluster name  (<group:name>)             | ------- |
| `spec.apiGatewaySpec.host`                   | Carbon Black Container api host                                 | ------- |
| `spec.eventsGatewaySpec.host`                | Carbon Black Container events host                              | ------- |

### Optional parameters

| Parameter                                    | Description                                              | Default                     |
| ---------------------------------------------| ---------------------------------------------------------| ------------                |
| `spec.apiGatewaySpec.scheme`                 | Carbon Black Container api scheme                        | `https`                     |
| `spec.apiGatewaySpec.port`                   | Carbon Black Container api port                          | 443                         |
| `spec.apiGatewaySpec.adapter`                | Carbon Black Container api adapter                       | `containers`                |
| `spec.apiGatewaySpec.accessTokenSecretName`  | Carbon Black Container api access token secret name      | `cbcontainers-access-token` |
| `spec.eventsGatewaySpec.port`                | Carbon Black Container events port                       | 443                         |