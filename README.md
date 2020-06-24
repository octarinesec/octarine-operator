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
- Kubernetes 1.13+

## Deployment

### Deploy using Helm
The operator Helm chart support Helm 3 and Helm 2 (if you're using Helm 2, make sure that Tiller pod is running).  
Install the chart with the release name `octarine` in the `octarine-dataplane` namespace:
```shell script
helm repo add octarine-operator https://octarinesec.github.io/octarine-operator
helm upgrade --install --namespace octarine-dataplane octarine octarine-operator/octarine-operator
```

### Deploy plain K8s resources
Create the `octarine-dataplane` namespace and deploy the resources from the `deploy` dir:
```shell script
kubectl create namespace octarine-dataplane
kubectl label namespace octarine-dataplane name=octarine-dataplane
kubectl apply -n octarine-dataplane -Rf deploy
```
> The namespace label is required due to the validating webhook configured by Guardrails, and in order to ensure the service availability.

*After deploying the `octarine-operator`, please refer to the Octarine [Custom Resource documentation](docs/octarine_cr.md) in order to deploy Octarine dataplane components.*

## Rolling upgrade
Upgrade the `octarine` release to the desired version:
```shell script
helm upgrade octarine ./helm-charts/octarine-operator/ --reuse-values
```

## Uninstalling the Octarine operator
**If you created an Octarine resource to install the Octarine components, please delete it before uninstalling the operator.**

### Uninstall a Helm release
If you deployed the operator using Helm:
1. Uninstall the `octarine` release:
```shell script
helm delete octarine
```
2. Delete the Octarine CRD which was created by helm:
```shell script
kubectl delete octarines.operator.octarinesec.com octarine
```

### Uninstall plain K8s resources
If you deployed the operator using its plain K8s resources, uninstall it by running:
```shell script
kubectl delete -n octarine-dataplane -Rf deploy
```

## Customize the configuration
The following table lists the configurable parameters of the octarine operator chart and their default values.

Parameter | Description | Default
--------- | ----------- | -------
`replicaCount` | The number of the operator replicas to run | `1`

## Logs
You can enable verbose logging and set the verbosity level using the `--zap-level` flag of the operator executable.  
See the `args` value within the `values.yaml`. 
