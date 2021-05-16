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


