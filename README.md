# Octarine Operator Helm Charts
This repo contains helm chart for installing Octarine components using Octarine operator.

## About
The *octarine-operator* runs within a Kubernetes cluster. It's a set of controllers which deploy and manage Octarine components.  
Some of the operator capabilities:
* Deploy & manage *Guardrails*
* Deploy & manage *Nodeguard*
* Automatically fetch & deploy Octarine private image registry secret
* Automatically register Octarine domain
* Manage the *Guardrails* validating webhook based on the service availability - deploy it when the service is available, delete when it isn't.
* Monitor the Octarine components and send health report to Octarine backend.

Octarine operator utilizes the operator-framework to create a hybrid operator, which combines helm and go operators.  
The helm controller within the operator is responsible for managing the Octarine components deployment, and the go controller monitors & manages them. 

## Prerequisites
- Kubernetes 1.3+
- Helm installed, Tiller pod is running

## Deployment
Install the chart with the release name `octarine` in the `octarine-dataplane` namespace:
```shell script
helm upgrade --install --namespace octarine-dataplane octarine ./helm-charts/octarine-operator/
```
After deploying the `octarine-operator`, please refer to the Octarine [Custom Resource documentation](docs/octarine_cr.md) in order to deploy Octarine dataplane components.

## Rolling upgrade
Upgrade the `octarine` release to the desired version:
```shell script
helm upgrade octarine ./helm-charts/octarine-operator/ --reuse-values
```

## Uninstalling the Octarine operator
> If you created an Octarine resource to install the Octarine components, please delete it before uninstalling the operator.

1. Uninstall the `octarine` release:
```shell script
helm delete octarine
```
2. Delete the Octarine CRD which was created by helm:
```shell script
kubectl delete octarines.operator.octarinesec.com octarine
```

## Customize the configuration
The following table lists the configurable parameters of the octarine operator chart and their default values.

Parameter | Description | Default
--------- | ----------- | -------
`replicaCount` | The number of the operator replicas to run | `1`

## Logs
You can enable verbose logging and set the verbosity level using the `--zap-level` flag of the operator executable.  
See the `args` value within the `values.yaml`. 
