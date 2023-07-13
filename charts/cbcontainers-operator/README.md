# cbcontainers-operator

This is the official Helm chart for installation of the CBContainers Operator.

Helm 3 is supported.

## Installation

The chart can be installed as is, without any customization or modifications.

### Choosing a namespace for the Helm release

You can create the Helm release in any namespace that you want.

You can also customize the namespace in which the operator itself is installed.
See [Customization](#namespace).

### Installing the operator chart

Now, install the actual helm chart from source:

```sh
cd charts/cbcontainers-operator
git checkout v5.6.2 # install the latest version of the operator
helm install cbcontainers-operator ./cbcontainers-operator-chart
```

## Customization

| Parameter                        | Description                                         | Default                                                                            |
|----------------------------------|-----------------------------------------------------|------------------------------------------------------------------------------------|
| `spec.operator.image.repository` | The repository of the operator image                | `cbartifactory/octarine-operator`                                                  |
| `spec.operator.image.version`    | The version of the operator image                   | The latest version of the operator image                                           |
| `spec.operator.resources`        | Carbon Black Container Operator resources           | `{requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "256Mi", cpu: "200m"}}` |
| `spec.rbacProxy.resources`       | Kube RBAC Proxy resources                           | `{requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "256Mi", cpu: "200m"}}` |
| `spec.operator.environment`      | Environment variables to be set to the operator pod | []                                                                                 |

### Namespace

By default, the CBContainers Operator is installed in the `cbcontainers-dataplane` namespace.

If you want to change that, set the `operatorNamespace` field in your `values.yaml` file.

The chart will automatically create the namespace that you have chosen to install the operator into.
If you don't want to do that (because you have already created the namespace), set the `createOperatorNamespace` field in your `values.yaml` file to `false`.

If the namespace is pre-created, then it must also be labeled properly or the operator and agent might not reconcile successfully. 
The commands below show an example of creating a custom namespace, labeling and installing the operator inside.

```sh
NAMESPACE=<choose_your_value>
kubectl create namespace $NAMESPACE
kubectl label namespace $NAMESPACE control-plane=operator octarine=ignore
helm install cbcontainers-operator ./cbcontainers-operator-chart --set createOperatorNamespace=false,operatorNamespace=$NAMESPACE
```

### CRD Installation

By default, installing the chart will also create the `CBContainersAgent` CRD.

If you want to manage the CRD in a different way and don't want to install it along with the chart, set the `installCRD` field in your `values.yaml` file to `false`.

### HTTP Proxy

If you want to use an HTTP proxy for the communication with the CBC backend you need to set 3 environment variables.
These are exposed via the `Values.operator.proxy` parameters in the `values.yaml` file:

- `Values.operator.proxy.http`
- `Values.operator.proxy.https`
- `Values.operator.proxy.noProxy`

If you want to use HTTP proxy you need to set ALL 3 values.
For more info see <https://github.com/octarinesec/octarine-operator/tree/master#using-http-proxy>.

## Templates

This chart consists of four [templates](cbcontainers-operator-chart/templates).

The [operator.yaml](cbcontainers-operator-chart/templates/operator.yaml) file contains all resources, apart from the operator deployment.
It is generated via `kustomize`.
For more info see [config/default_chart](../../config/default_chart).

The [deployment.yaml](cbcontainers-operator-chart/templates/deployment.yaml) file contains the operator Deployment resource.
It is derived from [this Kustomize configuration](../../config/manager) but because it needs to be configurable via Helm it is heavily templated.
Because of that it cannot be generated automatically, so it should be maintained by hand.
If any changes are make to the [Kustomize configuration](../../config/manager), they should also be reflected in that file.

The [dataplane_rbac.yaml](cbcontainers-operator-chart/templates/dataplane_rbac.yaml) and [dataplane_service_accounts](cbcontainers-operator-chart/templates/dataplane_service_accounts.yaml) 
files contain necessary RBAC objects for the agent to work as expected. 
