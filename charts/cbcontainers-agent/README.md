# cbcontainers-agent

This is the official Helm chart for installation of the CBContainers agent.

Helm 3 is supported.

## Installation

In order for the chart to be installed it needs minimal configuration.

There are 8 required fields that need to be provided by the user:

| Parameter                                              | Description                                                      |
| ------------------------------------------------------ | ---------------------------------------------------------------- |
| `spec.orgKey`                                       | The org key of the organization using CBC                     |
| `spec.clusterName`                                     | The name of the cluster that will be added to CBC                |
| `spec.clusterGroup`                                    | The group that the cluster belongs to in CBC                     |
| `spec.version`                                         | The version of the agent images                                  |
| `spec.gateways.apiGatewayHost`                         | The URL of the CBC API Gateway                                   |
| `spec.gateways.coreEventsGatewayHost`                  | The URL of the CBC Core events Gateway                           |
| `spec.gateways.hardeningEventsGatewayHost`             | The URL of the CBC Hardening events Gateway                      |
| `spec.gateways.runtimeEventsGatewayHost`               | The URL of the CBC Runtime events Gateway                        |

After setting these required fields in a `values.yaml` file you can install the chart from our repo:

```sh
helm repo add vmware TODO-chart-repo/TODO-chart-name -f values.yaml
helm repo update
helm install cbcontainers-agent TODO-chart-repo/TODO-chart-name -f values.yaml
```

or from source

```sh
cd charts/cbcontainers-agent
helm install cbcontainers-agent ./cbcontainers-agent-chart -f values.yaml
```

## Customization

The way that the CBC Containers components are installed is highly customizable.

You can set different properties for the components or enable/disable components via the `spec.components` section of your `values.yaml` file.

For all the possible values see <https://github.com/octarinesec/octarine-operator/blob/master/docs%2Fcrds.md#basic-components-optional-parameters> and [`example-value.yaml`](cbcontainers-agent-chart/example-values.yaml).

### Secret creation

In order for the agent components to function correctly and be able to communicate with the CBC backend an access token is required.

This token is located in a secret.
By default the secret is named `"cbcontainers-access-token"`, but that is configurable via the `accessTokenSecretName` property.

If that secret does not exist, the operator will not start any of the agent components.

If you want to create the secret as part of the chart installation provide the `accessToken` value to the chart.

This means storing the secret as plain text in your `values.yaml` file.

If you prefer to create the `Secret` yourself in an alternative and more secure way, don't set the `accessToken` value and the chart will not create the `Secret` objects.
