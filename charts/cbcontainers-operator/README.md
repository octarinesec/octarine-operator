# cbcontainers-operator

This is the official Helm chart for installation of the CBContainers Operator.

## Installation

The chart can be installed as is, without any customization or modifications.

```sh
helm repo add vmware TODO-chart-repo/TODO-chart-name
helm repo update
helm install cbcontainers-operator TODO-chart-repo/TODO-chart-name
```

or from source

```sh
cd charts/cbcontainers-operator
helm install cbcontainers-operator ./cbcontainers-operator-chart
```

## Customization

| Parameter                                              | Description                                                      | Default                                                                            |
| ------------------------------------------------------ | ---------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| `spec.operator.image.repository`                       | The repository of the operator image                             | `cbartifactory/octarine-operator`                                                  |
| `spec.operator.image.version`                          | The version of the operator image                                | `5.1.0`                                                                            |
| `spec.rbacProxy.image.repository`                      | The repository of the Kube RBAC proxy image                      | `gcr.io/kubebuilder/kube-rbac-proxy`                                               |
| `spec.rbacProxy.image.version`                         | The version of the Kube RBAC proxy image                         | `v0.8.0`                                                                           |
| `spec.operator.resources`                              | Carbon Black Container Operator resources                        | `{requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "256Mi", cpu: "200m"}}` |
| `spec.rbacProxy.resources`                             | Kube RBAC Proxy resources                                        | `{requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "256Mi", cpu: "200m"}}` |
| `spec.operator.environment`                            | Environment variables to be set to the operator pod              | []                                                                                 |

### HTTP Proxy

If you want to use an HTTP proxy for the communication with the CBC backend you need to set 3 environment variables.
These are exposed via the `spec.operator.proxy` parameters in the `values.yaml` file:

- `spec.operator.proxy.http`
- `spec.operator.proxy.https`
- `spec.rbacProxy.proxy.noProxy`

If you want to use HTTP proxy you need to set ALL 3 values.
For more info see <https://github.com/octarinesec/octarine-operator/tree/master#using-http-proxy>.

## Templates

This chart consists of two [templates](cbcontainers-operator-chart/templates/).

The [operator.yaml](operator.yaml) file contains all resource, apart from the operator deployment.
It is generated with via `kustomize`.
For more info see [config/default_chart](../../../../config/default_chart/).

The [deployment.yaml](deployment.yaml) file contains the operator Deployment resource.
It is derived from [this Kustomize configuration](../../../../config/manager) but because it needs to be configurable via Helm it is heavily templated.
Because of that it cannot be generated automatically, so it should be maintained by hand.
If any changes are make to the [Kustomize configuration](../../../../config/manager), they should also be reflected in that file.
