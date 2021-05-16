# VMware Carbon Black Cloud Container Operator

## Developer Guide

### Running the operator locally without deploying it

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
