# cbcontainers-operator

This is the official Helm chart for installation of the CBContainers Operator.

Helm 3 is supported.

## Installation

The chart can be installed as is, without any customization or modifications.

### Choosing a namespace for the Helm release

Currently, the charts does not support running outside the `cbcontainers-dataplane` namespace, and they create and label that namespace as needed.
Therefore, there are two options for choosing the namespace for the actual helm release - creating the `cbcontainers-dataplane` namespace and "adopting" it via Helm _or_ installing the release in another namespace.

**Option 1**: Using `cbcontainers-dataplane` to manage the release (recommended)

Run the following commands to prepare the release namespace:

```sh
kubectl create namespace cbcontainers-dataplane
kubectl annotate namespace cbcontainers-dataplane meta.helm.sh/release-name=cbcontainers-operator meta.helm.sh/release-namespace=cbcontainers-dataplane
kubectl label namespace cbcontainers-dataplane app.kubernetes.io/managed-by=Helm
```

And use `cbcontainers-dataplane` in all commands below that require a namespace (`--namespace X`).
With this option, future commands like `helm install`, `helm list` should be run in the context of the `cbcontainers-dataplane` namespace.

**Option 2**: Using a different namespace to manage the release

Choose a namespace that exists in the cluster - `my-namespace` and use that for all commands that require a namespace below (`--namespace X`).
Note that in this case the resources are still installed in `cbcontainers-dataplane` namespace, but the actual Helm release does not live there.
With this option, future commands like `helm install`, `helm list`, etc. must be run in the context of the chosen namespace `my-namespace`.

### Installing the operator chart

Now, install the actual helm chart in the namespace based on the chosen option 1 or 2 above.

```sh
helm repo add vmware TODO-chart-repo/TODO-chart-name
helm repo update
helm install cbcontainers-operator TODO-chart-repo/TODO-chart-name --namespace X
```

or from source:

```sh
cd charts/cbcontainers-operator
helm install cbcontainers-operator ./cbcontainers-operator-chart --namespace X
```

## Customization

| Parameter                        | Description                                         | Default                                                                            |
|----------------------------------|-----------------------------------------------------|------------------------------------------------------------------------------------|
| `spec.operator.image.repository` | The repository of the operator image                | `cbartifactory/octarine-operator`                                                  |
| `spec.operator.image.version`    | The version of the operator image                   | The latest version of the operator image                                           |
| `spec.operator.resources`        | Carbon Black Container Operator resources           | `{requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "256Mi", cpu: "200m"}}` |
| `spec.rbacProxy.resources`       | Kube RBAC Proxy resources                           | `{requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "256Mi", cpu: "200m"}}` |
| `spec.operator.environment`      | Environment variables to be set to the operator pod | []                                                                                 |

### HTTP Proxy

If you want to use an HTTP proxy for the communication with the CBC backend you need to set 3 environment variables.
These are exposed via the `spec.operator.proxy` parameters in the `values.yaml` file:

- `spec.operator.proxy.http`
- `spec.operator.proxy.https`
- `spec.rbacProxy.proxy.noProxy`

If you want to use HTTP proxy you need to set ALL 3 values.
For more info see <https://github.com/octarinesec/octarine-operator/tree/master#using-http-proxy>.

## Templates

This chart consists of two [templates](cbcontainers-operator-chart/templates).

The [operator.yaml](cbcontainers-operator-chart/templates/operator.yaml) file contains all resources, apart from the operator deployment.
It is generated via `kustomize`.
For more info see [config/default_chart](../../config/default_chart).

The [deployment.yaml](cbcontainers-operator-chart/templates/deployment.yaml) file contains the operator Deployment resource.
It is derived from [this Kustomize configuration](../../config/manager) but because it needs to be configurable via Helm it is heavily templated.
Because of that it cannot be generated automatically, so it should be maintained by hand.
If any changes are make to the [Kustomize configuration](../../config/manager), they should also be reflected in that file.
