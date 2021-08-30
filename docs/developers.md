# VMware Carbon Black Cloud Container Operator

## Developer Guide

### Deploying the operator without using an image

#### Installing dependencies (verify the kube config context)
```
make deploy OPERATOR_REPLICAS=0
```

#### Running the operator from the terminal
* Run the following commands (verify the kube config context)
```
make run
```
The `run` command executes with the local GO environment the main.go file

#### From your editor
* Run/Debug the main.go from your editor (verify the `KUBECONFIG` env var)


### Installing the Data Plane on your own control plane

Under the Carbon Black Container Cluster CR:
```
spec:
  apiGatewaySpec:
    adapter: {MY-ADAPTER-NAME}
```

Change {MY-ADAPTER-NAME} to your control plane adapter name.
The default value is `containers`

### Changing the security context settings

#### Hardening enforcer/state_reporter security context settings
Under `cbcontainers/state/hardening/objects`
for `enforcer_deployment.go` or `state_reporter_deployment.go`
You can change the values on the top of the file to suite your needs.

### Using defaults 
One of the new features in `apiextensions/v1` version of CustomResourceDefinitions is defaults in the OpenAPISchema. These are supported by kubebuilder via tags - e.g. `kubebuilder:default=something`
For backwards compatibility reasons, all defaults should also be implemented and set in the controllers to ensure they work on clusters v1.15 and below.

One issue with defaults is that kubebuilder does not support empty object as default value - see
[related issue](https://github.com/kubernetes-sigs/controller-tools/issues/550). The issue is about maps but the same code causes problems with objects.

What this means is that the following spec will _not_ apply the default for foo unless the user specifies bar. 

```yaml
spec:
  properties:
    bar:
      properties:
        foo:
          default: 10
          type: integer
```
So applying this YAML will lead to empty object for `bar` being saved:
```yaml
spec: {}
```

Instead, applying 
```yaml
spec: { bar: {} }
``` 
would work as expected and save the following object:
```yaml
spec: { bar: { foo: 10 }}
```

This would work as expected when adding an empty object default to the example as below:
```yaml
spec:
  properties:
    bar:
      default: {}
      properties:
        foo:
          default: 10
          type: integer
```
Unfortunately kubebuilder cannot produce that output today.
Therefore, a special `make` target works around this by replacing all instance of `<>` with `{}` so using `kubebuilder:default=<>` will produce the correct output.

Defaulting is not supported by `v1beta1` versions of CRDs so warnings are expected when generating those since kubebuilder.

## Debugging locally

To debug locally, run `make run-delve` which will build and start a delve debugger in headless mode.
Then use your editor to start a remote session and connect to the delve instance.

For goland, the built-in `go remote` configuration works fine.