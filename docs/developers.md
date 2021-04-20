# VMware Carbon Black Cloud Container Operator

## Developer Guide

### Running the operator locally without deploying it

#### From the terminal
* Run the following command (verify the kube config context)
```
make install run
```

#### From your editor
* Run the following command (verify the kube config context)
```
make install
```
* Run/Debug the main.go from your editor (verify the `KUBECONFIG` env var)


The `install` command deploys the custom resource definitions and the rbac resources.

The `run` command executes with the local GO environment the main.go file

